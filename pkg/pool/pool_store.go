package pool

import (
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type PoolStore interface {
	Pool(id string) (Pool, error)
	SelfPool() (Pool, error)
	PoolByName(name string) (Pool, error)
}

type PoolStoreConfig struct {
	POOL_NAME string
}

type PoolStoreBase struct {
	PoolStoreConfig
	selfPool  Pool
	pools     map[string]Pool
	poolsById map[string]Pool
}

func (p *PoolStoreBase) Config() interface{} {
	return &p.PoolStoreConfig
}

func NewPoolStore() *PoolStoreBase {
	p := &PoolStoreBase{}
	p.pools = make(map[string]Pool)
	p.poolsById = make(map[string]Pool)
	return p
}

func (p *PoolStoreBase) Init(ctx op_context.Context, ctrl PoolController, configPath ...string) error {

	c := ctx.TraceInMethod("PoolStore.Init")
	ctx.TraceOutMethod()

	// load configuration
	err := object_config.LoadLogValidate(ctx.App().Cfg(), ctx.App().Logger(), ctx.App().Validator(), p, "pools", configPath...)
	if err != nil {
		c.SetMessage("failed to init PoolStore")
		return c.SetError(err)
	}

	if p.POOL_NAME == "" {
		pools, _, err := ctrl.GetPools(ctx, nil)
		if err != nil {
			c.SetMessage("failed to load pools")
			return c.SetError(err)
		}
		for _, pool := range pools {
			p.pools[pool.GetID()] = pool
			p.pools[pool.Name()] = pool
		}
	} else {
		pool, err := ctrl.FindPool(ctx, p.POOL_NAME, true)
		if err != nil {
			c.SetMessage("failed to load self pool")
			return c.SetError(err)
		}
		if pool == nil {
			return c.SetErrorStr("self pool not found")
		}
		p.pools[pool.GetID()] = pool
		p.pools[pool.Name()] = pool
		p.selfPool = pool
	}

	// done
	return nil
}
