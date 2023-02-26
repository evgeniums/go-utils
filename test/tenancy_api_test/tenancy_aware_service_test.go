package tenancy_api_test

import (
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_providers/pubsub_factory"
	"github.com/stretchr/testify/require"
)

type SampleService struct {
	api_server.ServiceBase
	SampleResource api.Resource
}

func NewSampleService() *SampleService {

	s := &SampleService{}

	s.Init("samples")
	s.SampleResource = api.NewResource("sample")
	s.AddChild(s.SampleResource)
	s.SampleResource.AddOperation(List(s))

	return s
}

type SampleEndpoint struct {
	service *SampleService
	api_server.EndpointBase
}

func (e *SampleEndpoint) Construct(service *SampleService, op api.Operation) {
	e.service = service
	e.EndpointBase.Construct(op)
}

type ListEndpoint struct {
	SampleEndpoint
}

type ListResponse struct {
	api.ResponseCount
	api.ResponseHateous
	Items []*InTenancySample `json:"items"`
}

func (e *ListEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	var err error
	c := request.TraceInMethod("samples.List")
	defer request.TraceOutMethod()

	// get
	resp := &ListResponse{}
	resp.Count, err = request.Db().FindWithFilter(request, nil, &resp.Items)
	if err != nil {
		return c.SetError(err)
	}
	// set response message
	request.Response().SetMessage(resp)

	// done
	return nil
}

func List(s *SampleService) *ListEndpoint {
	e := &ListEndpoint{}
	e.Construct(s, api.List("list_samples"))
	return e
}

type SampleClient struct {
	api_client.ServiceClient
	SampleResource api.Resource
	list           api.Operation
}

func NewSampleClient(client api_client.Client) *SampleClient {

	c := &SampleClient{}

	c.Init(client, "samples")
	c.SampleResource = api.NewResource("sample")
	c.AddChild(c.SampleResource)
	c.list = api.List("list_samples")

	c.SampleResource.AddOperations(
		c.list,
	)

	return c
}

type ClientList struct {
	result *ListResponse
}

func (l *ClientList) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("ClientList")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, nil, l.result)
	c.SetError(err)
	return err
}

func (s *SampleClient) List(ctx op_context.Context) ([]*InTenancySample, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("SampleClient.List")
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// prepare and exec handler
	handler := &ClientList{
		result: &ListResponse{},
	}
	err = s.list.Exec(ctx, api_client.MakeOperationHandler(s.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return nil, err
	}

	// done
	return handler.result.Items, nil
}

func TestTenancyAwareService(t *testing.T) {

	t.Skip("fix sessions with tenancies")

	// prepare app with multiple pools and single pool
	multiPoolCtx, singlePoolCtx := PrepareAppWithTenancies(t)

	// add first tenancy to the same pool as single pool app, add via mutipool app
	tenancyData1 := &multitenancy.TenancyData{}
	tenancyData1.POOL_ID = "pool2"
	tenancyData1.ROLE = "dev"
	tenancyData1.DESCRIPTION = "tenancy for development"
	tenancyData1.CUSTOMER_ID = "customer1"
	addedTenancy1, err := multiPoolCtx.RemoteTenancyController.Add(multiPoolCtx.ClientOp, tenancyData1)
	require.NoError(t, err)
	require.NotNil(t, addedTenancy1)
	err = multiPoolCtx.RemoteTenancyController.Activate(multiPoolCtx.ClientOp, addedTenancy1.GetID())
	require.NoError(t, err)

	// check if tenancy was loaded by single app
	loadedTenancy1, err := singlePoolCtx.AppWithTenancy.Multitenancy().Tenancy(addedTenancy1.GetID())
	require.NoError(t, err)
	require.NotNil(t, loadedTenancy1)

	// add document to tenancy database
	sample1 := &InTenancySample{Field1: "hello world", Field2: 10}
	err = loadedTenancy1.Db().Create(singlePoolCtx.AdminOp, sample1)
	require.NoError(t, err)

	// add sample service
	sampleService := NewSampleService()
	api_server.AddServiceToServer(singlePoolCtx.Server.ApiServer(), sampleService, true)

	// create sample client
	sampleClient := NewSampleClient(singlePoolCtx.RestApiClient)
	tenancyResource := api.NamedResource("tenancy")
	tenancyResource.AddChild(sampleClient)

	// invoke operation
	tenancyResource.SetId(loadedTenancy1.Path())
	samples, err := sampleClient.List(singlePoolCtx.ClientOp)
	require.NoError(t, err)
	require.NotNil(t, samples)

	// close apps
	multiPoolCtx.Close()
	singlePoolCtx.Close()
	pubsub_factory.ResetSingletonInmemPubsub()
}
