package user

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_login_phash"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_sms"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type User interface {
	common.Object
	auth.User
	auth_login_phash.User
	auth_sms.UserWithPhone

	SetLogin(login string)
	SetPhone(phone string)

	Email() string
	SetEmail(email string)

	DbUser() interface{}
}

// TODO Configure somewhere unique indexes for phone and login if required
// TODO Add validation rules for user fields
type UserBaseDB struct {
	common.ObjectBase
	auth_login_phash.UserBase
	LOGIN   string `gorm:"uniqueIndex" json:"login"`
	PHONE   string `gorm:"index" json:"phone"`
	EMAIL   string `gorm:"index" json:"email"`
	BLOCKED bool   `gorm:"index" json:"blocked"`
}

func (u *UserBaseDB) Display() string {
	return u.LOGIN
}

func (u *UserBaseDB) Login() string {
	return u.LOGIN
}

func (u *UserBaseDB) SetLogin(login string) {
	u.LOGIN = login
}

func (u *UserBaseDB) Phone() string {
	return u.PHONE
}

func (u *UserBaseDB) SetPhone(phone string) {
	u.PHONE = phone
}

func (u *UserBaseDB) IsBlocked() bool {
	return u.BLOCKED
}

func (u *UserBaseDB) Email() string {
	return u.EMAIL
}

func (u *UserBaseDB) SetEmail(email string) {
	u.EMAIL = email
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

type SetUserFields[UserType User] func(ctx op_context.Context, user UserType) error

func Phone[UserType User](phone string, userSample ...UserType) SetUserFields[UserType] {
	return func(ctx op_context.Context, user UserType) error {
		user.SetPhone(phone)
		return nil
	}
}

func Email[UserType User](email string, userSample ...UserType) SetUserFields[UserType] {
	return func(ctx op_context.Context, user UserType) error {
		user.SetEmail(email)
		return nil
	}
}

func FindByLogin(controller crud.CRUD, ctx op_context.Context, login string, user interface{}, dest ...interface{}) (bool, error) {
	return controller.Read(ctx, db.Fields{"login": login}, user, dest...)
}
