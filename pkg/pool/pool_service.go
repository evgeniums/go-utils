package pool

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
)

type ServiceConfig interface {
	PublicHost() string
	PublicPort() uint16
	PublicUrl() string
	PrivateHost() string
	PrivatePort() uint16
	PrivateUrl() string
	Secret1() string
	Secret2() string
	Parameter1() string
	Parameter2() string
	Parameter3() string
}

type PoolServiceEssentials interface {
	common.WithUniqueName
	common.WithLongName
	common.WithDescription
	common.WithActive
	common.WithType
	common.WithRefId
	ServiceConfig
}

type PoolService interface {
	common.Object
	PoolServiceEssentials
}

var NilService PoolService

type ServiceConfigBase struct {
	PUBLIC_HOST  string `gorm:"index" json:"public_host"`
	PUBLIC_PORT  uint16 `gorm:"index" json:"public_port"`
	PUBLIC_URL   string `gorm:"index" json:"public_url"`
	PRIVATE_HOST string `gorm:"index" json:"private_host"`
	PRIVATE_PORT uint16 `gorm:"index" json:"private_port"`
	PRIVATE_URL  string `gorm:"index" json:"private_url"`
	SECRET1      string `json:"-" mask:"true"`
	SECRET2      string `json:"-" mask:"true"`
	PARAMETER1   string `gorm:"index" json:"parameter1"`
	PARAMETER2   string `gorm:"index" json:"parameter2"`
	PARAMETER3   string `gorm:"index" json:"parameter3"`
}

func (s *ServiceConfigBase) PublicHost() string {
	return s.PUBLIC_HOST
}

func (s *ServiceConfigBase) PublicPort() uint16 {
	return s.PUBLIC_PORT
}

func (s *ServiceConfigBase) PublicUrl() string {
	return s.PUBLIC_URL
}

func (s *ServiceConfigBase) PrivateHost() string {
	return s.PRIVATE_HOST
}

func (s *ServiceConfigBase) PrivatePort() uint16 {
	return s.PRIVATE_PORT
}

func (s *ServiceConfigBase) PrivateUrl() string {
	return s.PRIVATE_URL
}

func (s *ServiceConfigBase) Secret1() string {
	return s.SECRET1
}

func (s *ServiceConfigBase) Secret2() string {
	return s.SECRET2
}

func (s *ServiceConfigBase) Parameter1() string {
	return s.PARAMETER1
}

func (s *ServiceConfigBase) Parameter2() string {
	return s.PARAMETER2
}

func (s *ServiceConfigBase) Parameter3() string {
	return s.PARAMETER3
}

type PoolServiceBaseEssentials struct {
	common.WithUniqueNameBase
	common.WithLongNameBase
	common.WithDescriptionBase
	common.WithActiveBase
	common.WithTypeBase
	common.WithRefIdBase
	ServiceConfigBase
}

type PoolServiceBase struct {
	common.ObjectBase
	PoolServiceBaseEssentials
}

func NewPoolService() *PoolServiceBase {
	p := &PoolServiceBase{}
	return p
}

func (PoolServiceBase) TableName() string {
	return "pool_services"
}

// type PoolServiceBinding interface {
// 	common.Object
// 	common.WithName
// 	common.WithLongName
// 	common.WithDescription
// 	Pool() string
// 	Type() string
// 	Service() string
// }

// type PoolServiceBindingBaseConfig struct {
// 	common.ObjectBase
// 	common.WithNameBase
// 	common.WithLongNameBase
// 	common.WithDescriptionBase
// 	POOL_ID    string `gorm:"index;index:,unique,composite:u" json:"pool_id"`
// 	TYPE       string `gorm:"index;index:,unique,composite:u" json:"type"`
// 	SERVICE_ID string `gorm:"index" json:"service_id"`
// }

// func (p *PoolServiceBindingBaseConfig) Pool() string {
// 	return p.POOL_ID
// }

// func (p *PoolServiceBindingBaseConfig) Type() string {
// 	return p.TYPE
// }

// func (p *PoolServiceBindingBaseConfig) Service() string {
// 	return p.SERVICE_ID
// }

// type PoolServiceBindingBase struct {
// 	PoolServiceBindingBaseConfig
// }

// func (p *PoolServiceBindingBase) Config() interface{} {
// 	return &p.PoolServiceBindingBaseConfig
// }

// func (PoolServiceBindingBase) TableName() string {
// 	return "pool_service_bindings"
// }
