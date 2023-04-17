package confirmation_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/confirmation_control/confirmation_control_api"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
)

type ConfirmationCallbackClient struct {
	api_client.ServiceClient

	CallbackResource api.Resource
}

func NewConfirmationCallbackClient(client api_client.Client) *ConfirmationCallbackClient {

	c := &ConfirmationCallbackClient{}

	c.Init(client, confirmation_control_api.ServiceName)

	c.CallbackResource = api.NamedResource(confirmation_control_api.CallbackResource)
	c.AddChild(c.CallbackResource.Parent())

	return c
}

func (cl *ConfirmationCallbackClient) ConfirmationCallback(ctx multitenancy.TenancyContext, operationId string, codeOrStatus string) (string, error) {

	// setup
	c := ctx.TraceInMethod("ConfirmationCallbackClient.ConfirmationCallback")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// prepare and exec handler
	cmd := &confirmation_control_api.CallbackConfirmationCmd{
		CodeOrStatus: codeOrStatus,
	}
	handler := api_client.NewHandler(cmd, &confirmation_control_api.CallbackConfirmationResponse{})
	op := api.NamedResourceOperation(cl.CallbackResource, operationId, confirmation_control_api.CallbackConfirmation())
	err = op.Exec(ctx, api_client.MakeOperationHandler(cl.ApiClient(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return "", err
	}

	// done
	return handler.Result.Url, nil
}
