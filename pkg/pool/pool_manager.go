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

type PoolManager interface {
	AddPool(ctx op_context.Context, pool Pool) error
	FindPool(ctx op_context.Context, id string) (Pool, error)
	FindPoolByName(ctx op_context.Context, name string) (Pool, error)
	UpdatePool(ctx op_context.Context, poolName string, fields db.Fields) error
	GetPools(ctx op_context.Context, filter *db.Filter, pools interface{}) error
	GetPoolsBase(ctx op_context.Context, filter *db.Filter) ([]*PoolBase, error)

	AddService(ctx op_context.Context, service PoolService) error
	FindService(ctx op_context.Context, id string) (PoolService, error)
	FindServiceByName(ctx op_context.Context, name string) (PoolService, error)
	UpdateService(ctx op_context.Context, serviceName string, fields db.Fields) error
	GetServices(ctx op_context.Context, filter *db.Filter, poolServices interface{}) error
	GetServicesBase(ctx op_context.Context, filter *db.Filter) ([]*PoolServiceBase, error)

	AddBinding(ctx op_context.Context, binding PoolServiceBinding) error
	UpdateBinding(ctx op_context.Context, id string, fields db.Fields) error
	DeleteBinding(ctx op_context.Context, id string) error
	GetBindings(ctx op_context.Context, filter *db.Filter, bindings interface{}) error
	GetBindingsBase(ctx op_context.Context, filter *db.Filter) ([]*PoolServiceBindingBase, error)

	GetPoolBindings(ctx op_context.Context, poolId string, bindings interface{}) error
	GetPoolBindingsBase(ctx op_context.Context, poolId string) ([]*PoolServiceBindingBase, error)

	BindPoolService(ctx op_context.Context, poolName string, serviceName string, bindingType string, bindingName ...string) (PoolServiceBinding, error)
	FindPoolServiceBindingType(ctx op_context.Context, poolId string, bindingType string) (PoolService, error)
}

type PoolManagerBase struct {
	CRUD crud.CRUD
}

func (m *PoolManagerBase) AddPool(ctx op_context.Context, pool Pool) error {
	c := ctx.TraceInMethod("PoolManager.AddPool", logger.Fields{"pool_name": pool.Name()})
	defer ctx.TraceOutMethod()

	return c.SetError(ctx.DB().Create(ctx, pool))
}

func (m *PoolManagerBase) FindPool(ctx op_context.Context, id string) (Pool, error) {
	return crud.FindByField(m.CRUD, ctx, "PoolManager.FindPool", "id", id, &PoolBase{})
}

func (m *PoolManagerBase) FindPoolByName(ctx op_context.Context, name string) (Pool, error) {
	return crud.FindByField(m.CRUD, ctx, "PoolManager.FindPoolByName", "name", name, &PoolBase{})
}

func (m *PoolManagerBase) UpdatePool(ctx op_context.Context, poolName string, fields db.Fields) error {
	return crud.FindUpdate(m.CRUD, ctx, "PoolManager.UpdatePool", "name", poolName, fields, &PoolBase{}, logger.Fields{"pool_name": poolName})
}

func (m *PoolManagerBase) GetPools(ctx op_context.Context, filter *db.Filter, pools interface{}) error {
	return crud.List(m.CRUD, ctx, "PoolManager.GetPools", filter, pools)
}

func (m *PoolManagerBase) GetPoolsBase(ctx op_context.Context, filter *db.Filter) ([]*PoolBase, error) {
	var pools []*PoolBase
	err := crud.List(m.CRUD, ctx, "PoolManager.GetPoolsBase", filter, &pools)
	if err != nil {
		return nil, err
	}
	return pools, nil
}

func (m *PoolManagerBase) AddService(ctx op_context.Context, service PoolService) error {
	c := ctx.TraceInMethod("PoolManager.AddService", logger.Fields{"service_name": service.Name(), "service_type": service.Type()})
	defer ctx.TraceOutMethod()

	return c.SetError(ctx.DB().Create(ctx, service))
}

func (m *PoolManagerBase) FindService(ctx op_context.Context, id string) (PoolService, error) {
	return crud.FindByField(m.CRUD, ctx, "PoolManager.FindService", "id", id, &PoolServiceBase{})
}

func (m *PoolManagerBase) FindServiceByName(ctx op_context.Context, name string) (PoolService, error) {
	return crud.FindByField(m.CRUD, ctx, "PoolManager.FindServiceByName", "name", name, &PoolServiceBase{})
}

