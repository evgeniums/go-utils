package auth_csrf

import (
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const AntiCsrfProtocol = "csrf"

type AuthCsrfConfig struct {
	common.WithNameBaseConfig
	SECRET     string `validate:"required"`
	EXPIRATION uint   `default:"300"`
}

type AuthCsrf struct {
	auth.AuthHandlerBase
	AuthCsrfConfig
}

func (a *AuthCsrf) Config() interface{} {
	return a
}

func (a *AuthCsrf) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	err := object_config.LoadLogValidate(cfg, log, vld, a, "auth.methods.csrf", configPath...)
	if err != nil {
		return log.Fatal("failed to load application configuration", err)
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

	return true, nil
}
