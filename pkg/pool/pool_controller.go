package pool

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const ErrorCodePoolNotFound = "pool_not_found"
const ErrorCodeServiceNotFound = "service_not_found"
const ErrorCodePoolServiceBindingsExist = "pool_service_bindings_exist"

type PoolController interface {
	AddPool(ctx op_context.Context, pool Pool) error
	FindPool(ctx op_context.Context, id string, idIsName ...bool) (Pool, error)
	UpdatePool(ctx op_context.Context, id string, fields db.Fields, idIsName ...bool) error
	DeletePool(ctx op_context.Context, id string, idIsName ...bool) error
	GetPools(ctx op_context.Context, filter *db.Filter) ([]*PoolBase, error)

	AddService(ctx op_context.Context, service PoolService) error
	FindService(ctx op_context.Context, id string, idIsName ...bool) (PoolService, error)
	UpdateService(ctx op_context.Context, id string, fields db.Fields, idIsName ...bool) error
	DeleteService(ctx op_context.Context, id string, idIsName ...bool) error
	GetServices(ctx op_context.Context, filter *db.Filter) ([]*PoolServiceBase, error)

	AddServiceToPool(ctx op_context.Context, poolId string, serviceId string, role string, idIsName ...bool) error
	RemoveServiceFromPool(ctx op_context.Context, poolId string, role string, idIsName ...bool) error
	RemoveAllServicesFromPool(ctx op_context.Context, poolId string, idIsName ...bool) error
	RemoveServiceFromAllPools(ctx op_context.Context, id string, idIsName ...bool) error

	GetPoolBindings(ctx op_context.Context, id string, idIsName ...bool) ([]*PoolServiceBindingBase, error)
	GetServiceBindings(ctx op_context.Context, id string, idIsName ...bool) ([]*PoolServiceBindingBase, error)
}

func NewPoolController(crud crud.CRUD) *PoolControllerBase {
	c := &PoolControllerBase{}
	c.CRUD = crud
	return c
}

type PoolControllerBase struct {
	CRUD crud.CRUD
}

func fieldName(idIsName ...bool) string {
	fieldName := "id"
	if utils.OptionalArg(false, idIsName...) {
		fieldName = "name"
	}
	return fieldName
}

func (m *PoolControllerBase) AddPool(ctx op_context.Context, pool Pool) error {
	return crud.Create(m.CRUD, ctx, "PoolController.AddPool", pool, logger.Fields{"pool_name": pool.Name()})
}

func (m *PoolControllerBase) FindPool(ctx op_context.Context, id string, idIsName ...bool) (Pool, error) {
	field := fieldName(idIsName...)
	pool, err := crud.FindByField(m.CRUD, ctx, "PoolController.FindPool", field, id, &PoolBase{}, logger.Fields{field: id})
	if err != nil {
		return nil, err
	}
	if pool == nil {
		ctx.SetGenericErrorCode(ErrorCodePoolNotFound)
	}
	return pool, nil
}

func (m *PoolControllerBase) UpdatePool(ctx op_context.Context, id string, fields db.Fields, idIsName ...bool) error {
	field := fieldName(idIsName...)
	return crud.FindUpdate(m.CRUD, ctx, "PoolController.UpdatePool", field, id, fields, &PoolBase{}, logger.Fields{field: id})
}

func (m *PoolControllerBase) DeletePool(ctx op_context.Context, id string, fields db.Fields, idIsName ...bool) error {
	c := ctx.TraceInMethod("PoolController.DeletePool")
	defer ctx.TraceOutMethod()

	poolId, err := m.PoolId(c, ctx, id, idIsName...)
	if err != nil {
		return err
	}
	if poolId == "" {
		ctx.SetGenericError(nil)
		return nil
	}

	filter := db.NewFilter()
	filter.AddField("pool_id", poolId)
	exists, err := m.CRUD.Exists(ctx, filter, &PoolServiceAssociacionBase{})
	if err != nil {
		c.SetMessage("failed to check associations")
		return c.SetError(err)
	}
	if exists {
		ctx.SetGenericErrorCode(ErrorCodePoolServiceBindingsExist)
		return c.SetError(errors.New("can not delete pool with services, remove all service bindings first"))
	}

	return crud.Delete(m.CRUD, ctx, "PoolController.DeletePool", "id", poolId, &PoolBase{}, logger.Fields{"id": id})
}

