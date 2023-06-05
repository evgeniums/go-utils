package tenancy_manager

import (
	"github.com/evgeniums/go-backend-helpers/pkg/cache"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
)

type TenancyBaseData struct {
	multitenancy.TenancyDb

	db.WithDBBase
	Cache          cache.Cache
	Pool           pool.Pool
	Customer       *customer.Customer
	TenancyManager *TenancyManager
}

type TenancyBase struct {
	TenancyBaseData
}

func NewTenancy(manager *TenancyManager) *TenancyBase {
	t := &TenancyBase{}
	t.TenancyManager = manager
	return t
}

func NewContextPathOnly(tenancyPath string, fromCtx ...op_context.Context) *multitenancy.TenancyContextBase {
	c := multitenancy.NewContext(fromCtx...)
	tenancy := &TenancyBase{}
	tenancy.SetPath(tenancyPath)
	c.SetTenancy(tenancy)
	return c
}

func (d *TenancyBase) IsActive() bool {
	return d.TenancyDb.IsActive() && !d.Customer.IsBlocked() && d.TenancyBaseData.Pool.IsActive()
}

func (t *TenancyBase) Pool() pool.Pool {
	return t.TenancyBaseData.Pool
}

func (t *TenancyBase) Cache() cache.Cache {
	return t.TenancyBaseData.Cache
}

func (t *TenancyBase) SetCache(c cache.Cache) {
	t.TenancyBaseData.Cache = c
}

func (t *TenancyBase) Init(ctx op_context.Context, data *multitenancy.TenancyDb) (bool, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("TenancyBase.ConnectDatabase", logger.Fields{"customer": t.CUSTOMER_ID, "role": t.ROLE})
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	t.TenancyDb = *data
	t.SetCache(ctx.Cache())

	// find customer
	t.Customer, err = t.TenancyManager.Customers.Find(ctx, data.CUSTOMER_ID)
	if err != nil {
		c.SetMessage("failed to find customer")
		return false, err
	}
	if t.Customer == nil {
		c.SetMessage("failed to find customer")
		return false, err
	}
	c.SetLoggerField("tenancy", multitenancy.TenancyDisplay(t))

	// check if pool exists
	t.TenancyBaseData.Pool, err = t.TenancyManager.Pools.Pool(data.POOL_ID)
	if err != nil {
		ctx.SetGenericErrorCode(pool.ErrorCodePoolNotFound)
		c.SetMessage("unknown pool")
		return false, err
	}
	if !t.TenancyBaseData.Pool.IsActive() {
		c.Logger().Warn("skipping tenancy because pool is not active", logger.Fields{"pool": t.TenancyBaseData.Pool.Name()})
		return true, nil
	}

	// init database
	err = t.ConnectDatabase(ctx)
	if err != nil {
		return false, err
	}

	// check tenancy database
	err = multitenancy.CheckTenancyDatabase(ctx, t.Db(), t.GetID())
	if err != nil {
		t.Db().Close()
		return false, err
	}

	// done
	return false, nil
}

func (t *TenancyBase) ConnectDatabase(ctx op_context.Context, newDb ...bool) error {

	// setup
	var err error
	c := ctx.TraceInMethod("TenancyBase.ConnectDatabase", logger.Fields{"customer": t.CUSTOMER_ID, "role": t.ROLE})
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// connect to database
	db, err := pool.ConnectDatabaseService(ctx, t.Pool(), multitenancy.TENANCY_DATABASE_ROLE, t.DbName(), newDb...)
	if err != nil {
		return err
	}

	// done
	t.WithDBBase.Init(db)
	return nil
}

func (t *TenancyBase) CustomerDisplay() string {
	if t.Customer == nil {
		return ""
	}
	return t.Customer.Display()
}
