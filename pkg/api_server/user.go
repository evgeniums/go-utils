package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
)

type User interface {
	auth.User
	access_control.Subject
}

type UserBase struct {
	common.ObjectBase

	login string `gorm:"uniqueIndex"`
	// password string
	// phone    string `gorm:"index"`
	// pubkey   string
	// email    string `gorm:"index"`
}

func (u *UserBase) Display() string {
	return u.login
}

func (u *UserBase) GetAuthParameter(authMethodProtocol string, key string) string {
	// TODO return basic fields
	return ""
}

func (u *UserBase) Roles() []access_control.Role {
	// TODO implement user roles
	return nil
}
