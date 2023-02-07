package auth_token

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const NewTokenProtocol = "new_token"

type AuthNewTokenHandler struct {
	AuthTokenHandler
}

func NewNewToken(users user_manager.WithUserSessionManager) *AuthNewTokenHandler {
	a := &AuthNewTokenHandler{}
	a.users = users
	return a
}

func (a *AuthNewTokenHandler) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	err := a.AuthTokenHandler.Init(cfg, log, vld, configPath...)
	if err != nil {
		return err
	}

	a.AuthHandlerBase.Init(NewTokenProtocol)

	return nil
}

func (a *AuthNewTokenHandler) Handle(ctx auth.AuthContext) (bool, error) {

	// setup
	c := ctx.TraceInMethod("AuthNewTokenHandler.Handle")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// check if user was already authenticated
	if ctx.AuthUser() == nil {
		return false, nil
	}

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

	// generate refresh token
	err = a.GenRefreshToken(ctx, session)
	if err != nil {
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return true, err
	}

	// done
	return true, nil
}
