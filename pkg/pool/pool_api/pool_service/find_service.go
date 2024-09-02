package pool_service

import (
	"errors"

	"github.com/evgeniums/go-utils/pkg/api/api_server"
	"github.com/evgeniums/go-utils/pkg/pool"
	"github.com/evgeniums/go-utils/pkg/pool/pool_api"
)

type FindServiceEndpoint struct {
	PoolEndpoint
}

func (e *FindServiceEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("pool.FindService")
	defer request.TraceOutMethod()

	// find service
	s, err := e.service.Pools.FindService(request, request.GetResourceId("service"))
	if err != nil {
		c.SetMessage("failed to find service")
		return c.SetError(err)
	}
	if s == nil {
		request.SetGenericErrorCode(pool.ErrorCodeServiceNotFound)
		return c.SetError(errors.New("service not found"))
	}

	// set response
	resp := &pool_api.ServiceResponse{}
	resp.PoolServiceBase = s.(*pool.PoolServiceBase)
	request.Response().SetMessage(resp)

	// done
	return nil
}

func FindService(s *PoolService) *FindServiceEndpoint {
	e := &FindServiceEndpoint{}
	e.Construct(s, pool_api.FindService())
	return e
}
