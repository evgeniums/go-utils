package tenancy_manager

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type TenancyController struct {
	CRUD    crud.CRUD
	Manager *TenancyManager
}

func NewTenancyController(crud crud.CRUD, manager *TenancyManager) *TenancyController {
	c := &TenancyController{}
	c.CRUD = crud
	c.Manager = manager
	c.Manager.SetController(c)
	return c
}

func (t *TenancyController) OpLog(ctx op_context.Context, operation string, oplog *multitenancy.OpLogTenancy) {
	oplog.SetOperation(operation)
	ctx.Oplog(oplog)
}

func TenancyId(ctrl multitenancy.TenancyController, ctx op_context.Context, id string, idIsDisplay ...bool) (string, *multitenancy.TenancyItem, error) {

	useDisplay := utils.OptionalArg(false, idIsDisplay...)

	// setup
	c := ctx.TraceInMethod("TenancyController.TenancyId", logger.Fields{"tenancy": id, "use_display": useDisplay})
	defer ctx.TraceOutMethod()

	// return ID as is if it is not display format
	if !useDisplay {
		return id, nil, nil
	}

	// parse id
	customerLogin, role, vErr := multitenancy.ParseTenancyDisplay(id)
	if vErr != nil {
		c.SetMessage("failed to parse display")
		ctx.SetGenericError(vErr.GenericError())
		return "", nil, c.SetError(vErr)
	}

	// find tenancy by login and role
	filter := db.NewFilter()
	filter.AddField("customer_login", customerLogin)
	filter.AddField("role", role)
	filter.Limit = 1
	tenancies, _, err := ctrl.List(ctx, filter)
	if err != nil {
		c.SetMessage("failed to find tenancy")
		return "", nil, c.SetError(err)
	}
	if len(tenancies) == 0 {
		ctx.SetGenericErrorCode(multitenancy.ErrorCodeTenancyNotFound)
		return "", nil, c.SetError(errors.New("tenancy not found"))
	}
	tenancy := tenancies[0]

	// done
	return tenancy.GetID(), tenancy, nil
}

func FindTenancy(ctrl multitenancy.TenancyController, ctx op_context.Context, id string, idIsDisplay ...bool) (*multitenancy.TenancyItem, error) {

	// setup
	c := ctx.TraceInMethod("TenancyController.Find")
	defer ctx.TraceOutMethod()

	// adjust ID
	id, tenancy, err := TenancyId(ctrl, ctx, id, idIsDisplay...)
	if err != nil {
		return nil, c.SetError(err)
	}

	// maybe done
	if tenancy != nil {
		return tenancy, nil
	}

	// find tenancy
	filter := db.NewFilter()
	filter.AddField("tenancies.id", id)
	filter.Limit = 1
	tenancies, _, err := ctrl.List(ctx, filter)
	if err != nil {
		c.SetMessage("failed to find tenancy")
		return nil, c.SetError(err)
	}
	if len(tenancies) == 0 {
		ctx.SetGenericErrorCode(multitenancy.ErrorCodeTenancyNotFound)
		return nil, c.SetError(errors.New("tenancy not found"))
	}
	tenancy = tenancies[0]

	// done
	return tenancy, nil
}

func (t *TenancyController) PublishOp(tenancy *multitenancy.TenancyItem, op string) {
	t.Manager.PoolPubsub.PublishPools(multitenancy.PubsubTopicName,
		&multitenancy.PubsubNotification{Tenancy: tenancy.GetID(), Operation: op},
		tenancy.PoolId())
}

func (t *TenancyController) Add(ctx op_context.Context, data *multitenancy.TenancyData) (*multitenancy.TenancyItem, error) {

	// setup
	c := ctx.TraceInMethod("TenancyController.Add", logger.Fields{"customer": data.CUSTOMER_ID, "role": data.ROLE})
	defer ctx.TraceOutMethod()

	// create tenancy
	tenancy, err := t.Manager.CreateTenancy(ctx, data)
	if err != nil {
		c.SetMessage("failed to create tenancy")
		return nil, c.SetError(err)
	}

	// save tenancy in database
	err = t.CRUD.Create(ctx, tenancy)
	if err != nil {
		c.SetMessage("failed to save tenancy in database")
		return nil, c.SetError(err)
	}

	// save oplog
	t.OpLog(ctx, multitenancy.OpAdd, &multitenancy.OpLogTenancy{TenancyId: tenancy.GetID(),
		Role: tenancy.Role(), Path: tenancy.Path(), DbName: tenancy.DbName(), Pool: tenancy.PoolName, Customer: tenancy.CustomerDisplay()})

	// publish notification
	t.PublishOp(tenancy, multitenancy.OpAdd)

	// done
	return tenancy, nil
}

func (t *TenancyController) Find(ctx op_context.Context, id string, idIsDisplay ...bool) (*multitenancy.TenancyItem, error) {
	return FindTenancy(t, ctx, id, idIsDisplay...)
}

