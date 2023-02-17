package multitenancy

import (
	"github.com/evgeniums/go-backend-helpers/pkg/cache"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
)

type Tenancy interface {
	common.Object
	common.WithActive
	common.WithDescription

	Name() string
	Path() string
	CustomerId() string
	Role() string

	Db() db.DB
	Pool() pool.Pool
	Cache() cache.Cache
}

type WithTenancy interface {
	GetTenancy() Tenancy
}

type TenancyDbEssentials struct {
	CUSTOMER string `gorm:"index;index:,unique,composite:u" validate:"required,alphanum_|email" vmessage:"Invalid customer ID"`
	ROLE     string `gorm:"index;index:,unique,composite:u" validate:"required,alphanum_" vmessage:"Role must be aphanumeric not empty"`
	NAME     string `gorm:"index" validate:"required" vmessage:"Name can not be empty"`
	PATH     string `gorm:"uniqueIndex" validate:"omitempty,alphanum_" vmessage:"Path must be alhanumeric"`
	POOL_ID  string `gorm:"index" validate:"required,alphanum" vmessage:"Pool ID must be alhanumeric not empty"`
}

type TenancyDb struct {
	common.ObjectBase
	common.WithActiveBase
	common.WithDescriptionBase
	TenancyDbEssentials
}

func (TenancyDb) TableName() string {
	return "tenancies"
}

func (t *TenancyDb) Path() string {
	return t.PATH
}

func (t *TenancyDb) Name() string {
	return t.NAME
}

func (t *TenancyDb) CustomerId() string {
	return t.CUSTOMER
}

func (t *TenancyDb) Role() string {
	return t.ROLE
}

type TenancyBaseData struct {
	TenancyDb

	db.WithDBBase
	Cache cache.Cache
	Pool  pool.Pool
}

type TenancyBase struct {
	TenancyBaseData
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

func (t *TenancyBase) Init(ctx op_context.Context, pools pool.PoolStore, data *TenancyDb) error {

	t.TenancyDb = *data
	t.SetCache(ctx.Cache())

	// TODO find pool

	// TODO find database service in pool

	// done
	return nil
}

func (t *TenancyBase) ConnectServices(ctx op_context.Context) error {

	// TODO connect to database

	return nil
}

// func (t *TenancyBase) ConnectDB(ctx op_context.Context) error {

// 	localCtx := ctx.TraceInMethod("TenancyBase.ConnectDB")
// 	defer ctx.TraceOutMethod()

// 	t.WithDBBase.Init(ctx.Db().NewDB())
// 	return localCtx.SetError(t.Db().InitWithConfig(ctx, ctx.App().Validator(), &t.DBConfig))
// }
