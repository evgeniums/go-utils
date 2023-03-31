package rest_api_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
)

type RestApiClientWithConfigCfg struct {
	BASE_URL   string `validate:"required"`
	USER_AGENT string `default:"go-backend-helpers"`
}

type RestApiClientWithConfig struct {
	RestApiClientWithConfigCfg
	*RestApiClientBase
}

func (r *RestApiClientWithConfig) Config() interface{} {
	return &r.RestApiClientWithConfigCfg
}

func (r *RestApiClientWithConfig) Init(app app_context.Context, configPath ...string) error {

	err := object_config.LoadLogValidate(app.Cfg(), app.Logger(), app.Validator(), r, "rest_api_client", configPath...)
	if err != nil {
		return app.Logger().PushFatalStack("failed to load configuration of rest api client", err)
	}

	r.RestApiClientBase.Init(r.BASE_URL, r.USER_AGENT)

	return nil
}

func NewRestApiClientWithConfig(withBodySender DoRequest, withQuerySender DoRequest) *RestApiClientWithConfig {
	r := &RestApiClientWithConfig{}
	r.RestApiClientBase = &RestApiClientBase{}
	r.Construct(withBodySender, withQuerySender)
	return r
}
