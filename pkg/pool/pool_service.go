package pool

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
)

type Secrets interface {
	Secret1() string
	Secret2() string
}

type ServiceConfig interface {
	PublicHost() string
	PublicPort() uint16
	PublicUrl() string
	PrivateHost() string
	PrivatePort() uint16
	PrivateUrl() string
	Parameter1() string
	Parameter2() string
	Parameter3() string
}

type PoolServiceEssentials interface {
	common.WithLongName
	common.WithDescription
	common.WithActive
	common.WithType
	common.WithRefId
	ServiceConfig
}

type PoolService interface {
	common.Object
	common.WithUniqueName
	PoolServiceEssentials
	Secrets
}

var NilService PoolService

type SecretsBase struct {
	SECRET1 string `json:"secret1" long:"secret1" description:"Secret1 the service (optional)"`
	SECRET2 string `json:"secret2" long:"secret2" description:"Secret2 the service (optional)"`
}

func (s *SecretsBase) Secret1() string {
	return s.SECRET1
}

func (s *SecretsBase) Secret2() string {
	return s.SECRET2
}

type ServiceConfigBase struct {
	PUBLIC_HOST  string `gorm:"index" json:"public_host" long:"public-host" description:"Public host of the service (optional)"`
	PUBLIC_PORT  uint16 `gorm:"index" json:"public_port" long:"public-port" description:"Public port of the service (optional)"`
	PUBLIC_URL   string `gorm:"index" json:"public_url" long:"public-url" description:"Public url of the service (optional)"`
	PRIVATE_HOST string `gorm:"index" json:"private_host" long:"private-host" description:"Private host of the service (optional)"`
	PRIVATE_PORT uint16 `gorm:"index" json:"private_port" long:"private-port" description:"Private port of the service (optional)"`
	PRIVATE_URL  string `gorm:"index" json:"private_url" long:"private-url" description:"Private URL of the service (optional)"`
	PARAMETER1   string `gorm:"index;column:parameter1" json:"parameter1" long:"parameter1" description:"Generic parameter1 of the service (optional)"`
	PARAMETER2   string `gorm:"index;column:parameter2" json:"parameter2" long:"parameter2" description:"Generic parameter2 of the service (optional)"`
	PARAMETER3   string `gorm:"index;column:parameter3" json:"parameter3" long:"parameter3" description:"Generic parameter3 of the service (optional)"`
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

func (s *ServiceConfigBase) Parameter1() string {
	return s.PARAMETER1
}

func (s *ServiceConfigBase) Parameter2() string {
	return s.PARAMETER2
}

func (s *ServiceConfigBase) Parameter3() string {
	return s.PARAMETER3
}

type PoolServiceBaseData struct {
	common.WithLongNameBase
	common.WithDescriptionBase
	common.WithActiveBase
	common.WithTypeBase
	common.WithRefIdBase
	ServiceConfigBase
}

type PoolServiceBaseEssentials struct {
	PoolServiceBaseData
	common.WithUniqueNameBase
}

type PoolServiceBase struct {
	common.ObjectBase
	PoolServiceBaseEssentials
	SecretsBase
}

func NewService() *PoolServiceBase {
	p := &PoolServiceBase{}
	return p
}

func (PoolServiceBase) TableName() string {
	return "pool_services"
}