func (t *TenancyController) List(ctx op_context.Context, filter *db.Filter) ([]*multitenancy.TenancyItem, int64, error) {

	// setup
	c := ctx.TraceInMethod("TenancyController.List")
	defer ctx.TraceOutMethod()

	// construct join query
	queryBuilder := func() (db.JoinQuery, error) {
		return ctx.Db().Joiner().
			Join(&multitenancy.TenancyDb{}, "customer_id").On(&customer.Customer{}, "id").
			Join(&multitenancy.TenancyDb{}, "pool_id").On(&pool.PoolBase{}, "id").
			Destination(&multitenancy.TenancyItem{})
	}

	// invoke join
	var tenancies []*multitenancy.TenancyItem
	count, err := t.CRUD.Join(ctx, db.NewJoin(queryBuilder, "ListTenancies"), filter, &tenancies)
	if err != nil {
		c.SetMessage("failed to find tenancy")
		return nil, 0, c.SetError(err)
	}

	// done
	return tenancies, count, nil
}

func (t *TenancyController) Exists(ctx op_context.Context, fields db.Fields) (bool, error) {
	filter := db.NewFilter()
	filter.AddFields(fields)
	return crud.Exists(t.CRUD, ctx, "TenancyController.Exists", filter, &multitenancy.TenancyDb{})
}

func (t *TenancyController) SetPath(ctx op_context.Context, id string, path string, idIsDisplay ...bool) error {

	// setup
	c := ctx.TraceInMethod("TenancyController.SetPath")
	defer ctx.TraceOutMethod()

	// find tenancy
	tenancy, err := t.Find(ctx, id, idIsDisplay...)
	if err != nil {
		return c.SetError(err)
	}

	// check if tenancy with such path in that pool
	err = t.Manager.CheckDuplicatePath(ctx, c, tenancy.PoolId(), path)
	if err != nil {
		return err
	}

	// update field
	err = t.CRUD.Update(ctx, &tenancy.TenancyDb, db.Fields{"path": path})
	if err != nil {
		c.SetMessage("failed to update tenancy")
		return c.SetError(err)
	}

	// save oplog
	t.OpLog(ctx, multitenancy.OpSetPath, &multitenancy.OpLogTenancy{TenancyId: tenancy.GetID(),
		Role: tenancy.Role(), Path: tenancy.Path(), Customer: tenancy.CustomerDisplay()})

	// publish notification
	t.PublishOp(tenancy, multitenancy.OpSetPath)

	// done
	return nil
}

func (t *TenancyController) SetRole(ctx op_context.Context, id string, role string, idIsDisplay ...bool) error {

	// setup
	c := ctx.TraceInMethod("TenancyController.SetRole")
	defer ctx.TraceOutMethod()

	// find tenancy
	tenancy, err := t.Find(ctx, id, idIsDisplay...)
	if err != nil {
		return c.SetError(err)
	}

	// check if tenancy with such role for this customer exists
	err = t.Manager.CheckDuplicateRole(ctx, c, tenancy.CustomerId(), role)
	if err != nil {
		return err
	}

	// update field
	err = t.CRUD.Update(ctx, &tenancy.TenancyDb, db.Fields{"role": role})
	if err != nil {
		c.SetMessage("failed to update tenancy")
		return c.SetError(err)
	}

	// save oplog
	t.OpLog(ctx, multitenancy.OpSetRole, &multitenancy.OpLogTenancy{TenancyId: tenancy.GetID(),
		Role: tenancy.Role(), Customer: tenancy.CustomerDisplay()})

	// publish notification
	t.PublishOp(tenancy, multitenancy.OpSetRole)

	// done
	return nil
}

func (t *TenancyController) Activate(ctx op_context.Context, id string, idIsDisplay ...bool) error {

	// setup
	c := ctx.TraceInMethod("TenancyController.activate")
	defer ctx.TraceOutMethod()

	// find tenancy
	tenancy, err := t.Find(ctx, id, idIsDisplay...)
	if err != nil {
		return c.SetError(err)
	}

	// update field
	err = t.CRUD.Update(ctx, &tenancy.TenancyDb, db.Fields{"active": true})
	if err != nil {
		c.SetMessage("failed to update tenancy")
		return c.SetError(err)
	}

	// save oplog
	t.OpLog(ctx, multitenancy.OpActivate, &multitenancy.OpLogTenancy{TenancyId: tenancy.GetID(),
		Role: tenancy.Role(), Customer: tenancy.CustomerDisplay()})

	// publish notification
	t.PublishOp(tenancy, multitenancy.OpActivate)

	// done
	return nil
}

func (t *TenancyController) Deactivate(ctx op_context.Context, id string, idIsDisplay ...bool) error {

	// setup
	c := ctx.TraceInMethod("TenancyController.activate")
	defer ctx.TraceOutMethod()

	// find tenancy
	tenancy, err := t.Find(ctx, id, idIsDisplay...)
	if err != nil {
		return c.SetError(err)
	}

	// update field
	err = t.CRUD.Update(ctx, &tenancy.TenancyDb, db.Fields{"active": false})
	if err != nil {
		c.SetMessage("failed to update tenancy")
		return c.SetError(err)
	}

	// save oplog
	t.OpLog(ctx, multitenancy.OpDeactivate, &multitenancy.OpLogTenancy{TenancyId: tenancy.GetID(),
		Role: tenancy.Role(), Customer: tenancy.CustomerDisplay()})

	// publish notification
	t.PublishOp(tenancy, multitenancy.OpDeactivate)

	// done
	return nil
}

