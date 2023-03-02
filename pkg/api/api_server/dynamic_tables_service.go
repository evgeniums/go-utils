package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
)

type DynamicTableEndpoint struct {
	ResourceEndpoint
	service *DynamicTablesService
}

func NewDynamicTableEndpoint(service *DynamicTablesService) *DynamicTableEndpoint {
	ep := &DynamicTableEndpoint{service: service}
	InitResourceEndpoint(ep, "table-config", "DynamicTableConfig", access_control.Get)
	return ep
}

func (e *DynamicTableEndpoint) HandleRequest(request Request) error {

	// setup
	c := request.TraceInMethod("GetDynamicTable")
	defer request.TraceOutMethod()

	// parse command
	cmd := &DynamicTableQuery{}
	err := request.ParseValidate(cmd)
	if err != nil {
		c.SetMessage("failed to parse/validate command")
		return err
	}

	// get table
	table, err := e.service.Server().DynamicTables().Table(request, cmd.Path)
	if err != nil {
		c.SetMessage("failed to find table for path")
		request.SetGenericErrorCode(generic_error.ErrorCodeNotFound)
		return err
	}

	// set response
	request.Response().SetMessage(table)

	// done
	return nil
}

type DynamicTablesService struct {
	ServiceBase
}

func NewDynamicTablesService() *DynamicTablesService {

	s := &DynamicTablesService{}

	s.Init("dynamic-tables")
	s.AddChild(NewDynamicTableEndpoint(s))

	return s
}