func (m *PoolControllerBase) AddService(ctx op_context.Context, service PoolService) error {
	return crud.Create(m.CRUD, ctx, "PoolController.AddService", service, logger.Fields{"name": service.Name()})
}

func (m *PoolControllerBase) FindService(ctx op_context.Context, id string, idIsName ...bool) (PoolService, error) {
	field := fieldName(idIsName...)
	service, err := crud.FindByField(m.CRUD, ctx, "PoolController.FindService", field, id, &PoolServiceBase{}, logger.Fields{field: id})
	if err != nil {
		return nil, err
	}
	if service == nil {
		ctx.SetGenericErrorCode(ErrorCodeServiceNotFound)
	}
	return service, nil
}

func (m *PoolControllerBase) UpdateService(ctx op_context.Context, id string, fields db.Fields, idIsName ...bool) error {
	field := fieldName(idIsName...)
	return crud.FindUpdate(m.CRUD, ctx, "PoolController.UpdateService", field, id, fields, &PoolServiceBase{}, logger.Fields{field: id})
}

func (m *PoolControllerBase) DeleteService(ctx op_context.Context, id string, fields db.Fields, idIsName ...bool) error {

	c := ctx.TraceInMethod("PoolController.DeleteService")
	defer ctx.TraceOutMethod()

	serviceId, err := m.ServiceId(c, ctx, id, idIsName...)
	if err != nil {
		return err
	}
	if serviceId == "" {
		ctx.SetGenericError(nil)
		return nil
	}

	filter := db.NewFilter()
	filter.AddField("service_id", serviceId)
	exists, err := m.CRUD.Exists(ctx, filter, &PoolServiceAssociacionBase{})
	if err != nil {
		c.SetMessage("failed to check associations")
		return c.SetError(err)
	}
	if exists {
		ctx.SetGenericErrorCode(ErrorCodePoolServiceBindingsExist)
		return c.SetError(errors.New("can not delete service bound to pools, remove all service bindings first"))
	}

	return crud.Delete(m.CRUD, ctx, "PoolController.DeleteService", "id", serviceId, &PoolBase{}, logger.Fields{"id": id})
}

func (p *PoolControllerBase) GetPools(ctx op_context.Context, filter *db.Filter) ([]*PoolBase, error) {
	var pools []*PoolBase
	err := crud.List(p.CRUD, ctx, "", filter, &pools)
	if err != nil {
		return nil, err
	}
	return pools, nil
}

func (p *PoolControllerBase) GetServices(ctx op_context.Context, filter *db.Filter) ([]*PoolServiceBase, error) {
	var services []*PoolServiceBase
	err := crud.List(p.CRUD, ctx, "PoolController.GetServices", filter, &services)
	if err != nil {
		return nil, err
	}
	return services, nil
}

func (m *PoolControllerBase) AddServiceToPool(ctx op_context.Context, poolId string, serviceId string, role string, idIsName ...bool) (PoolServiceBinding, error) {

	field := fieldName(idIsName...)

	c := ctx.TraceInMethod("PoolController.AddServiceToPool", logger.Fields{utils.ConcatStrings("pool_", field): poolId, utils.ConcatStrings("service_", field): serviceId, "role": role})
	defer ctx.TraceOutMethod()

	pool, err := m.FindPool(ctx, poolId, idIsName...)
	if err != nil {
		c.SetMessage("failed to find pool")
		return nil, c.SetError(err)
	}
	if pool == nil {
		return nil, c.SetError(errors.New("pool not found"))
	}

	service, err := m.FindService(ctx, serviceId, idIsName...)
	if err != nil {
		c.SetMessage("failed to find service")
		return nil, c.SetError(err)
	}
	if service == nil {
		return nil, c.SetError(errors.New("unknown service"))
	}

	binding := &PoolServiceBindingBase{}
	binding.InitObject()
	binding.POOL_ID = pool.GetID()
	binding.SERVICE_ID = service.GetID()
	binding.ROLE = role
	err = m.CRUD.Create(ctx, binding)
	if err != nil {
		c.SetMessage("failed to save binding in database")
		return nil, c.SetError(err)
	}

	return binding, nil
}