func (m *PoolManagerBase) UpdateService(ctx op_context.Context, name string, fields db.Fields) error {
	return crud.FindUpdate(m.CRUD, ctx, "PoolManager.UpdateService", "name", name, fields, &PoolServiceBase{}, logger.Fields{"service_name": name})
}

func (m *PoolManagerBase) GetServices(ctx op_context.Context, filter *db.Filter, services interface{}) error {
	return crud.List(m.CRUD, ctx, "PoolManager.GetServices", filter, services)
}

func (m *PoolManagerBase) GetServicesBase(ctx op_context.Context, filter *db.Filter) ([]*PoolServiceBase, error) {
	var services []*PoolServiceBase
	err := crud.List(m.CRUD, ctx, "PoolManager.GetServicesBase", filter, &services)
	if err != nil {
		return nil, err
	}
	return services, nil
}

func (m *PoolManagerBase) AddBinding(ctx op_context.Context, binding PoolServiceBinding) error {
	c := ctx.TraceInMethod("PoolManager.AddBinding", logger.Fields{"pool_id": binding.Pool(), "service_id": binding.Service(), "binding_type": binding.Type()})
	defer ctx.TraceOutMethod()

	return c.SetError(ctx.DB().Create(ctx, binding))
}

func (m *PoolManagerBase) GetPoolBindings(ctx op_context.Context, poolId string, bindings interface{}) error {
	filter := &db.Filter{}
	filter.Fields["pool_id"] = poolId
	return crud.List(m.CRUD, ctx, "PoolManager.GetPoolBindings", filter, bindings)
}

func (m *PoolManagerBase) GetPoolBindingsBase(ctx op_context.Context, poolId string) ([]*PoolServiceBindingBase, error) {
	c := ctx.TraceInMethod("PoolManager.GetPoolBindingsBase")
	defer ctx.TraceOutMethod()

	var bindings []*PoolServiceBindingBase
	err := m.GetPoolBindings(ctx, poolId, &bindings)
	if err != nil {
		return nil, c.SetError(err)
	}

	return bindings, nil
}

func (m *PoolManagerBase) DeleteBinding(ctx op_context.Context, id string) error {
	c := ctx.TraceInMethod("PoolManager.DeleteBindingBase", logger.Fields{"binding_id": id})
	defer ctx.TraceOutMethod()

	return c.SetError(ctx.DB().DeleteByField(ctx, "id", id, &PoolServiceBindingBase{}))
}

func (m *PoolManagerBase) GetBindings(ctx op_context.Context, filter *db.Filter, bindings interface{}) error {
	return crud.List(m.CRUD, ctx, "PoolManager.GetBindings", filter, bindings)
}

func (m *PoolManagerBase) GetBindingsBase(ctx op_context.Context, filter *db.Filter) ([]*PoolServiceBindingBase, error) {
	var bindings []*PoolServiceBindingBase
	err := crud.List(m.CRUD, ctx, "PoolManager.GetBindingsBase", filter, &bindings)
	if err != nil {
		return nil, err
	}
	return bindings, nil
}

func (m *PoolManagerBase) UpdateBinding(ctx op_context.Context, id string, fields db.Fields) error {
	return crud.FindUpdate(m.CRUD, ctx, "PoolManager.UpdateBinding", "id", id, fields, &PoolServiceBindingBase{}, logger.Fields{"binding_id": id})
}

func (m *PoolManagerBase) FindPoolServiceBindingType(ctx op_context.Context, poolId string, bindingType string) (PoolService, error) {

	c := ctx.TraceInMethod("PoolManager.FindPoolServiceBindingType", logger.Fields{"pool_id": poolId, "binding_type": bindingType})
	defer ctx.TraceOutMethod()

	fields := db.Fields{"pool_id": poolId, "type": bindingType}
	binding, err := crud.Find(m.CRUD, ctx, "PoolManager.FindPoolServiceBindingType", fields, &PoolServiceBindingBase{})
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

func (m *PoolManagerBase) BindPoolService(ctx op_context.Context, poolName string, serviceName string, bindingType string, bindingName ...string) (PoolServiceBinding, error) {

	c := ctx.TraceInMethod("PoolManager.BindPoolService", logger.Fields{"pool_name": poolName, "service_name": serviceName, "binding_type": bindingType})
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
