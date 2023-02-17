package multitenancy

import (
	"github.com/evgeniums/go-backend-helpers/pkg/cache"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
)

type Tenancy interface {
	common.Object
	common.WithActive
	common.WithDescription

	Display() string
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

type TenancyData struct {
	common.WithDescriptionBase
	CUSTOMER string `gorm:"index;index:,unique,composite:u" validate:"required,alphanum_|email" vmessage:"Invalid customer ID"`
	ROLE     string `gorm:"index;index:,unique,composite:u" validate:"required,alphanum_" vmessage:"Role must be aphanumeric not empty"`
	PATH     string `gorm:"uniqueIndex" validate:"omitempty,alphanum_" vmessage:"Path must be alhanumeric"`
	POOL_ID  string `gorm:"index" validate:"required,alphanum" vmessage:"Pool ID must be alhanumeric not empty"`
	DBNAME   string `gorm:"index" validate:"omitempty,alphanum_" vmessage:"Database name must be alhanumeric"`
}

type TenancyDb struct {
	common.ObjectBase
	common.WithActiveBase
	TenancyData
}

func (TenancyDb) TableName() string {
	return "tenancies"
}

func (t *TenancyDb) Path() string {
	return t.PATH
}

func (t *TenancyDb) CustomerId() string {
	return t.CUSTOMER
}

func (t *TenancyDb) Role() string {
	return t.ROLE
}
