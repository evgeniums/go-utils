package pool

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
)

func ParseDbService(service *PoolServiceBaseData) (*db.DBConfig, error) {

	if service.Type() != TypeDatabase {
		return nil, errors.New("invalid service type")
	}

	d := &db.DBConfig{}
	d.DB_DSN = service.PrivateUrl()
	d.DB_HOST = service.PrivateHost()
	d.DB_PORT = service.PrivatePort()
	d.DB_PROVIDER = service.Provider()
	d.DB_USER = service.User()
	d.DB_PASSWORD = service.Secret1()
	d.DB_NAME = service.Parameter1()
	d.DB_EXTRA_CONFIG = service.Parameter2()
	return d, nil
}
