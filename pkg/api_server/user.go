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

type UserBaseDB struct {
	common.ObjectBase
	Login    string `gorm:"uniqueIndex"`
	Password string
}

type UserBase struct {
	auth.UserBase
	baseDb UserBaseDB
}

func (u *UserBase) Display() string {
	return u.baseDb.Login
}

func (u *UserBase) Login() string {
	return u.baseDb.Login
}

func (u *UserBase) GetID() string {
	return u.baseDb.GetID()
}

func (u *UserBase) GetAuthParameter(authMethodProtocol string, key string) string {
	// TODO return basic fields
	return ""
}

func (u *UserBase) Roles() []access_control.Role {
	// TODO implement user roles
	return nil
}
