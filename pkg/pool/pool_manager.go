package pool

import (
	"errors"
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type PoolController interface {
	AddPool(ctx op_context.Context, pool Pool) error
	FindPool(ctx op_context.Context, id string) (Pool, error)
	FindPoolByName(ctx op_context.Context, name string) (Pool, error)
	UpdatePool(ctx op_context.Context, poolName string, fields db.Fields) error

	AddService(ctx op_context.Context, service PoolService) error
	FindService(ctx op_context.Context, id string) (PoolService, error)
	FindServiceByName(ctx op_context.Context, name string) (PoolService, error)
	UpdateService(ctx op_context.Context, serviceName string, fields db.Fields) error

	AddBinding(ctx op_context.Context, binding PoolServiceBinding) error
	UpdateBinding(ctx op_context.Context, id string, fields db.Fields) error
	DeleteBinding(ctx op_context.Context, id string) error

	BindPoolService(ctx op_context.Context, poolName string, serviceName string, bindingType string, bindingName ...string) (PoolServiceBinding, error)
	FindPoolServiceBindingType(ctx op_context.Context, poolId string, bindingType string) (PoolService, error)
}

type PoolCollection[PoolType Pool] interface {
	GetPools(ctx op_context.Context, filter *db.Filter, pools *[]PoolType) error
}

type PoolServicesCollection[ServiceType PoolService] interface {
	GetServices(ctx op_context.Context, filter *db.Filter, poolServices *[]ServiceType) error
}

type PoolServiceBindingCollection[BindingType PoolServiceBinding] interface {
	GetBindings(ctx op_context.Context, filter *db.Filter, bindings *[]BindingType) error
	GetPoolBindings(ctx op_context.Context, poolId string, bindings *[]BindingType) error
}

type CompoundPoolManager[PoolType Pool, ServiceType PoolService, BindingType PoolServiceBinding] interface {
	PoolController
	PoolCollection[PoolType]
	PoolServicesCollection[ServiceType]
	PoolServiceBindingCollection[BindingType]
}

type PoolManager = CompoundPoolManager[Pool, PoolService, PoolServiceBinding]
type PoolManagerBase = CompoundPoolManagerBase[*PoolBase, *PoolServiceBase, *PoolServiceBindingBase]

type CompoundPoolManagerBase[PoolType Pool, ServiceType PoolService, BindingType PoolServiceBinding] struct {
	PoolControllerBase
	PoolCollectionBase[PoolType]
	PoolServicesCollectionBase[ServiceType]
	PoolServiceBindingCollectionBase[BindingType]
}

func NewPoolManager(crud crud.CRUD) *PoolManagerBase {
	m := &PoolManagerBase{}
	m.PoolControllerBase.CRUD = crud
	m.PoolCollectionBase.Manager = &m.PoolControllerBase
	m.PoolServicesCollectionBase.Manager = &m.PoolControllerBase
	m.PoolServiceBindingCollectionBase.Manager = &m.PoolControllerBase
	return m
}

type PoolControllerBase struct {
	CRUD crud.CRUD
}

func (m *PoolControllerBase) AddPool(ctx op_context.Context, pool Pool) error {
	c := ctx.TraceInMethod("PoolController.AddPool", logger.Fields{"pool_name": pool.Name()})
	defer ctx.TraceOutMethod()

	return c.SetError(ctx.DB().Create(ctx, pool))
}

func (m *PoolControllerBase) FindPool(ctx op_context.Context, id string) (Pool, error) {
	return crud.FindByField(m.CRUD, ctx, "PoolController.FindPool", "id", id, &PoolBase{})
}

func (m *PoolControllerBase) FindPoolByName(ctx op_context.Context, name string) (Pool, error) {
	return crud.FindByField(m.CRUD, ctx, "PoolController.FindPoolByName", "name", name, &PoolBase{})
}

func (m *PoolControllerBase) UpdatePool(ctx op_context.Context, poolName string, fields db.Fields) error {
	return crud.FindUpdate(m.CRUD, ctx, "PoolController.UpdatePool", "name", poolName, fields, &PoolBase{}, logger.Fields{"pool_name": poolName})
}

func (m *PoolControllerBase) AddService(ctx op_context.Context, service PoolService) error {
	c := ctx.TraceInMethod("PoolController.AddService", logger.Fields{"service_name": service.Name(), "service_type": service.Type()})
	defer ctx.TraceOutMethod()

	return c.SetError(ctx.DB().Create(ctx, service))
}

func (m *PoolControllerBase) FindService(ctx op_context.Context, id string) (PoolService, error) {
	return crud.FindByField(m.CRUD, ctx, "PoolController.FindService", "id", id, &PoolServiceBase{})
}

func (m *PoolControllerBase) FindServiceByName(ctx op_context.Context, name string) (PoolService, error) {
	return crud.FindByField(m.CRUD, ctx, "PoolController.FindServiceByName", "name", name, &PoolServiceBase{})
}

func (m *PoolControllerBase) UpdateService(ctx op_context.Context, name string, fields db.Fields) error {
	return crud.FindUpdate(m.CRUD, ctx, "PoolController.UpdateService", "name", name, fields, &PoolServiceBase{}, logger.Fields{"service_name": name})
}

func (m *PoolControllerBase) AddBinding(ctx op_context.Context, binding PoolServiceBinding) error {
	c := ctx.TraceInMethod("PoolController.AddBinding", logger.Fields{"pool_id": binding.Pool(), "service_id": binding.Service(), "binding_type": binding.Type()})
	defer ctx.TraceOutMethod()

	return c.SetError(ctx.DB().Create(ctx, binding))
}

func (m *PoolControllerBase) DeleteBinding(ctx op_context.Context, id string) error {
	c := ctx.TraceInMethod("PoolController.DeleteBindingBase", logger.Fields{"binding_id": id})
	defer ctx.TraceOutMethod()

	return c.SetError(ctx.DB().DeleteByField(ctx, "id", id, &PoolServiceBindingBase{}))
}

func (m *PoolControllerBase) UpdateBinding(ctx op_context.Context, id string, fields db.Fields) error {
	return crud.FindUpdate(m.CRUD, ctx, "PoolController.UpdateBinding", "id", id, fields, &PoolServiceBindingBase{}, logger.Fields{"binding_id": id})
}

func (m *PoolControllerBase) FindPoolServiceBindingType(ctx op_context.Context, poolId string, bindingType string) (PoolService, error) {

	c := ctx.TraceInMethod("PoolController.FindPoolServiceBindingType", logger.Fields{"pool_id": poolId, "binding_type": bindingType})
	defer ctx.TraceOutMethod()

	fields := db.Fields{"pool_id": poolId, "type": bindingType}
	binding, err := crud.Find(m.CRUD, ctx, "PoolController.FindPoolServiceBindingType", fields, &PoolServiceBindingBase{})
	if err != nil {
		c.SetMessage("failed to find binding")
		return nil, c.SetError(err)
	}
	if binding == nil {
		return nil, c.SetError(fmt.Errorf("unknown binding"))
	}
	c.SetLoggerField("service_id", binding.Service())

	service := &PoolServiceBase{}
	found, err := ctx.DB().FindByField(ctx, "id", binding.Service(), service)
	if err != nil {
		c.SetMessage("failed to find pool service")
		return nil, c.SetError(err)
	}
	if !found {
		return nil, c.SetError(fmt.Errorf("unknown pool service"))
	}

	return service, nil
}

func (m *PoolControllerBase) BindPoolService(ctx op_context.Context, poolName string, serviceName string, bindingType string, bindingName ...string) (PoolServiceBinding, error) {

	c := ctx.TraceInMethod("PoolController.BindPoolService", logger.Fields{"pool_name": poolName, "service_name": serviceName, "binding_type": bindingType})
	bName := utils.OptionalArg("", bindingName...)
	if bName != "" {
		c.SetLoggerField("binding_name", bName)
	}
	defer ctx.TraceOutMethod()

	pool, err := m.FindPoolByName(ctx, poolName)
	if err != nil {
		c.SetMessage("failed to find pool")
		return nil, c.SetError(err)
	}
	if pool == NilPool {
		return nil, c.SetError(errors.New("unknown pool"))
	}

	service, err := m.FindServiceByName(ctx, serviceName)
	if err != nil {
		c.SetMessage("failed to find service")
		return nil, c.SetError(err)
	}
	if service == NilService {
		return nil, c.SetError(errors.New("unknown service"))
	}

	binding := &PoolServiceBindingBase{}
	binding.InitObject()
	binding.SetName(bName)
	binding.POOL_ID = pool.GetID()
	binding.SERVICE_ID = service.GetID()
	binding.TYPE = bindingType
	err = ctx.DB().Create(ctx, binding)
	if err != nil {
		c.SetMessage("failed to save binding in database")
		return nil, c.SetError(err)
	}

	return binding, nil
}

type PoolCollectionBase[PoolType Pool] struct {
	Manager *PoolControllerBase
}

func (p *PoolCollectionBase[PoolType]) GetPools(ctx op_context.Context, filter *db.Filter, pools *[]PoolType) error {

	c := ctx.TraceInMethod("PoolCollection.GetPools")
	defer ctx.TraceOutMethod()

	err := p.Manager.CRUD.List(ctx, filter, pools)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

type PoolServicesCollectionBase[ServiceType PoolService] struct {
	Manager *PoolControllerBase
}

func (p *PoolServicesCollectionBase[ServiceType]) GetServices(ctx op_context.Context, filter *db.Filter, poolServices *[]ServiceType) error {

	c := ctx.TraceInMethod("PoolServicesCollection.GetServices")
	defer ctx.TraceOutMethod()

	err := p.Manager.CRUD.List(ctx, filter, poolServices)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

type PoolServiceBindingCollectionBase[BindingType PoolServiceBinding] struct {
	Manager *PoolControllerBase
}

func (p *PoolServiceBindingCollectionBase[BindingType]) GetBindings(ctx op_context.Context, filter *db.Filter, bindings *[]BindingType) error {

	c := ctx.TraceInMethod("PoolServiceBindingCollection.GetBindings")
	defer ctx.TraceOutMethod()

	err := p.Manager.CRUD.List(ctx, filter, bindings)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func (p *PoolServiceBindingCollectionBase[BindingType]) GetPoolBindings(ctx op_context.Context, poolId string, bindings *[]BindingType) error {
	filter := db.NewFilter()
	filter.AddField("pool_id", poolId)

	c := ctx.TraceInMethod("PoolServiceBindingCollection.GetPoolBindings")
	defer ctx.TraceOutMethod()

	err := p.Manager.CRUD.List(ctx, filter, bindings)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}
