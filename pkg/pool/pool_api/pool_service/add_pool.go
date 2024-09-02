package pool_service

import (
	"github.com/evgeniums/go-utils/pkg/api/api_server"
	"github.com/evgeniums/go-utils/pkg/pool"
	"github.com/evgeniums/go-utils/pkg/pool/pool_api"
)

type AddPoolEndpoint struct {
	PoolEndpoint
}

func (e *AddPoolEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("pool.AddPool")
	defer request.TraceOutMethod()

	// parse command
	cmd := pool.NewPool()
	err := request.ParseValidate(cmd)
	if err != nil {
		c.SetMessage("failed to parse/validate command")
		return err
	}

	// add pool
	p, err := e.service.Pools.AddPool(request, cmd)
	if err != nil {
		c.SetMessage("failed to add pool")
		return c.SetError(err)
	}

	// set response
	resp := &pool_api.PoolResponse{}
	resp.PoolBase = p.(*pool.PoolBase)
	request.Response().SetMessage(resp)

	// done
	return nil
}

func AddPool(s *PoolService) *AddPoolEndpoint {
	e := &AddPoolEndpoint{}
	e.Construct(s, pool_api.AddPool())
	return e
}
