package rest_api_gin_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	database "github.com/evgeniums/go-backend-helpers/pkg/db"
)

type Tenancy struct {
	common.ObjectBase
	common.WithNameAndPathBase

	db database.DB
}

func NewTenancy() *Tenancy {
	return &Tenancy{}
}

func (t *Tenancy) DB() database.DB {
	return t.db
}
