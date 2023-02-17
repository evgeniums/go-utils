package tenancy_manager

import (
	"github.com/evgeniums/go-backend-helpers/pkg/cache"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type TenancyBaseData struct {
	multitenancy.TenancyDb

	db.WithDBBase
	Cache    cache.Cache
	Pool     pool.Pool
	Customer customer.Customer
}

type TenancyBase struct {
	TenancyBaseData
}

func NewTenancy() *TenancyBase {
	t := &TenancyBase{}
	return t
}

func (d *TenancyBase) IsActive() bool {
	return d.TenancyDb.IsActive() && !d.Customer.IsBlocked()
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

func (t *TenancyBase) Init(ctx op_context.Context, pools pool.PoolStore, data *multitenancy.TenancyDb) error {

	t.TenancyDb = *data
	t.SetCache(ctx.Cache())

	// TODO find customer

	// TODO find pool

	// TODO find database service in pool

	// done
	return nil
}

func (t *TenancyBase) Display() string {
	return utils.ConcatStrings(t.Customer.Login(), "/", t.Role())
}
