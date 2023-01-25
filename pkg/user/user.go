package user

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_login_phash"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_sms"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
)

type User interface {
	auth.User
	auth_login_phash.User
	auth_sms.UserWithPhone

	DbUser() interface{}
}

type UserBaseDB struct {
	common.ObjectBase
	auth_login_phash.UserBase
	LOGIN   string `gorm:"uniqueIndex"`
	PHONE   string `gorm:"uniqueIndex"`
	BLOCKED bool   `gorm:"index"`
}

func (u *UserBaseDB) Display() string {
	return u.LOGIN
}

func (u *UserBaseDB) Login() string {
	return u.LOGIN
}

func (u *UserBaseDB) Phone() string {
	return u.PHONE
}

func (u *UserBaseDB) IsBlocked() bool {
	return u.BLOCKED
}

type UserBase struct {
	UserBaseDB
}

func NewUser() *UserBase {
	u := &UserBase{}
	return u
}

func (u *UserBase) DbUser() interface{} {
	return &u.UserBaseDB
}
