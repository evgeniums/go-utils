package confirmation_api_client

import (
	"github.com/evgeniums/go-utils/pkg/api"
	"github.com/evgeniums/go-utils/pkg/api/api_client"
	"github.com/evgeniums/go-utils/pkg/confirmation_control/confirmation_control_api"
	"github.com/evgeniums/go-utils/pkg/multitenancy"
)

type ConfirmationInternalClient struct {
	api_client.ServiceClient

	OperationResource api.Resource
	prepare_operation api.Operation
}

func NewConfirmationInternalClient(client api_client.Client) *ConfirmationInternalClient {

	c := &ConfirmationInternalClient{}

	c.Init(client, confirmation_control_api.ServiceName)

	c.OperationResource = api.NewResource(confirmation_control_api.OperationResource)
	c.AddChild(c.OperationResource)

	c.prepare_operation = confirmation_control_api.PrepareOperation()
	c.OperationResource.AddOperation(c.prepare_operation)

	api.NewTenancyResource().AddChild(c)

	return c
}

func (cl *ConfirmationInternalClient) SendConfirmation(ctx multitenancy.TenancyContext, operationId string, recipient string, failedUrl string, parameters ...map[string]interface{}) (string, error) {

	// setup
	c := ctx.TraceInMethod("ConfirmationInternalClient.SendConfirmation")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// prepare and exec handler
	cmd := &confirmation_control_api.PrepareOperationCmd{
		Id:        operationId,
		Recipient: recipient,
		FailedUrl: failedUrl,
	}
	if len(parameters) != 0 {
		cmd.Parameters = parameters[0]
		ttlI, ok := cmd.Parameters["ttl"]
		if ok {
			ttl, ok := ttlI.(int)
			if ok && ttl > 0 {
				cmd.Ttl = ttl
			}
		}
	}
	handler := api_client.NewHandlerInTenancy(cmd, &confirmation_control_api.PrepareOperationResponse{})
	err = cl.prepare_operation.ExecInTenancy(ctx, api_client.MakeTenancyOperationHandler(cl.ApiClient(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return "", err
	}

	// done
	return handler.Result.Url, nil
}
