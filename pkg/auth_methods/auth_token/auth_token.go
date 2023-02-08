package auth_token

import (
	"errors"
	"net/http"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_session"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const CheckTokenProtocol = "check_token"
const TokenProtocol = "token"
const AccessTokenName = "access-token"
const RefreshTokenName = "refresh-token"

type AuthTokenHandlerConfig struct {
	ACCESS_TOKEN_TTL_SECONDS  int    `default:"900" validate:"gt=0"`
	REFRESH_TOKEN_TTL_SECONDS int    `default:"43200" validate:"gt=0"`
	AUTO_PROLONGATE_ACCESS    bool   `default:"true"`
	AUTO_PROLONGATE_REFRESH   bool   `default:"true"`
	REFRESH_PATH              string `default:"/auth/refresh"`
	LOGOUT_PATH               string `default:"/auth/logout"`
}

type AuthTokenHandler struct {
	auth.AuthHandlerBase
	AuthTokenHandlerConfig
	users      auth_session.WithUserSessionManager
	encryption auth.AuthParameterEncryption
}

type Token struct {
	auth.ExpireToken
	Id          string `json:"id"`
	UserId      string `json:"user_id"`
	UserDisplay string `json:"user_display"`
	SessionId   string `json:"session_id"`
	Tenancy     string `json:"tenancy"`
}

func (a *AuthTokenHandler) Config() interface{} {
	return &a.AuthTokenHandlerConfig
}

func New(users auth_session.WithUserSessionManager) *AuthTokenHandler {
	a := &AuthTokenHandler{}
	a.users = users
	return a
}

func (a *AuthTokenHandler) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	a.AuthHandlerBase.Init(CheckTokenProtocol)

	path := utils.OptionalArg("auth.methods.token", configPath...)

	err := object_config.LoadLogValidate(cfg, log, vld, a, path)
	if err != nil {
		return log.PushFatalStack("failed to load configuration of TOKEN handler", err)
	}

	encryption := &auth.AuthParameterEncryptionBase{}
	err = encryption.Init(cfg, log, vld, path)
	if err != nil {
		return log.PushFatalStack("failed to load configuration of CSRF encryption", err)
	}
	a.encryption = encryption

	return nil
}

const ErrorCodeTokenExpired = "auth_token_expired"
const ErrorCodeInvalidToken = "auth_token_invalid"
const ErrorCodeSessionExpired = "session_expired"
const ErrorCodeUnknownUser = "unknown_user"

func (a *AuthTokenHandler) ErrorDescriptions() map[string]string {
	m := map[string]string{
		ErrorCodeTokenExpired:   "Provided authentication token token expired",
		ErrorCodeInvalidToken:   "Invalid authentication token token",
		ErrorCodeSessionExpired: "Session expired",
		ErrorCodeUnknownUser:    "Unknown user",
	}
	return m
}

func (a *AuthTokenHandler) ErrorProtocolCodes() map[string]int {
	m := map[string]int{
		ErrorCodeTokenExpired:   http.StatusUnauthorized,
		ErrorCodeInvalidToken:   http.StatusUnauthorized,
		ErrorCodeSessionExpired: http.StatusUnauthorized,
		ErrorCodeUnknownUser:    http.StatusUnauthorized,
	}
	return m
}

