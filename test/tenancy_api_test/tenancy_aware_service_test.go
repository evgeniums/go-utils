package tenancy_api_test

import (
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_methods/auth_token"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_providers/pubsub_factory"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/assert.v1"
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

type ListResponse = api.ResponseList[*InTenancySample]

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
	loadedTenancy1, err := multiPoolCtx.AppWithTenancy.Multitenancy().Tenancy(addedTenancy1.GetID())
	require.NoError(t, err)
	require.NotNil(t, loadedTenancy1)

	// add second tenancy to the same pool as single pool app, add via mutipool app
	tenancyData2 := tenancyData1
	tenancyData2.CUSTOMER_ID = "customer2"
	addedTenancy2, err := multiPoolCtx.RemoteTenancyController.Add(multiPoolCtx.ClientOp, tenancyData2)
	require.NoError(t, err)
	require.NotNil(t, addedTenancy2)
	loadedTenancy2, err := multiPoolCtx.AppWithTenancy.Multitenancy().Tenancy(addedTenancy2.GetID())
	require.NoError(t, err)
	require.NotNil(t, loadedTenancy1)

	// add document to tenancy database
	sample1 := &InTenancySample{Field1: "hello world", Field2: 10}
	err = loadedTenancy1.Db().Create(multiPoolCtx.AdminOp, sample1)
	require.NoError(t, err)

	// add sample service
	sampleService := NewSampleService()
	api_server.AddServiceToServer(multiPoolCtx.Server.ApiServer(), sampleService, true)

	// create sample client
	sampleClient := NewSampleClient(multiPoolCtx.RestApiClient)
	tenancyResource := api.NamedResource("tenancy")
	tenancyResource.AddChild(sampleClient)

	// invoke operation on tenancy 1
	tenancyResource.SetId(loadedTenancy1.Path())
	samples, err := sampleClient.List(multiPoolCtx.ClientOp)
	require.NoError(t, err)
	require.NotNil(t, samples)
	require.Equal(t, 1, len(samples))
	assert.Equal(t, sample1, samples[0])

	// invoke operation on tenancy 2
	tenancyResource.SetId(loadedTenancy2.Path())
	samples, err = sampleClient.List(multiPoolCtx.ClientOp)
	require.NoError(t, err)
	require.NotNil(t, samples)
	assert.Equal(t, 0, len(samples))

	// try to invoke operation on unknown tenancy
	tenancyResource.SetId("unknowntenancypath")
	_, err = sampleClient.List(multiPoolCtx.ClientOp)
	test_utils.CheckGenericError(t, err, generic_error.ErrorCodeNotFound)

	// init service for singlepool app
	singlePoolSampleService := NewSampleService()
	api_server.AddServiceToServer(singlePoolCtx.Server.ApiServer(), singlePoolSampleService, true)
	singlePoolSampleClient := NewSampleClient(singlePoolCtx.RestApiClient)
	singlePoolTenancyResource := api.NamedResource("tenancy")
	singlePoolTenancyResource.AddChild(singlePoolSampleClient)

	// try to invoke operation in single pool app where auth is from tenancies database
	singlePoolTenancyResource.SetId(loadedTenancy1.Path())
	_, err = singlePoolSampleClient.List(multiPoolCtx.ClientOp)
	test_utils.CheckGenericError(t, err, auth_token.ErrorCodeSessionExpired)

	// close apps
	multiPoolCtx.Close()
	singlePoolCtx.Close()

	// check disallowed not active tenancy
	multiPoolCtx = initContext(t, false, "tenancy_notactive")
	sampleService1 := NewSampleService()
	api_server.AddServiceToServer(multiPoolCtx.Server.ApiServer(), sampleService1, true)
	sampleClient1 := NewSampleClient(multiPoolCtx.RestApiClient)
	tenancyResource1 := api.NamedResource("tenancy")
	tenancyResource1.AddChild(sampleClient1)
	tenancyResource1.SetId(loadedTenancy1.Path())
	samples, err = sampleClient1.List(multiPoolCtx.ClientOp)
	require.NoError(t, err)
	require.NotNil(t, samples)
	require.Equal(t, 1, len(samples))
	assert.Equal(t, sample1, samples[0])
	tenancyResource1.SetId(loadedTenancy2.Path())
	_, err = sampleClient1.List(multiPoolCtx.ClientOp)
	test_utils.CheckGenericError(t, err, generic_error.ErrorCodeNotFound)

	// close apps
	multiPoolCtx.Close()
	pubsub_factory.ResetSingletonInmemPubsub()
}
