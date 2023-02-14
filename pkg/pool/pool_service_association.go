package pool

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
)

type PoolServiceAssociation interface {
	common.Object
	Pool() string
	Role() string
	Service() string
}

type PoolId struct {
	POOL_ID string `gorm:"index;index:,unique,composite:u" json:"pool_id"`
}

type PoolServiceAssociationRole struct {
	ROLE string `gorm:"index;index:,unique,composite:u" json:"type" validate:"required,alphanum_" vmessage:"Role name can contain only digits and letters"`
}

type PoolServiceAssociationCmd struct {
	PoolServiceAssociationRole
	SERVICE_ID string `gorm:"index" json:"service_id" validate:"required" vmessage:"Service ID can not be empty"`
}

type PoolServiceAssociationBase struct {
	common.ObjectBase
	PoolId
	PoolServiceAssociationCmd
}

func (p *PoolServiceAssociationBase) Pool() string {
	return p.POOL_ID
}

func (p *PoolServiceAssociationBase) Role() string {
	return p.ROLE
}

func (p *PoolServiceAssociationBase) Service() string {
	return p.SERVICE_ID
}

func (PoolServiceAssociationBase) TableName() string {
	return "pool_service_associations"
}

type PoolServiceBinding interface {
	PoolServiceAssociation
}

type PoolServiceBindingBase struct {
	PoolServiceAssociationBase
}
