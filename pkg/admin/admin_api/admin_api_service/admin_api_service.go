package admin_api_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/admin"
	"github.com/evgeniums/go-backend-helpers/pkg/admin/admin_api"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api/user_service"
)

type AdminService = user_service.UserService[*admin.Admin]

func NewAdminService(admins *admin.Manager) *AdminService {
	return user_service.NewUserService[*admin.Admin](admins, admin_api.NewAdminFieldsSetter, "admin")
}
