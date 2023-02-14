package pool

import (
	"errors"
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const ErrorCodePoolNotFound = "pool_not_found"
const ErrorCodeServiceNotFound = "service_not_found"
const ErrorCodePoolNameConflict = "pool_name_conflict"
const ErrorCodeServiceNameConflict = "service_name_conflict"
const ErrorCodeServiceRoleConflict = "service_role_conflict"
const ErrorCodePoolServiceBindingsExist = "pool_service_bindings_exist"

var ErrorDescriptions = map[string]string{
	ErrorCodePoolNotFound:             "Pool not found.",
	ErrorCodeServiceNotFound:          "Service not found.",
	ErrorCodePoolNameConflict:         "Pool with such name already exists, choose another name.",
	ErrorCodeServiceNameConflict:      "Service with such name already exists, choose another name.",
	ErrorCodeServiceRoleConflict:      "Pool already has service for that role.",
	ErrorCodePoolServiceBindingsExist: "Can't delete pool with services. First, remove all services from the pool.",
}

var ErrorHttpCodes = map[string]int{
	ErrorCodePoolNotFound:    http.StatusNotFound,
	ErrorCodeServiceNotFound: http.StatusNotFound,
}

type PoolController interface {
	AddPool(ctx op_context.Context, pool Pool) (Pool, error)
	FindPool(ctx op_context.Context, id string, idIsName ...bool) (Pool, error)
	UpdatePool(ctx op_context.Context, id string, fields db.Fields, idIsName ...bool) error
	DeletePool(ctx op_context.Context, id string, idIsName ...bool) error
	GetPools(ctx op_context.Context, filter *db.Filter) ([]*PoolBase, int64, error)

	AddService(ctx op_context.Context, service PoolService) (PoolService, error)
	FindService(ctx op_context.Context, id string, idIsName ...bool) (PoolService, error)
	UpdateService(ctx op_context.Context, id string, fields db.Fields, idIsName ...bool) error
	DeleteService(ctx op_context.Context, id string, idIsName ...bool) error
	GetServices(ctx op_context.Context, filter *db.Filter) ([]*PoolServiceBase, int64, error)

	AddServiceToPool(ctx op_context.Context, poolId string, serviceId string, role string, idIsName ...bool) (PoolServiceBinding, error)
	RemoveServiceFromPool(ctx op_context.Context, poolId string, role string, idIsName ...bool) error
	RemoveAllServicesFromPool(ctx op_context.Context, poolId string, idIsName ...bool) error
	RemoveServiceFromAllPools(ctx op_context.Context, id string, idIsName ...bool) error

	// GetPoolBindings(ctx op_context.Context, id string, idIsName ...bool) ([]*PoolServiceBindingBase, error)
	// GetServiceBindings(ctx op_context.Context, id string, idIsName ...bool) ([]*PoolServiceBindingBase, error)
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

func (m *PoolControllerBase) OpLog(ctx op_context.Context, operation string, oplog *OpLogPool) {
	oplog.SetOperation(operation)
	ctx.Oplog(oplog)
}

func (m *PoolControllerBase) AddPool(ctx op_context.Context, pool Pool) (Pool, error) {
	pool.InitObject()
	err := crud.Create(m.CRUD, ctx, "PoolController.AddPool", pool, logger.Fields{"pool_name": pool.Name()})
	if err != nil {
		return nil, err
	}
	m.OpLog(ctx, "add_pool", &OpLogPool{PoolId: pool.GetID(), PoolName: pool.Name()})
	return pool, nil
}

func (m *PoolControllerBase) FindPool(ctx op_context.Context, id string, idIsName ...bool) (Pool, error) {
	field := fieldName(idIsName...)
	pool, err := crud.FindByField(m.CRUD, ctx, "PoolController.FindPool", field, id, &PoolBase{})
	if err != nil {
		return nil, err
	}
	if pool == nil {
		ctx.SetGenericErrorCode(ErrorCodePoolNotFound)
	}
	return pool, nil
}

func (m *PoolControllerBase) UpdatePool(ctx op_context.Context, id string, fields db.Fields, idIsName ...bool) error {

	c := ctx.TraceInMethod("PoolController.UpdatePool", logger.Fields{"pool": id, "user_name": utils.OptionalArg(false, idIsName...)})
	defer ctx.TraceOutMethod()

	// check if name is unique
	if name, found := fields["name"]; found {
		filter := db.NewFilter()
		filter.AddField("name", name)
		exists, err := m.CRUD.Exists(ctx, filter, &PoolBase{})
		if err != nil {
			c.SetMessage("failed to check existence of pool with desired name")
			return c.SetError(err)
		}
		if exists {
			ctx.SetGenericErrorCode(ErrorCodePoolNameConflict)
			return c.SetError(errors.New("pool with desired name exists"))
		}
	}

	// update
	field := fieldName(idIsName...)
	obj, err := crud.FindUpdate(m.CRUD, ctx, "PoolController.FindUpdatePool", field, id, fields, &PoolBase{}, logger.Fields{field: id})
	if err != nil {
		return err
	}
	if obj == nil {
		ctx.SetGenericErrorCode(ErrorCodePoolNotFound)
		return c.SetError(errors.New("pool not found"))
	}
	m.OpLog(ctx, "update_pool", &OpLogPool{PoolId: obj.GetID(), PoolName: obj.Name()})
	return nil
}

func (m *PoolControllerBase) DeletePool(ctx op_context.Context, id string, idIsName ...bool) error {
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

	err = crud.Delete(m.CRUD, ctx, "PoolController.DeletePool", "id", poolId, &PoolBase{}, logger.Fields{"id": id})
	if err != nil {
		return err
	}

	o := &OpLogPool{PoolId: poolId}
	if utils.OptionalArg(false, idIsName...) {
		o.PoolName = id
	}
	m.OpLog(ctx, "delete_pool", o)
	return nil
}

func (m *PoolControllerBase) AddService(ctx op_context.Context, service PoolService) (PoolService, error) {
	service.InitObject()
	err := crud.Create(m.CRUD, ctx, "PoolController.AddService", service, logger.Fields{"name": service.Name()})
	if err != nil {
		return nil, err
	}
	m.OpLog(ctx, "add_service", &OpLogPool{ServiceId: service.GetID(), ServiceName: service.Name()})
	return service, nil
}

func (m *PoolControllerBase) FindService(ctx op_context.Context, id string, idIsName ...bool) (PoolService, error) {
	field := fieldName(idIsName...)
	service, err := crud.FindByField(m.CRUD, ctx, "PoolController.FindService", field, id, &PoolServiceBase{})
	if err != nil {
		return nil, err
	}
	if service == nil {
		ctx.SetGenericErrorCode(ErrorCodeServiceNotFound)
	}
	return service, nil
}

func (m *PoolControllerBase) UpdateService(ctx op_context.Context, id string, fields db.Fields, idIsName ...bool) error {

	c := ctx.TraceInMethod("PoolController.UpdateService", logger.Fields{"service": id, "use_name": utils.OptionalArg(false, idIsName...)})
	defer ctx.TraceOutMethod()

	// check if name is unique
	if name, found := fields["name"]; found {
		filter := db.NewFilter()
		filter.AddField("name", name)
		exists, err := m.CRUD.Exists(ctx, filter, &PoolServiceBase{})
		if err != nil {
			return c.SetError(err)
		}
		if exists {
			ctx.SetGenericErrorCode(ErrorCodeServiceNameConflict)
			return c.SetError(errors.New("service with such name exists"))
		}
	}

	// update
	field := fieldName(idIsName...)
	obj, err := crud.FindUpdate(m.CRUD, ctx, "PoolController.FindUpdateService", field, id, fields, &PoolServiceBase{}, logger.Fields{field: id})
	if err != nil {
		return c.SetError(err)
	}
	if obj == nil {
		ctx.SetGenericErrorCode(ErrorCodeServiceNotFound)
		return c.SetError(errors.New("service not found"))
	}
	m.OpLog(ctx, "update_service", &OpLogPool{ServiceId: obj.GetID(), ServiceName: obj.Name()})
	return nil
}

func (m *PoolControllerBase) DeleteService(ctx op_context.Context, id string, idIsName ...bool) error {

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

	err = crud.Delete(m.CRUD, ctx, "PoolController.DeleteService", "id", serviceId, &PoolBase{}, logger.Fields{"id": id})
	if err != nil {
		return err
	}

	o := &OpLogPool{ServiceId: serviceId}
	if utils.OptionalArg(false, idIsName...) {
		o.ServiceName = id
	}
	m.OpLog(ctx, "delete_service", o)
	return nil
}

func (p *PoolControllerBase) GetPools(ctx op_context.Context, filter *db.Filter) ([]*PoolBase, int64, error) {
	var pools []*PoolBase
	count, err := crud.List(p.CRUD, ctx, "", filter, &pools)
	if err != nil {
		return nil, 0, err
	}
	return pools, count, nil
}

func (p *PoolControllerBase) GetServices(ctx op_context.Context, filter *db.Filter) ([]*PoolServiceBase, int64, error) {
	var services []*PoolServiceBase
	count, err := crud.List(p.CRUD, ctx, "PoolController.GetServices", filter, &services)
	if err != nil {
		return nil, 0, err
	}
	return services, count, nil
}

func (m *PoolControllerBase) AddServiceToPool(ctx op_context.Context, poolId string, serviceId string, role string, idIsName ...bool) (PoolServiceBinding, error) {

	field := fieldName(idIsName...)

	c := ctx.TraceInMethod("PoolController.AddServiceToPool", logger.Fields{utils.ConcatStrings("pool_", field): poolId, utils.ConcatStrings("service_", field): serviceId, "role": role})
	defer ctx.TraceOutMethod()

	// find pool
	pool, err := m.FindPool(ctx, poolId, idIsName...)
	if err != nil {
		c.SetMessage("failed to find pool")
		return nil, c.SetError(err)
	}
	if pool == nil {
		return nil, c.SetError(errors.New("pool not found"))
	}

	// find service
	service, err := m.FindService(ctx, serviceId, idIsName...)
	if err != nil {
		c.SetMessage("failed to find service")
		return nil, c.SetError(err)
	}
	if service == nil {
		return nil, c.SetError(errors.New("unknown service"))
	}

	// check if name is unique
	binding := &PoolServiceBindingBase{}
	filter := db.NewFilter()
	filter.AddField("pool_id", pool.GetID())
	filter.AddField("role", role)
	exists, err := m.CRUD.Exists(ctx, filter, &PoolServiceBindingBase{})
	if err != nil {
		c.SetMessage("failed to check existence of pool service binding")
		return nil, c.SetError(err)
	}
	if exists {
		ctx.SetGenericErrorCode(ErrorCodeServiceRoleConflict)
		return nil, c.SetError(errors.New("pool already has service with such role"))
	}

	// create binding
	binding.InitObject()
	binding.POOL_ID = pool.GetID()
	binding.SERVICE_ID = service.GetID()
	binding.ROLE = role
	err = m.CRUD.Create(ctx, binding)
	if err != nil {
		c.SetMessage("failed to save binding in database")
		return nil, c.SetError(err)
	}

	// add oplog
	m.OpLog(ctx, "add_service_to_pool", &OpLogPool{ServiceId: service.GetID(), ServiceName: service.Name(),
		PoolId: pool.GetID(), PoolName: pool.Name(),
		Role: role,
	})

	// done
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

func (m *PoolControllerBase) RemoveServiceFromPool(ctx op_context.Context, id string, role string, idIsName ...bool) error {

	c := ctx.TraceInMethod("PoolController.RemoveServiceFromPool", logger.Fields{"role": role})
	defer ctx.TraceOutMethod()

	poolId, err := m.PoolId(c, ctx, id, idIsName...)
	if err != nil {
		return err
	}
	if poolId == "" {
		ctx.SetGenericError(nil)
		return nil
	}

	association := &PoolServiceAssociacionBase{}
	fields := db.Fields{"pool_id": poolId, "role": role}
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

	o := &OpLogPool{PoolId: poolId, Role: role}
	if utils.OptionalArg(false, idIsName...) {
		o.PoolName = id
	}
	m.OpLog(ctx, "remove_service_from_pool", o)
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

	o := &OpLogPool{PoolId: poolId}
	if utils.OptionalArg(false, idIsName...) {
		o.PoolName = id
	}
	m.OpLog(ctx, "remove_all_services_from_pool", o)
	return nil
}

func (m *PoolControllerBase) RemoveServiceFromAllPools(ctx op_context.Context, id string, idIsName ...bool) error {

	c := ctx.TraceInMethod("PoolController.RemoveServiceFromAllPools")
	defer ctx.TraceOutMethod()

	serviceId, err := m.ServiceId(c, ctx, id, idIsName...)
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

	o := &OpLogPool{ServiceId: serviceId}
	if utils.OptionalArg(false, idIsName...) {
		o.ServiceName = id
	}
	m.OpLog(ctx, "remove_service_from_all_pools", o)
	return nil
}
