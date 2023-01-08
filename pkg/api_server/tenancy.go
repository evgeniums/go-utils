package api_server

import "github.com/evgeniums/go-backend-helpers/pkg/common"

type Tenancy interface {
	common.Object
	common.WithNameAndPath
}
