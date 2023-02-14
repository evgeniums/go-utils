package pool_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api"
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
		c.SetMessage("faield to parse/validate command")
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
	e.Construct(s, api.Add())
	return e
}
