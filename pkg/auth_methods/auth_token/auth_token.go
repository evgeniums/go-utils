package auth_token

import (
	"errors"
	"net/http"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const TokenProtocol = "token"
const AccessTokenName = "access-token"
const RefreshTokenName = "refresh-token"

type AuthTokenHandlerConfig struct {
	ACCESS_TOKEN_TTL_SECONDS  int    `default:"900" validate:"gt=0"`
	REFRESH_TOKEN_TTL_MINUTES int    `default:"720" validate:"gt=0"`
	AUTO_PROLONGATE_ACCESS    bool   `default:"true"`
	AUTO_PROLONGATE_REFRESH   bool   `default:"true"`
	REFRESH_PATH              string `default:"/auth/refresh"`
	LOGOUT_PATH               string `default:"/auth/logout"`
}

type AuthTokenHandler struct {
	auth.AuthHandlerBase
	AuthTokenHandlerConfig
	users      user_manager.WithSessionManager
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

func New(users user_manager.WithSessionManager) *AuthTokenHandler {
	a := &AuthTokenHandler{}
	a.users = users
	return a
}

func (a *AuthTokenHandler) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	a.AuthHandlerBase.Init(TokenProtocol)

	err := object_config.LoadLogValidate(cfg, log, vld, a, "auth.methods.token", configPath...)
	if err != nil {
		return log.PushFatalStack("failed to load configuration of TOKEN handler", err)
	}

	encryption := &auth.AuthParameterEncryptionBase{}
	err = object_config.LoadLogValidate(cfg, log, vld, encryption, "auth.methods.csrf", configPath...)
	if err != nil {
		return log.PushFatalStack("failed to load configuration of TOKEN encryption", err)
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

	// check if user was already authenticated
	if ctx.AuthUser() != nil {

		// user was authenticated, just create or update session client and add tokens

		sessionId := ctx.GetSessionId()
		var session user_manager.Session
		if sessionId == "" {
			// create session
			session, err = a.users.SessionManager().CreateSession(ctx, a.SessionExpiration())
			if err != nil {
				ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
				return true, err
			}
			ctx.SetSessionId(session.GetID())
		} else {
			// find session
			session, err = a.users.SessionManager().FindSession(ctx, sessionId)
			if err != nil {
				ctx.SetGenericErrorCode(ErrorCodeSessionExpired)
				return true, err
			}
		}

		// update session client
		err = a.users.SessionManager().UpdateSessionClient(ctx)
		if err != nil {
			ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
			return true, err
		}

		// generate refresh token
		err = a.GenRefreshToken(ctx, session)
		if err != nil {
			ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
			return true, err
		}

		// generate access token
		err = a.GenAccessToken(ctx)
		if err != nil {
			ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
			return true, err
		}

		// done
		return true, nil
	}

	// chek if it is REFRESH token request or normal access
	path := ctx.GetRequestPath()
	refresh := path == a.REFRESH_PATH
	tokenName := AccessTokenName
	if refresh {
		tokenName = RefreshTokenName
	}
	c.Fields()["refresh"] = refresh

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
	c.Fields()["token"] = prev.Id
	ctx.SetLoggerField("user", prev.UserDisplay)
	ctx.SetLoggerField("session", prev.SessionId)
	if prev.Expired() {
		c.SetMessage("token expired")
		ctx.SetGenericErrorCode(ErrorCodeTokenExpired)
		return true, err
	}

	// check tenancy
	if ctx.Tenancy() != nil || prev.Tenancy != "" {
		if ctx.Tenancy() == nil || prev.Tenancy != ctx.Tenancy().GetID() {
			err = errors.New("invalid tenancy")
			c.Fields()["token_tenancy"] = prev.Tenancy
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
	user := a.users.UserManager().MakeUser()
	found, err := user_manager.FindByLogin(a.users.UserManager(), ctx, session.GetUserLogin(), user)
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

			if refresh && a.AUTO_PROLONGATE_REFRESH {
				// generate refresh token
				err = a.GenRefreshToken(ctx, session)
				if err != nil {
					ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
					return true, err
				}
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

func (a *AuthTokenHandler) GenRefreshToken(ctx auth.AuthContext, session user_manager.Session) error {
	c := ctx.TraceInMethod("AuthTokenHandler.GenRefreshToken")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	expirationSeconds := a.REFRESH_TOKEN_TTL_MINUTES * 60
	session.SetExpiration(a.SessionExpiration())
	err = a.users.SessionManager().UpdateSessionExpiration(ctx, session)
	if err != nil {
		c.SetMessage("failed to update session expiration")
		return err
	}
	err = a.GenToken(ctx, AccessTokenName, expirationSeconds)
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
	if ctx.Tenancy() != nil {
		token.Tenancy = ctx.Tenancy().GetID()
	}

	token.SetTTL(expirationSeconds)
	return c.SetError(a.encryption.SetAuthParameter(ctx, a.Protocol(), paramName, token))
}

func (a *AuthTokenHandler) SessionExpiration() time.Time {
	expirationSeconds := a.REFRESH_TOKEN_TTL_MINUTES * 60
	return time.Now().Add(time.Second * time.Duration(expirationSeconds))
}
