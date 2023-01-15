package multitenancy

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type Tenancy interface {
	common.Object
	db.WithDB

	Name() string
	Path() string
}

type TenancyObjectBase struct {
	common.ObjectBase
	db.DBConfig

	name string `gorm:"index"`
	path string `gorm:"uniqueIndex"`
}

func (TenancyObjectBase) TableName() string {
	return "tenancies"
}

func (t *TenancyObjectBase) SetDbConfig(cfg db.DBConfig) {
	t.DBConfig = cfg
}

func (t *TenancyObjectBase) Path() string {
	return t.path
}

func (t *TenancyObjectBase) Name() string {
	return t.name
}

type TenancyBase struct {
	TenancyObjectBase

	db.WithDBBase
}

func (t *TenancyBase) ConnectDB(ctx op_context.Context) error {

	localCtx := ctx.TraceInMethod("TenancyBase.ConnectDB")
	defer ctx.TraceOutMethod()

	t.WithDBBase.Init(ctx.DB().NewDB())
	return localCtx.SetError(t.DB().InitWithConfig(ctx, ctx.App().Validator(), &t.DBConfig))
}
