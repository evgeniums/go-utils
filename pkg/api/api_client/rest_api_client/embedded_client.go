package rest_api_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type EmbeddedClientConfig struct {
	ClientBase
	TOKEN_CACHE_KEY string `validate:"required" default:"client_refresh_token"`
	LOGIN           string `validate:"required"`
	PASSWORD        string `mask:"true"`
}

type EmbeddedClient struct {
	app_context.WithAppBase
	EmbeddedClientConfig
	Client *RestApiClientBase
}

func NewEmbeddedClient(app app_context.Context) *EmbeddedClient {
	e := &EmbeddedClient{}
	e.WithAppBase.Init(app)
	e.Client = AutoReconnectRestApiClient(e)
	return e
}

func (e *EmbeddedClient) Config() interface{} {
	return e.EmbeddedClientConfig
}

func (e *EmbeddedClient) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	err := object_config.LoadLogValidate(cfg, log, vld, e, "rest_api_client", configPath...)
	if err != nil {
		return log.PushFatalStack("failed to load configuration of rest api client", err)
	}

	e.Client.Init(e.BASE_URL, e.USER_AGENT)

	return nil
}

func (e *EmbeddedClient) Setup(ctx op_context.Context, baseUrl string, login string, password string, tokenCacheKey string, userAgent ...string) error {

	e.BASE_URL = baseUrl
	e.LOGIN = login
	e.PASSWORD = password
	e.TOKEN_CACHE_KEY = tokenCacheKey
	e.USER_AGENT = utils.OptionalArg("go-backend-helpers", userAgent...)
	e.Client.Init(e.BASE_URL, e.USER_AGENT)

	return nil
}

type TokenKeeper struct {
	Token string `json:"token"`
}

func (e *EmbeddedClient) GetRefreshToken() string {
	tokenKeeper := &TokenKeeper{}
	found, err := e.App().Cache().Get(e.TOKEN_CACHE_KEY, tokenKeeper)
	if found && err != nil {
		return tokenKeeper.Token
	}
	e.App().Logger().Warn("client refresh token not found in cache")
	return ""
}

func (e *EmbeddedClient) SaveRefreshToken(ctx op_context.Context, token string) {
	tokenKeeper := &TokenKeeper{Token: token}
	err := e.App().Cache().Set(e.TOKEN_CACHE_KEY, tokenKeeper)
	if err != nil {
		ctx.Logger().Error("failed to save client refresh token in cache", err)
	}
}

func (e *EmbeddedClient) GetCredentials(ctx op_context.Context) (login string, password string, err error) {
	return e.LOGIN, e.PASSWORD, nil
}
