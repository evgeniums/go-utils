package admin

import (
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_session_default"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
)

type Admin struct {
	user_session_default.User
}

func NewAdmin() *Admin {
	return &Admin{}
}

type AdminSession struct {
	user_manager.SessionBase
}

func NewAdminSession() *AdminSession {
	return &AdminSession{}
}

type AdminSessionClient struct {
	user_session_default.UserSessionClient
}

func NewAdminSessionClient() *AdminSessionClient {
	return &AdminSessionClient{}
}
