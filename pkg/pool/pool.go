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

type PoolBaseData struct {
	common.WithUniqueNameBase
	common.WithLongNameBase
	common.WithDescriptionBase
	common.WithActiveBase
}

func (p *PoolBaseData) Fill(pool Pool) {
	p.SetName(pool.Name())
	p.SetActive(pool.IsActive())
	p.SetDescription(pool.Description())
	p.SetLongName(pool.LongName())
}

type PoolBaseEssentials struct {
	common.ObjectBase
	PoolBaseData
}

type PoolItem struct {
	PoolBaseEssentials

	Services []PoolServiceBinding `json:"services"`
}

type PoolBase struct {
	PoolBaseEssentials
	Services map[string]PoolServiceBinding `gorm:"-:all"`
}

func NewPool() *PoolBase {
	p := &PoolBase{}
	return p
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
