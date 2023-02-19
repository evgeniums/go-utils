package tenancy_api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
)

const ServiceName string = "tenancies"
const TenancyResource string = "tenancy"

type TenancyResponse struct {
	api.ResponseHateous
	*multitenancy.TenancyItem
}

type ListTenanciesResponse struct {
	api.ResponseCount
	api.ResponseHateous
	Tenancies []*multitenancy.TenancyItem `json:"tenancies"`
}

type DeleteTenancyCmd struct {
	WithDatabase bool `schema:"with_database" url:"with_database"`
}

var (
	List           = func() api.Operation { return api.List("list_tenancies") }
	Add            = func() api.Operation { return api.Add("add_tenancy") }
	Delete         = func() api.Operation { return api.Delete("delete_tenancy") }
	SetActive      = func() api.Operation { return api.Update("set_tenancy_active") }
	SetPath        = func() api.Operation { return api.Update("set_tenancy_path") }
	SetRole        = func() api.Operation { return api.Update("set_tenancy_role") }
	SetCustomer    = func() api.Operation { return api.Update("set_tenancy_customer") }
	ChangePoolOrDb = func() api.Operation { return api.Update("change_tenancy_pool_or_db") }
)