func (a *AuthTokenHandler) Handle(ctx auth.AuthContext) (bool, error) {

	// setup
	c := ctx.TraceInMethod("AuthTokenHandler.Handle")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// chek if it is REFRESH token request or normal access
	path := ctx.GetRequestPath()
	refresh := path == a.REFRESH_PATH
	tokenName := AccessTokenName
	if refresh {
		tokenName = RefreshTokenName
	}
	c.LoggerFields()["refresh"] = refresh

	// check token in request
	prev := &Token{}
	exists, err := a.encryption.GetAuthParameter(ctx, a.Protocol(), tokenName, prev)
	if !exists {
		return false, err
	}
	if err != nil {
		c.SetMessage("failed to get encrypted auth parameter")
		ctx.SetGenericErrorCode(ErrorCodeInvalidToken)
		return true, err
	}
	c.LoggerFields()["token"] = prev.Id
	ctx.SetLoggerField("user", prev.UserDisplay)
	ctx.SetLoggerField("session", prev.SessionId)
	if prev.Expired() {
		c.SetMessage("token expired")
		if refresh {
			ctx.SetGenericErrorCode(ErrorCodeSessionExpired)
		} else {
			ctx.SetGenericErrorCode(ErrorCodeTokenExpired)
		}
		return true, err
	}

	// check tenancy
	if ctx.GetTenancy() != nil || prev.Tenancy != "" {
		if ctx.GetTenancy() == nil || prev.Tenancy != ctx.GetTenancy().GetID() {
			err = errors.New("invalid tenancy")
			c.LoggerFields()["token_tenancy"] = prev.Tenancy
			ctx.SetGenericErrorCode(ErrorCodeSessionExpired)
			return true, err
		}
	}

	// find session
	session, err := a.users.SessionManager().FindSession(ctx, prev.SessionId)
	if err != nil {
		ctx.SetGenericErrorCode(ErrorCodeSessionExpired)
		return true, err
	}

	// check if session is valid
	if !session.IsValid() {
		err = errors.New("session invalidated")
		ctx.SetGenericErrorCode(ErrorCodeSessionExpired)
		return true, err
	}
	now := time.Now()
	if now.After(session.GetExpiration()) {
		err = errors.New("session expired")
		ctx.SetGenericErrorCode(ErrorCodeSessionExpired)
		return true, err
	}

	// load user
	user := a.users.AuthUserManager().MakeAuthUser()
	found, err := a.users.AuthUserManager().FindAuthUser(ctx, session.GetUserLogin(), user)
	if err != nil {
		c.SetMessage("failed to load user")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return true, err
	}
	if !found {
		c.SetMessage("user not found")
		ctx.SetGenericErrorCode(ErrorCodeUnknownUser)
		return true, err
	}
	ctx.SetLoggerField("user", user.Display())

	// check if user blocked
	if user.IsBlocked() {
		err = errors.New("user blocked")
		ctx.SetGenericErrorCode(ErrorCodeSessionExpired)
		return true, err
	}

	// set auth user
	ctx.SetAuthUser(user)

	// set user session
	ctx.SetSessionId(session.GetID())

	// update session client
	err = a.users.SessionManager().UpdateSessionClient(ctx)
	if err != nil {
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return true, err
	}

	// add tokens if applicable
	if path != a.LOGOUT_PATH {

		if refresh || !refresh && a.AUTO_PROLONGATE_ACCESS {
			// generate access token
			err = a.GenAccessToken(ctx)
			if err != nil {
				ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
				return true, err
			}
		}

		tokenExpirationTime := now.Add(time.Second * time.Duration(a.ACCESS_TOKEN_TTL_SECONDS))
		regenerateRefreshToken := a.AUTO_PROLONGATE_REFRESH && (refresh || tokenExpirationTime.After(session.GetExpiration()))
		if regenerateRefreshToken {
			// generate refresh token
			err = a.GenRefreshToken(ctx, session)
			if err != nil {
				ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
				return true, err
			}
		}

	} else {

		ctx.SetAuthUser(nil)

		// invalidate session on logout path
		err = a.users.SessionManager().InvalidateSession(ctx, session.GetUserId(), session.GetID())
		if err != nil {
			ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
			return true, err
		}
	}

	// done
	return true, nil
}

func (a *AuthTokenHandler) GenAccessToken(ctx auth.AuthContext) error {
	c := ctx.TraceInMethod("AuthTokenHandler.GenAccessToken")
	defer ctx.TraceOutMethod()

	return c.SetError(a.GenToken(ctx, AccessTokenName, a.ACCESS_TOKEN_TTL_SECONDS))
}

func (a *AuthTokenHandler) GenRefreshToken(ctx auth.AuthContext, session auth_session.Session) error {
	c := ctx.TraceInMethod("AuthTokenHandler.GenRefreshToken")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	expirationSeconds := a.REFRESH_TOKEN_TTL_SECONDS
	session.SetExpiration(a.SessionExpiration())
	err = a.users.SessionManager().UpdateSessionExpiration(ctx, session)
	if err != nil {
		c.SetMessage("failed to update session expiration")
		return err
	}
	err = a.GenToken(ctx, RefreshTokenName, expirationSeconds)
	return err
}

func (a *AuthTokenHandler) GenToken(ctx auth.AuthContext, paramName string, expirationSeconds int) error {

	c := ctx.TraceInMethod("AuthTokenHandler.GenToken")
	defer ctx.TraceOutMethod()

	token := &Token{}
	token.Id = utils.GenerateRand64()
	token.SessionId = ctx.GetSessionId()
	token.UserDisplay = ctx.AuthUser().Display()
	token.UserId = ctx.AuthUser().GetID()
	if ctx.GetTenancy() != nil {
		token.Tenancy = ctx.GetTenancy().GetID()
	}

	token.SetTTL(expirationSeconds)
	return c.SetError(a.encryption.SetAuthParameter(ctx, a.Protocol(), paramName, token))
}

func (a *AuthTokenHandler) SessionExpiration() time.Time {
	expirationSeconds := a.REFRESH_TOKEN_TTL_SECONDS
	return time.Now().Add(time.Second * time.Duration(expirationSeconds))
}

func (a *AuthTokenHandler) SetAuthManager(manager auth.AuthManager) {
	manager.Schemas().AddHandler(a)
}

type TokenSchema struct {
	auth.AuthSchema

	Token *AuthTokenHandler
}

func NewSchema(users auth_session.WithUserSessionManager) *TokenSchema {
	l := &TokenSchema{}
	l.Construct()
	l.Token = New(users)
	return l
}

func (t *TokenSchema) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	t.AuthHandlerBase.Init(TokenProtocol)

	err := t.Token.Init(cfg, log, vld, configPath...)
	if err != nil {
		return log.PushFatalStack("failed to init token handler", err)
	}

	t.AuthSchema.AppendHandlers(t.Token)
	return nil
}

func (t *TokenSchema) Handlers() []auth.AuthHandler {
	return t.AuthSchema.Handlers()
}

func (t *TokenSchema) SetAuthManager(manager auth.AuthManager) {
	manager.Schemas().AddHandler(t)
}

func ReloginRequired(code string) bool {
	return code == ErrorCodeInvalidToken || code == ErrorCodeSessionExpired || code == ErrorCodeUnknownUser
}

func RefreshRequired(code string) bool {
	return code == ErrorCodeTokenExpired
}
