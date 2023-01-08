package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
)

type Tenancy interface {
	common.Object
	common.WithNameAndPath
	DB() db.DB
}
