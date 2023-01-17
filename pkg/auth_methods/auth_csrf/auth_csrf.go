package auth_csrf

import (
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const AntiCsrfProtocol = "csrf"
const AntiCsrfTokenName = "csrf-token"

type AuthCsrfConfig struct {
	common.WithNameBaseConfig
	EXPIRATION_SECONDS int `default:"300" validate:"gt=0"`
	IGNORE_PATHS       []string
}

type AuthCsrf struct {
	auth.AuthHandlerBase
	AuthCsrfConfig
	Encryption auth.AuthParameterEncryption
	skipPaths  map[string]bool
}

func (a *AuthCsrf) Config() interface{} {
	return a
}

func (a *AuthCsrf) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	err := object_config.LoadLogValidate(cfg, log, vld, a, "auth.methods.csrf", configPath...)
	if err != nil {
		return log.Fatal("failed to load configuration of CSRF handler", err)
	}

	encryption := &auth.AuthParameterEncryptionBase{}
	err = object_config.LoadLogValidate(cfg, log, vld, encryption, "auth.methods.csrf", configPath...)
	if err != nil {
		return log.Fatal("failed to load configuration of CSRF encryption", err)
	}
	a.Encryption = encryption

	a.skipPaths = make(map[string]bool)
	for _, path := range a.IGNORE_PATHS {
		a.skipPaths[path] = true
	}

	return nil
}

func (a *AuthCsrf) Protocol() string {
	return AntiCsrfProtocol
}

const ErrorCodeAntiCsrfRequired = "anti_csrf_token_required"
const ErrorCodeTokenExpired = "anti_csrf_token_expired"
const ErrorCodeInvalidToken = "anti_csrf_token_invalid"

func (a *AuthCsrf) ErrorDescriptions() map[string]string {
	m := map[string]string{
		ErrorCodeAntiCsrfRequired: "Request must be protected with anti-CSRF token",
		ErrorCodeTokenExpired:     "Anti-CSRF token expired",
		ErrorCodeInvalidToken:     "Invalid anti-CSRF token",
	}
	return m
}

func (a *AuthCsrf) ErrorProtocolCodes() map[string]int {
	m := map[string]int{
		ErrorCodeAntiCsrfRequired: http.StatusForbidden,
		ErrorCodeTokenExpired:     http.StatusForbidden,
		ErrorCodeInvalidToken:     http.StatusForbidden,
	}
	return m
}

func (a *AuthCsrf) Handle(ctx auth.AuthContext) (bool, error) {

	// setup
	c := ctx.TraceInMethod("AuthCsrf.Handle")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// check token in request
	_, skip := a.skipPaths[ctx.GetRequestPath()]
	if !skip {
		prev := &auth.ExpireToken{}
		exists, err := a.Encryption.GetAuthParameter(ctx, a.Protocol(), AntiCsrfTokenName, prev)
		if !exists {
			c.SetMessage("CSRF token not found")
			ctx.SetGenericErrorCode(ErrorCodeAntiCsrfRequired)
			return false, err
		}
		if err != nil {
			c.SetMessage("failed to get encrypted auth parameter")
			ctx.SetGenericErrorCode(ErrorCodeInvalidToken)
			return true, err
		}

		if prev.Expired() {
			c.SetMessage("token expired")
			ctx.SetGenericErrorCode(ErrorCodeTokenExpired)
			return true, err
		}
	}

	// set token in response
	next := &auth.ExpireToken{}
	next.SetTTL(a.EXPIRATION_SECONDS)
	err = a.Encryption.SetAuthParameter(ctx, a.Protocol(), AntiCsrfTokenName, next)
	if err != nil {
		c.SetMessage("failed to set encrypted auth parameter")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return true, err
	}

	// done
	return true, nil
}
