package auth_sms

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/crypt_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/sms"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const SmsProtocol = "sms"

const TokenName = "sms-token"
const PhoneName = "sms-phone"
const DelayName = "sms-delay"
const CodeName = "sms-code"

const SmsDelayCacheKey = "sms-delay"
const SmsTokenCacheKey = "sms-token"

type UserWithPhone interface {
	Phone() string
}

type SmsDelay struct {
	common.CreatedAtBase
}

type SmsCacheToken struct {
	Try      int    `json:"try"`
	Code     string `json:"code"`
	Checksum string `json:"checksum"`
	SmsId    string `json:"sms_id"`
}

type SmsToken struct {
	auth.ExpireToken
	common.IDBase
}

type AuthSmsConfig struct {
	TOKEN_TTL_SECONDS int    `default:"300" validate:"gt=0"`
	SMS_DELAY_SECONDS int    `default:"30" validate:"gt=0"`
	SECRET            string `validate:"required" mask:"true"`
	MAX_TRIES         int    `default:"3" validate:"gt=1"`
	CODE_LENGTH       int    `default:"5" validate:"gte=4"`
}

type AuthSms struct {
	auth.AuthHandlerBase
	AuthSmsConfig
	Encryption auth.AuthParameterEncryption
	smsManager sms.SmsManager
}

func (a *AuthSms) Config() interface{} {
	return &a.AuthSmsConfig
}

func NewAuthSms(smsManager sms.SmsManager) *AuthSms {
	a := &AuthSms{}
	a.smsManager = smsManager
	return a
}

func (a *AuthSms) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	a.AuthHandlerBase.Init(SmsProtocol)

	err := object_config.LoadLogValidate(cfg, log, vld, a, "auth.methods.sms", configPath...)
	if err != nil {
		return log.Fatal("Failed to load configuration of auth SMS handler", err)
	}

	encryption := &auth.AuthParameterEncryptionBase{}
	err = object_config.LoadLogValidate(cfg, log, vld, encryption, "auth.methods.sms", configPath...)
	if err != nil {
		return log.Fatal("Failed to load configuration of auth SMS encryption", err)
	}
	a.Encryption = encryption

	return nil
}

const ErrorCodeSmsConfirmationRequired = "sms_confirmation_required"
const ErrorCodeSmsTokenRequired = "sms_token_required"
const ErrorCodeTokenExpired = "sms_token_expired"
const ErrorCodeInvalidToken = "sms_token_invalid"
const ErrorCodeInvalidSmsCode = "sms_code_invalid"
const ErrorCodeWaitDelay = "sms_wait_delay"
const ErrorCodeContentMismatch = "sms_content_mismatch"
const ErrorCodeInvalidPhone = "sms_invalid_phone"
const ErrorCodeTooManyTries = "sms_too_many_tries"

func (a *AuthSms) ErrorDescriptions() map[string]string {
	m := map[string]string{
		ErrorCodeSmsConfirmationRequired: "Request must be confirmed with SMS",
		ErrorCodeSmsTokenRequired:        "SMS token must be present in request",
		ErrorCodeTokenExpired:            "SMS token expired",
		ErrorCodeInvalidToken:            "Invalid SMS token",
		ErrorCodeInvalidSmsCode:          "Invalid SMS code",
		ErrorCodeWaitDelay:               "Wait before requesting new SMS code",
		ErrorCodeContentMismatch:         "Content of initial request and content of current request mismatch",
		ErrorCodeInvalidPhone:            "Failed to send SMS confirmation because of invalid phone number",
		ErrorCodeTooManyTries:            "Too many code tries",
	}
	return m
}

