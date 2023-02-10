package pool

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type Pool interface {
	common.Object
	common.WithUniqueName
	common.WithLongName
	common.WithDescription
	common.WithActive

	Services() []PoolService
	Service(id string) PoolService
	ServiceByName(name string) PoolService
	ServiceByRefId(serviceType string, refid string) PoolService

	AddService(service PoolService)
	DeleteService(id string)
}

var NilPool Pool

type PoolBaseConfig struct {
	common.ObjectBase
	common.WithUniqueNameBase
	common.WithLongNameBase
	common.WithDescriptionBase
	common.WithActiveBase
}

type PoolBase struct {
	PoolBaseConfig
	services       map[string]PoolService `json:"-"`
	servicesByName map[string]PoolService `json:"-"`
}

func NewPool() *PoolBase {
	p := &PoolBase{}
	p.services = make(map[string]PoolService)
	p.servicesByName = make(map[string]PoolService)
	return p
}

func (p *PoolBase) Config() interface{} {
	return &p.PoolBaseConfig
}

func (p *PoolBase) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	err := object_config.LoadLogValidate(cfg, log, vld, p, "pool", configPath...)
	if err != nil {
		return log.PushFatalStack("failed to init Pool", err)
	}

	return nil
}

func (p *PoolBase) Services() []PoolService {
	services := utils.AllMapValues(p.services)
	return services
}

func (p *PoolBase) Service(id string) PoolService {
	s := p.services[id]
	return s
}

func (p *PoolBase) ServiceByName(name string) PoolService {
	s := p.servicesByName[name]
	return s
}

func (p *PoolBase) ServiceByRefId(serviceType string, refid string) PoolService {
	for _, s := range p.services {
		if s.Type() == serviceType && s.RefId() == refid {
			return s
		}
	}
	return nil
}

func (p *PoolBase) AddService(service PoolService) {
	p.servicesByName[service.Name()] = service
	p.services[service.GetID()] = service
}

func (p *PoolBase) DeleteService(id string) {
	s := p.Service(id)
	if s != nil {
		delete(p.servicesByName, s.Name())
		delete(p.services, id)
	}
}

func (PoolBase) TableName() string {
	return "pools"
}
