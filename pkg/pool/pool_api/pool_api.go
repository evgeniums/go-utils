package pool_api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
)

type PoolResponse struct {
	api.ResponseHateous
	Pool *pool.PoolBase `json:"pool"`
}

type PoolServiceResponse struct {
	api.ResponseHateous
	pool.PoolServiceBase `json:"pool_service"`
}
