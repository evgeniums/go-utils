package pool

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/common"
)

type Pool interface {
	common.Object
	common.WithUniqueName
	common.WithLongName
	common.WithDescription
	common.WithActive

	Service(role string) (PoolServiceBinding, error)
}

type PoolBaseEssentials struct {
	common.ObjectBase
	common.WithUniqueNameBase
	common.WithLongNameBase
	common.WithDescriptionBase
	common.WithActiveBase
}

type PoolItem struct {
	PoolBaseEssentials

	Services []PoolServiceBinding `json:"services"`
}

type PoolBase struct {
	PoolBaseEssentials
	Services map[string]PoolServiceBinding `gorm:"-:all"`
}

func (PoolBase) TableName() string {
	return "pools"
}

func (p *PoolBase) Service(role string) (PoolServiceBinding, error) {

	service, ok := p.Services[role]
	if !ok {
		return nil, errors.New("service not found")
	}

	return service, nil
}
