package user_session_default

import (
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_default"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
)

type User = user_default.User

type Users struct {
	user.UsersWithSession[*user_default.User, *user_manager.SessionBase, *user_manager.SessionClientBase]
}

func NewUsers() *Users {
	m := &Users{}
	m.MakeUser = user_default.NewUser
	m.MakeSession = user_manager.NewSession
	m.MakeSessionClient = user_manager.NewSessionClient
	return m
}
