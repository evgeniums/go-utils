package rest_api_client

import (
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type RestApiMethod func(ctx op_context.Context, path string, cmd interface{}, response interface{}, headers ...map[string]string) (Response, error)

type Client struct {
	restApiClient RestApiClient
	methods       map[access_control.AccessType]RestApiMethod
}

func New(restApiClient RestApiClient) *Client {
	c := &Client{restApiClient: restApiClient}
	c.methods = make(map[access_control.AccessType]RestApiMethod, 0)

	c.methods[access_control.HttpMethod2Access(http.MethodPost)] = c.restApiClient.Post
	c.methods[access_control.HttpMethod2Access(http.MethodPut)] = c.restApiClient.Put
	c.methods[access_control.HttpMethod2Access(http.MethodPatch)] = c.restApiClient.Patch
	c.methods[access_control.HttpMethod2Access(http.MethodGet)] = c.restApiClient.Get
	c.methods[access_control.HttpMethod2Access(http.MethodDelete)] = c.restApiClient.Delete

	return c
}

func (cl *Client) Exec(ctx op_context.Context, operation api.Operation, cmd interface{}, response interface{}) generic_error.Error {

	// TODO support hateoas links of resource

	c := ctx.TraceInMethod("RestApiClientBase.Login")
	defer ctx.TraceOutMethod()

	method, ok := cl.methods[operation.AccessType()]
	if !ok {
		c.SetLoggerField("access_type", operation.AccessType())
		genericError := generic_error.NewFromMessage("access type not supported")
		c.SetError(genericError)
		return genericError
	}
	resp, err := method(ctx, operation.Resource().ActualPath(), cmd, response)
	genericError := api.ResponseGenericError(resp.Error())
	if err != nil {
		genericError.SetOriginal(err)
		c.SetError(err)
	}

	return genericError
}