func (a *AuthSms) ErrorProtocolCodes() map[string]int {
	m := map[string]int{
		ErrorCodeSmsConfirmationRequired: http.StatusUnauthorized,
		ErrorCodeSmsTokenRequired:        http.StatusUnauthorized,
		ErrorCodeTokenExpired:            http.StatusUnauthorized,
		ErrorCodeInvalidToken:            http.StatusUnauthorized,
		ErrorCodeInvalidSmsCode:          http.StatusUnauthorized,
		ErrorCodeWaitDelay:               http.StatusUnauthorized,
		ErrorCodeTooManyTries:            http.StatusUnauthorized,
	}
	return m
}

func (a *AuthSms) Handle(ctx auth.AuthContext) (bool, error) {

	// setup
	c := ctx.TraceInMethod("AuthSms.Handle")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// check if user authenticated
	if ctx.AuthUser() == nil {
		c.SetMessage("unknown user")
		ctx.SetGenericErrorCode(auth.ErrorCodeUnauthorized)
		return true, err
	}
	userId := ctx.AuthUser().GetID()

	// get SMS code from request
	code := ctx.GetAuthParameter(a.Protocol(), CodeName)

	// check if code is set in request
	if code != "" {

		// code is set in request

		// extract and check token from request
		token := &SmsToken{}
		exists, err := a.Encryption.GetAuthParameter(ctx, a.Protocol(), TokenName, token)
		if !exists {
			c.SetMessage("SMS token not found")
			ctx.SetGenericErrorCode(ErrorCodeSmsTokenRequired)
			return false, err
		}
		if err != nil {
			c.SetMessage("failed to get encrypted SMS token")
			ctx.SetGenericErrorCode(ErrorCodeInvalidToken)
			return true, err
		}
		if token.Expired() {
			c.SetMessage("token expired")
			ctx.SetGenericErrorCode(ErrorCodeTokenExpired)
			return true, err
		}

		// find corresponding cache token
		cacheToken := &SmsCacheToken{}
		oldCacheKey := a.smsTokenCacheKey(token.GetID())
		found, err := ctx.Cache().Get(oldCacheKey, cacheToken)
		if err != nil {
			c.SetMessage("failed to get cache token")
			ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
			return true, err
		}
		if !found {
			c.SetMessage("cache token expired")
			ctx.SetGenericErrorCode(ErrorCodeTokenExpired)
			return true, err
		}
		ctx.Cache().Unset(oldCacheKey)

		// check tries count
		if cacheToken.Try >= a.MAX_TRIES {
			err = errors.New("too many tries")
			ctx.SetGenericErrorCode(ErrorCodeTooManyTries)
			return true, err
		}

		// check if this is the same request as initial
		h := a.hmacOfRequest(ctx, userId)
		err = h.CheckStr(cacheToken.Checksum)
		if err != nil {
			c.SetMessage("invalid request checksum")
			ctx.SetGenericErrorCode(ErrorCodeContentMismatch)
			return false, err
		}

		// check SMS code
		if code != cacheToken.Code {

			// bad SMS code

			// regenerate token
			token.GenerateID()
			token.SetTTL(a.TOKEN_TTL_SECONDS)

			// keep cache token with increased tries count
			cacheToken.Try += 1
			err = a.setToken(ctx, c, cacheToken, token)
			if err != nil {
				return true, err
			}

			// done
			ctx.SetGenericErrorCode(ErrorCodeInvalidSmsCode)
			err = errors.New("invalid SMS code")
			return true, err
		}

		// good SMS code

		// unset SMS delay
		ctx.Cache().Unset(a.smsDelayCacheKey(userId))

		// done
		return true, nil
	}

	// SMS code is not set in request

	// check if SMS delay expired
	delayCacheKey := a.smsDelayCacheKey(userId)
	delayItem := &SmsDelay{}
	found, err := ctx.Cache().Get(delayCacheKey, delayItem)
	if err != nil {
		c.SetMessage("failed to get delay item from cache")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return true, err
	}
	if found {
		// set delay parameter in request
		now := time.Now()
		diff := now.Sub(delayItem.GetCreatedAt())
		delay := int(diff.Seconds())
		if delay > a.SMS_DELAY_SECONDS {
			delay = a.SMS_DELAY_SECONDS
		}
		ctx.SetAuthParameter(SmsProtocol, DelayName, fmt.Sprintf("%d", delay))

		// done
		err = errors.New("wait for delay")
		ctx.SetGenericErrorCode(ErrorCodeWaitDelay)
		return true, err
	}

	// prepare SMS
	message := ""
	err = ctx.CheckRequestContent(&message)
	if err != nil {
		c.SetMessage("failed to check request content")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeFormat)
		return true, err
	}

	// user must be of UserWithPhone interface
	user, ok := ctx.AuthUser().(UserWithPhone)
	if !ok {
		c.SetMessage("user must be of UserWithPhone interface")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return true, err
	}

	// get user's phone number
	phone := user.Phone()
	if phone == "" {
		c.SetMessage("unknown phone number")
		ctx.SetGenericErrorCode(ErrorCodeInvalidPhone)
		return true, err
	}
	ctx.SetAuthParameter(SmsProtocol, PhoneName, utils.MaskPhone(phone))

	// prepare token
	token := &SmsToken{}
	token.GenerateID()
	cacheToken := &SmsCacheToken{}
	cacheToken.Code = a.genCode()
	cacheToken.Try = 1
	h := a.hmacOfRequest(ctx, userId)
	cacheToken.Checksum = h.SumStr()

	// send SMS
	if message == "" {
		message = "code %s"
	}
	message = fmt.Sprintf(message, cacheToken.Code)
	cacheToken.SmsId, err = a.smsManager.Send(ctx, message, phone)
	if err != nil {
		c.SetMessage("failed to send SMS")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return true, err
	}

	// set token
	err = a.setToken(ctx, c, cacheToken, token)
	if err != nil {
		return true, err
	}

	// save SMS delay in cache
	delayItem.InitCreatedAt()
	err1 := ctx.Cache().Set(delayCacheKey, delayItem, a.SMS_DELAY_SECONDS)
	if err1 != nil {
		c.Logger().Error("failed to save SMS delay item in cache", err1)
	}
	// set delay parameter
	ctx.SetAuthParameter(SmsProtocol, DelayName, fmt.Sprintf("%d", a.SMS_DELAY_SECONDS))

	// set response code
	ctx.SetGenericErrorCode(ErrorCodeSmsConfirmationRequired)

	// done
	return true, errors.New("sms code not found")
}

