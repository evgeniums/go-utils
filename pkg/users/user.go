package users

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_login_phash"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_sms"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
)

type User interface {
	auth.User
	auth_login_phash.UserWithPasswordHash
	auth_sms.UserWithPhone

	DbUser() interface{}
}

type UserBaseDB struct {
	common.ObjectBase
	Login        string `gorm:"uniqueIndex"`
	PasswordHash string
	PasswordSalt string
	Phone        string `gorm:"uniqueIndex"`
}

type UserBase struct {
	auth.UserBase
	dbUser UserBaseDB
}

func (u *UserBase) Display() string {
	return u.dbUser.Login
}

func (u *UserBase) Login() string {
	return u.dbUser.Login
}

func (u *UserBase) GetID() string {
	return u.dbUser.GetID()
}

func (u *UserBase) PasswordHash() string {
	return u.dbUser.PasswordHash
}

func (u *UserBase) PasswordSalt() string {
	return u.dbUser.PasswordSalt
}

func (u *UserBase) Phone() string {
	return u.dbUser.Phone
}
