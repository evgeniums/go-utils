package pool_api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
)

type PoolResponse struct {
	api.ResponseHateous
	*pool.PoolBase
}

type ServiceResponse struct {
	api.ResponseHateous
	*pool.PoolServiceBase
}

type ListServicesResponse struct {
	api.ResponseHateous
	Services []*pool.PoolServiceBase `json:"services"`
}