func (t *TenancyController) SetCustomer(ctx op_context.Context, id string, customer string, idIsDisplay ...bool) error {

	// setup
	c := ctx.TraceInMethod("TenancyController.SetRole")
	defer ctx.TraceOutMethod()

	// find tenancy
	tenancy, err := t.Find(ctx, id, idIsDisplay...)
	if err != nil {
		return c.SetError(err)
	}

	// find customer
	cust, err := t.Manager.FindCustomer(ctx, c, customer)
	if err != nil {
		return c.SetError(err)
	}

	// check if tenancy with such role for this customer exists
	err = t.Manager.CheckDuplicateRole(ctx, c, cust.GetID(), tenancy.Role())
	if err != nil {
		return err
	}

	// update field
	err = t.CRUD.Update(ctx, &tenancy.TenancyDb, db.Fields{"customer_id": cust.GetID()})
	if err != nil {
		c.SetMessage("failed to update tenancy")
		return c.SetError(err)
	}

	// save oplog
	t.OpLog(ctx, multitenancy.OpSetCustomer, &multitenancy.OpLogTenancy{TenancyId: tenancy.GetID(),
		Role: tenancy.Role(), Customer: cust.Display()})

	// publish notification
	t.PublishOp(tenancy, multitenancy.OpSetCustomer)

	// done
	return nil
}

func (t *TenancyController) ChangePoolOrDb(ctx op_context.Context, id string, poolId string, dbName string, idIsDisplay ...bool) error {

	// setup
	c := ctx.TraceInMethod("TenancyController.ChangePoolOrDb")
	defer ctx.TraceOutMethod()

	// find tenancy
	tenancy, err := t.Find(ctx, id, idIsDisplay...)
	if err != nil {
		return c.SetError(err)
	}

	// find pool
	pId := poolId
	if pId == "" {
		pId = tenancy.PoolId()
	}
	pool, err := t.Manager.FindPool(ctx, c, pId)
	if err != nil {
		return c.SetError(err)
	}

	// check if tenancy with such path in new pool exists
	if pool.GetID() != tenancy.PoolId() {
		err = t.Manager.CheckDuplicatePath(ctx, c, pool.GetID(), tenancy.Path())
		if err != nil {
			return err
		}
	}

	// check database
	dbN := dbName
	if dbN == "" {
		dbN = tenancy.DbName()
	}
	tenancy.DBNAME = dbN
	tenancy.POOL_ID = pool.GetID()
	checkTenancy := NewTenancy(t.Manager)
	err = checkTenancy.Init(ctx, &tenancy.TenancyDb)
	if err != nil {
		c.SetMessage("failed to init tenancy with new parameters")
		return c.SetError(err)
	}

	// update fields
	err = t.CRUD.Update(ctx, &tenancy.TenancyDb, db.Fields{"pool_id": pool.GetID(), "dbname": dbN})
	if err != nil {
		c.SetMessage("failed to update tenancy")
		return c.SetError(err)
	}

	// save oplog
	t.OpLog(ctx, multitenancy.OpChangePoolOrDb, &multitenancy.OpLogTenancy{TenancyId: tenancy.GetID(),
		Role: tenancy.Role(), Customer: tenancy.CustomerDisplay(), Pool: pool.Name(), DbName: dbN})

	// publish notification
	t.PublishOp(tenancy, multitenancy.OpChangePoolOrDb)

	// done
	return nil
}

func (t *TenancyController) Delete(ctx op_context.Context, id string, withDatabase bool, idIsDisplay ...bool) error {

	// setup
	c := ctx.TraceInMethod("TenancyController.Delete")
	defer ctx.TraceOutMethod()

	if withDatabase {
		ctx.SetGenericError(generic_error.New(generic_error.ErrorCodeUnsupported, "Database can not be droped through administrator. Use raw database tools for database dropping."))
		return c.SetError(errors.New("database can not be dropped"))
	}

	// find tenancy
	tenancy, err := t.Find(ctx, id, idIsDisplay...)
	if err != nil {
		return c.SetError(err)
	}

	// delete tenancy
	err = t.CRUD.Delete(ctx, tenancy)
	if err != nil {
		c.SetMessage("failed to delete tenancy")
		return c.SetError(err)
	}

	// save oplog
	t.OpLog(ctx, multitenancy.OpDelete, &multitenancy.OpLogTenancy{TenancyId: tenancy.GetID(),
		Role: tenancy.Role(), Customer: tenancy.CustomerDisplay()})

	// publish notification
	t.PublishOp(tenancy, multitenancy.OpDelete)

	// done
	return nil
}
