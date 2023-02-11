package pool_api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
)

type PoolResponse struct {
	api.ResponseHateous
	pool.Pool `json:"pool"`
}

type PoolServiceResponse struct {
	api.ResponseHateous
	pool.PoolServiceBase `json:"pool_service"`
}
