package admin_api_service

import (
	"github.com/evgeniums/go-utils/pkg/admin"
	"github.com/evgeniums/go-utils/pkg/admin/admin_api"
	"github.com/evgeniums/go-utils/pkg/api/api_server"
	"github.com/evgeniums/go-utils/pkg/user/user_api/user_service"
)

type AdminService = user_service.UserService[*admin.Admin]

func NewAdminService(admins *admin.Manager) *AdminService {
	s := user_service.NewUserService[*admin.Admin](admins, admin_api.NewAdminFieldsSetter, "admin")

	adminTableConfig := &api_server.DynamicTableConfig{Model: &admin.Admin{}, Operation: s.ListOperation()}
	s.AddDynamicTables(adminTableConfig)

	return s
}