func (a *AuthSms) hmacOfRequest(ctx auth.AuthContext, userId string) *crypt_utils.Hmac {
	h := crypt_utils.NewHmac(a.SECRET)
	h.Calc([]byte(userId), []byte(ctx.GetRequestMethod()), []byte(ctx.GetRequestPath()), ctx.GetRequestContent())
	return h
}

func (a *AuthSms) smsDelayCacheKey(userId string) string {
	return fmt.Sprintf("%s/%s", SmsDelayCacheKey, userId)
}

func (a *AuthSms) smsTokenCacheKey(userId string) string {
	return fmt.Sprintf("%s/%s", SmsTokenCacheKey, userId)
}

func (a *AuthSms) genCode() string {
	r := rand.Uint32()
	str := fmt.Sprintf("%08d", r)
	return str[len(str)-a.CODE_LENGTH:]
}

func (a *AuthSms) setToken(ctx auth.AuthContext, c op_context.CallContext, cacheToken *SmsCacheToken, requestToken *SmsToken) error {

	// keep in cache
	newCacheKey := a.smsTokenCacheKey(requestToken.GetID())
	err := ctx.Cache().Set(newCacheKey, cacheToken, a.TOKEN_TTL_SECONDS)
	if err != nil {
		c.SetMessage("failed to save token in cache")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return err
	}

	// put token to response
	err = a.Encryption.SetAuthParameter(ctx, a.Protocol(), TokenName, requestToken)
	if err != nil {
		c.SetMessage("failed to put token to response")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return err
	}

	return nil
}
