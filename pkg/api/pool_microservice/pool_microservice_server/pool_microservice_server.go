package pool_microservice_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/noauth_server"
)

type PoolMicroserviceServer struct {
	noauth_server.NoAuthServer
}

type Config struct {
	noauth_server.Config
}

func New(config ...*Config) *PoolMicroserviceServer {
	s := &PoolMicroserviceServer{}
	s.Construct(config...)
	return s
}

func (s *PoolMicroserviceServer) Construct(config ...*Config) {
	if len(config) != 0 {
		s.NoAuthServer.Construct(config[0].Config)
	}
}
