package pool

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type PoolService interface {
	common.Object
	common.WithName
	common.WithDescription
	common.WithActive
	common.WithType
	common.WithRefId
}

type PoolServiceBaseConfig struct {
	common.ObjectBase
	common.WithNameBase
	common.WithDescriptionBase
	common.WithActiveBase
	common.WithTypeBase
	common.WithRefIdBase
}

type PoolServiceBase struct {
	PoolServiceBaseConfig
}

func NewPoolService() *PoolServiceBase {
	p := &PoolServiceBase{}
	return p
}

func (p *PoolServiceBase) Config() interface{} {
	return &p.PoolServiceBaseConfig
}

func (p *PoolServiceBase) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	err := object_config.LoadLogValidate(cfg, log, vld, p, "pool_service", configPath...)
	if err != nil {
		return log.PushFatalStack("failed to init PoolService", err)
	}

	return nil
}

func (PoolServiceBase) TableName() string {
	return "pool_services"
}
