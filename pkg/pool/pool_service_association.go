package pool

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
)

type PoolServiceAssociacion interface {
	common.Object
	Pool() string
	Role() string
	Service() string
}

type PoolServiceAssociacionBase struct {
	common.ObjectBase
	POOL_ID    string `gorm:"index;index:,unique,composite:u" json:"pool_id" validate:"required" vmessage:"Pool ID can not be empty"`
	ROLE       string `gorm:"index;index:,unique,composite:u" json:"type" validate:"required" vmessage:"Role can not be empty"`
	SERVICE_ID string `gorm:"index" json:"service_id" validate:"required" vmessage:"Service ID can not be empty"`
}

func (p *PoolServiceAssociacionBase) Pool() string {
	return p.POOL_ID
}

func (p *PoolServiceAssociacionBase) Role() string {
	return p.ROLE
}

func (p *PoolServiceAssociacionBase) Service() string {
	return p.SERVICE_ID
}

func (PoolServiceAssociacionBase) TableName() string {
	return "pool_service_associations"
}

type PoolServiceBinding interface {
	PoolServiceAssociacion
	PoolServiceEssentials
}

type PoolServiceBindingBase struct {
	PoolServiceAssociacionBase
	PoolServiceBaseEssentials
}
