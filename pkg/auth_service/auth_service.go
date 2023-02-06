package auth_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
)

// Login endpoint is derived from no handler endpoint because all processing in performed in auth preprocessing.
type LoginEndpoint struct {
	api_server.ResourceEndpoint
	api_server.EndpointNoHandler
}

func NewLoginEndpoint() *LoginEndpoint {
	ep := &LoginEndpoint{}
	ep.Init("login", "Login", ep, access_control.Post)
	return ep
}

// Logout endpoint is derived from no handler endpoint because all processing in performed in auth preprocessing.
type LogoutEndpoint struct {
	api_server.ResourceEndpoint
	api_server.EndpointNoHandler
}

func NewLogoutEndpoint() *LogoutEndpoint {
	ep := &LogoutEndpoint{}
	ep.Init("logout", "Logout", ep, access_control.Post)
	return ep
}

// Refresh endpoint is derived from no handler endpoint because all processing in performed in auth preprocessing.
type RefreshEndpoint struct {
	api_server.ResourceEndpoint
	api_server.EndpointNoHandler
}

func NewRefreshEndpoint() *RefreshEndpoint {
	ep := &RefreshEndpoint{}
	ep.Init("refresh", "Refresh", ep, access_control.Post)
	return ep
}

type AuthService struct {
	api_server.ServiceBase
}

func NewAuthService() *AuthService {
	s := &AuthService{}
	s.Init("auth")
	s.AddChildren(NewLoginEndpoint(), NewLogoutEndpoint(), NewRefreshEndpoint())
	return s
}
