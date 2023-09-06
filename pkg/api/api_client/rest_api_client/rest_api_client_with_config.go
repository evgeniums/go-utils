package rest_api_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/http_request"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type RestApiClientWithConfigCfg struct {
	BASE_URL     string `validate:"required"`
	USER_AGENT   string `default:"go-backend-helpers"`
	TENANCY_TYPE string `default:"tenancy"`
	TENANCY_PATH string
}

type RestApiClientWithConfig struct {
	RestApiClientWithConfigCfg
	*RestApiClientBase
}

func (r *RestApiClientWithConfig) Config() interface{} {
	return &r.RestApiClientWithConfigCfg
}

func (r *RestApiClientWithConfig) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	path := utils.OptionalString("rest_api_client", configPath...)
	err := object_config.LoadLogValidate(cfg, log, vld, r, path)
	if err != nil {
		return log.PushFatalStack("failed to load configuration of rest api client", err)
	}

	httpClient := http_request.NewHttpClient()
	httpClientConfigPath := object_config.Key(path, "http_client")
	err = httpClient.Init(cfg, log, vld, httpClientConfigPath)
	if err != nil {
		return log.PushFatalStack("failed to init http client in rest api client", err)
	}

	if r.TENANCY_PATH != "" {
		tenancy := &TenancyAuth{TenancyType: r.TENANCY_TYPE, TenancyPath: r.TENANCY_PATH}
		r.RestApiClientBase.Init(httpClient, r.BASE_URL, r.USER_AGENT, tenancy)
	} else {
		r.RestApiClientBase.Init(httpClient, r.BASE_URL, r.USER_AGENT)
	}

	return nil
}

func NewRestApiClientWithConfig(withBodySender DoRequest, withQuerySender DoRequest) *RestApiClientWithConfig {
	r := &RestApiClientWithConfig{}
	r.RestApiClientBase = NewRestApiClientBase(withBodySender, withQuerySender)
	return r
}

func DefaultRestApiClientWithConfig() *RestApiClientWithConfig {
	c := NewRestApiClientWithConfig(DefaultSendWithBody, DefaultSendWithQuery)
	return c
}
