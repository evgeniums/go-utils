package admin_api_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/admin"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api/user_client"
)

type AdminClient = user_client.UserClient[*admin.Admin]

func NewAdminClient(client api_client.Client) *AdminClient {
	return user_client.NewUserClient(client, admin.NewAdmin, "admin")
}
