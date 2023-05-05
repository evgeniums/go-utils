package api_server

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type ServiceEachEndpointHandler = func(ep Endpoint)

// Interface of service that implements one or more groups of API endpoints.
type Service interface {
	generic_error.ErrorsExtender
	api.Resource

	SetSupportsMultitenancy(enable bool)
	SupportsMultitenancy() bool

	Server() Server
	AttachToServer(server Server) error

	AddDynamicTables(tables ...*DynamicTableConfig)
	DynamicTables() []*DynamicTableConfig
}

type ServiceBase struct {
	api.ResourceBase
	generic_error.ErrorsExtenderBase
	server        Server
	dynamicTables []*DynamicTableConfig

	multitenancy bool
}

func (s *ServiceBase) Init(pathName string, multitenancy ...bool) {
	s.ResourceBase.Init(pathName, api.ResourceConfig{Service: true})
	s.dynamicTables = make([]*DynamicTableConfig, 0)
	s.multitenancy = utils.OptionalArg(false, multitenancy...)
}

func (s *ServiceBase) SetSupportsMultitenancy(enable bool) {
	s.multitenancy = enable
}

func (s *ServiceBase) SupportsMultitenancy() bool {
	return s.multitenancy
}

func (s *ServiceBase) Server() Server {
	return s.server
}

func (s *ServiceBase) DynamicTables() []*DynamicTableConfig {
	return s.dynamicTables
}

func (s *ServiceBase) AddDynamicTables(tables ...*DynamicTableConfig) {
	s.dynamicTables = append(s.dynamicTables, tables...)
}

func (s *ServiceBase) AttachToServer(server Server) error {

	s.server = server

	dynamicTables := server.DynamicTables()
	if dynamicTables != nil {
		for _, dynamicTable := range s.DynamicTables() {
			server.DynamicTables().AddTable(dynamicTable)
		}
	}

	return s.EachOperation(func(op api.Operation) error {
		ep, ok := op.(Endpoint)
		if !ok {
			return fmt.Errorf("invalid opertaion type, must be endpoint: %s", op.Name())
		}
		server.AddEndpoint(ep, s.SupportsMultitenancy())
		return nil
	})
}
