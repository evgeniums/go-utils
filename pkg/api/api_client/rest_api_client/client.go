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
	RestApiClient RestApiClient
	methods       map[access_control.AccessType]RestApiMethod
}

func New(restApiClient RestApiClient) *Client {
	c := &Client{RestApiClient: restApiClient}
	c.methods = make(map[access_control.AccessType]RestApiMethod, 0)

	c.methods[access_control.HttpMethod2Access(http.MethodPost)] = c.RestApiClient.Post
	c.methods[access_control.HttpMethod2Access(http.MethodPut)] = c.RestApiClient.Put
	c.methods[access_control.HttpMethod2Access(http.MethodPatch)] = c.RestApiClient.Patch
	c.methods[access_control.HttpMethod2Access(http.MethodGet)] = c.RestApiClient.Get
	c.methods[access_control.HttpMethod2Access(http.MethodDelete)] = c.RestApiClient.Delete

	return c
}

func (cl *Client) Transport() interface{} {
	return cl.RestApiClient
}

func (cl *Client) Exec(ctx op_context.Context, operation api.Operation, cmd interface{}, response interface{}, tenancyId ...string) error {

	// TODO support hateoas links of resource

	c := ctx.TraceInMethod("Client.Exec")
	defer ctx.TraceOutMethod()

	method, ok := cl.methods[operation.AccessType()]
	if !ok {
		c.SetLoggerField("access_type", operation.AccessType())
		genericError := generic_error.NewFromMessage("access type not supported")
		c.SetError(genericError)
		return genericError
	}
	var path string
	if len(tenancyId) == 0 {
		path = operation.Resource().FullActualPath()
	} else {
		path = operation.Resource().FullActualTenancyPath(tenancyId[0])
	}
	resp, err := method(ctx, path, cmd, response)
	if err != nil {
		c.SetMessage("failed to invoke HTTP method")
		return c.SetError(err)
	}
	genericError := api.ResponseGenericError(resp.Error())
	if genericError != nil {
		c.SetLoggerField("response_code", genericError.Code())
		c.SetLoggerField("response_message", genericError.Message())
		c.SetLoggerField("response_details", genericError.Details())
		c.SetMessage("server returned error")
		return c.SetError(genericError)
	}

	return nil
}
