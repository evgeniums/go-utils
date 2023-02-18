package multitenancy

import "github.com/evgeniums/go-backend-helpers/pkg/common"

type TenancyMeta struct {
	common.ObjectBase
}

func DbModels() []interface{} {
	return []interface{}{&TenancyMeta{}, &TenancyDb{}, &OpLogTenancy{}}
}