func (m *PoolControllerBase) PoolId(c op_context.CallContext, ctx op_context.Context, id string, idIsName ...bool) (string, error) {

	if !utils.OptionalArg(false, idIsName...) {
		c.SetLoggerField("id", id)
		return id, nil
	}

	c.SetLoggerField("name", id)
	pool, err := m.FindPool(ctx, id, true)
	if err != nil {
		c.SetMessage("failed to find pool")
		return "", c.SetError(err)
	}
	if pool == nil {
		return "", nil
	}
	pId := pool.GetID()
	c.SetLoggerField("id", pId)
	return pId, nil
}

func (m *PoolControllerBase) ServiceId(c op_context.CallContext, ctx op_context.Context, id string, idIsName ...bool) (string, error) {

	if !utils.OptionalArg(false, idIsName...) {
		c.SetLoggerField("id", id)
		return id, nil
	}

	c.SetLoggerField("name", id)
	service, err := m.FindService(ctx, id, true)
	if err != nil {
		c.SetMessage("failed to find service")
		return "", c.SetError(err)
	}
	if service == nil {
		return "", nil
	}
	pId := service.GetID()
	c.SetLoggerField("id", pId)
	return pId, nil
}

func (m *PoolControllerBase) RemoveServiceFromPool(ctx op_context.Context, poolId string, role string, idIsName ...bool) error {

	c := ctx.TraceInMethod("PoolController.RemoveServiceFromPool")
	defer ctx.TraceOutMethod()

	pId, err := m.PoolId(c, ctx, poolId, idIsName...)
	if err != nil {
		return err
	}
	if pId == "" {
		ctx.SetGenericError(nil)
		return nil
	}

	association := &PoolServiceAssociacionBase{}
	fields := db.Fields{"pool_id": pId, "role": role}
	found, err := m.CRUD.Read(ctx, fields, association)
	if err != nil {
		c.SetMessage("failed to find association")
		return c.SetError(err)
	}
	if !found {
		return nil
	}

	err = m.CRUD.Delete(ctx, association)
	if err != nil {
		c.SetMessage("failed to delete association")
		return c.SetError(err)
	}

	return nil
}

func (m *PoolControllerBase) RemoveAllServicesFromPool(ctx op_context.Context, id string, idIsName ...bool) error {

	c := ctx.TraceInMethod("PoolController.RemoveAllServicesFromPool")
	defer ctx.TraceOutMethod()

	poolId, err := m.PoolId(c, ctx, id, idIsName...)
	if err != nil {
		return err
	}
	if poolId == "" {
		ctx.SetGenericError(nil)
		return nil
	}

	fields := db.Fields{"pool_id": poolId}
	err = m.CRUD.DeleteByFields(ctx, fields, &PoolServiceBindingBase{})
	if err != nil {
		return err
	}
	return nil
}

func (m *PoolControllerBase) RemoveServiceFromAllPools(ctx op_context.Context, id string, idIsName ...bool) error {

	c := ctx.TraceInMethod("PoolController.RemoveServiceFromAllPools")
	defer ctx.TraceOutMethod()

	serviceId, err := m.PoolId(c, ctx, id, idIsName...)
	if err != nil {
		return err
	}
	if serviceId == "" {
		ctx.SetGenericError(nil)
		return nil
	}

	fields := db.Fields{"service_id": serviceId}
	err = m.CRUD.DeleteByFields(ctx, fields, &PoolServiceBindingBase{})
	if err != nil {
		return err
	}
	return nil
}
