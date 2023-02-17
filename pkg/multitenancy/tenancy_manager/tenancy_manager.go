package tenancy_manager

import (
	"errors"
	"sync"

	"github.com/evgeniums/go-backend-helpers/pkg/crypt_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type TenancyManagerConfig struct {
	MULTITENANCY bool
	DB_PREFIX    string `validate:"required,aphanum" vmessage:"Invalid prefix for names of databases"`
}

func (s *TenancyManagerConfig) IsMultiTenancy() bool {
	return s.MULTITENANCY
}

type TenancyManager struct {
	TenancyManagerConfig
	mutex           sync.Mutex
	tenanciesById   map[string]multitenancy.Tenancy
	tenanciesByPath map[string]multitenancy.Tenancy
	controller      multitenancy.TenancyController
	pools           pool.PoolStore
	customers       customer.CustomerController
}

func NewTenancyManager(pools pool.PoolStore, controller multitenancy.TenancyController) *TenancyManager {
	m := &TenancyManager{}
	m.pools = pools
	m.controller = controller
	m.tenanciesById = make(map[string]multitenancy.Tenancy)
	m.tenanciesByPath = make(map[string]multitenancy.Tenancy)
	return m
}

func (s *TenancyManager) Tenancy(id string) (multitenancy.Tenancy, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	tenancy, ok := s.tenanciesById[id]
	if !ok {
		return nil, errors.New("unknown tenancy")
	}
	return tenancy, nil
}

func (s *TenancyManager) TenancyByPath(path string) (multitenancy.Tenancy, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	tenancy, ok := s.tenanciesByPath[path]
	if !ok {
		return nil, errors.New("tenancy not found")
	}
	return tenancy, nil
}

func (s *TenancyManager) UnloadTenancy(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	tenancy, ok := s.tenanciesById[id]
	if ok {
		delete(s.tenanciesById, id)
		delete(s.tenanciesByPath, tenancy.Path())
	}
}

func (s *TenancyManager) LoadTenancy(ctx op_context.Context, id string) (multitenancy.Tenancy, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("TenancyManager.LoadTenancy", logger.Fields{"tenancy": id})
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// load from database
	tenancyDb, err := s.controller.Find(ctx, id)
	if err != nil {
		c.SetMessage("failed to find tenancy")
		return nil, err
	}
	if tenancyDb == nil {
		err := errors.New("tenancy not found")
		return nil, err
	}

	// init tenancy
	tenancy := NewTenancy()
	err = tenancy.Init(ctx, s.pools, tenancyDb)
	if err != nil {
		return nil, err
	}

	// keep it
	s.tenanciesById[tenancy.GetID()] = tenancy
	s.tenanciesByPath[tenancy.Path()] = tenancy

	// done
	return tenancy, nil
}

func (s *TenancyManager) CreateTenancy(ctx op_context.Context, data *multitenancy.TenancyData) (*multitenancy.TenancyDb, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("TenancyManager.LoadTenancy", logger.Fields{"customer": data.CUSTOMER, "role": data.ROLE})
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// check if customer exists
	owner, err := s.customers.Find(ctx, data.CUSTOMER)
	if err != nil {
		c.SetMessage("failed to find customer")
		return nil, err
	}
	if owner == nil {
		// try to find by login
		owner, err = s.customers.FindByLogin(ctx, data.CUSTOMER)
		if err != nil {
			c.SetMessage("failed to find customer")
			return nil, err
		}
	}
	if owner == nil {
		err = errors.New("customer not found")
		// TODO load customer errors in tenancies service
		ctx.SetGenericErrorCode(customer.ErrorCodeCustomerNotFound, true)
		return nil, err
	}

	// check if pool exists
	_, err = s.pools.Pool(data.POOL_ID)
	if err != nil {
		ctx.SetGenericErrorCode(pool.ErrorCodePoolNotFound)
		c.SetMessage("unknown pool")
		return nil, err
	}

	// create
	tenancy := &multitenancy.TenancyDb{}
	tenancy.InitObject()
	tenancy.TenancyData = *data
	tenancy.CUSTOMER = owner.GetID()
	if tenancy.PATH == "" {
		tenancy.PATH = crypt_utils.GenerateString()
	}
	if tenancy.DBNAME == "" {
		tenancy.DBNAME = utils.ConcatStrings(s.DB_PREFIX, "_", owner.Login(), "_", data.ROLE)
	}

	// done
	return tenancy, nil
}
