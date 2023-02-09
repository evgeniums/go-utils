package user

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_methods/auth_login_phash"
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_methods/auth_sms"
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
	SetBlocked(val bool)

	Email() string
	SetEmail(email string)

	DbUser() interface{}

	ToCmd(password string) interface{}

	api.WithHateoasLinks
}

// TODO Configure somewhere unique indexes for phone and login if required
// TODO Add validation rules for user fields

type UserBaseFields struct {
	LOGIN   string `gorm:"uniqueIndex" json:"login"`
	PHONE   string `gorm:"index" json:"phone"`
	EMAIL   string `gorm:"index" json:"email"`
	BLOCKED bool   `gorm:"index" json:"blocked"`
}

func (u *UserBaseFields) Display() string {
	return u.LOGIN
}

func (u *UserBaseFields) Login() string {
	return u.LOGIN
}

func (u *UserBaseFields) SetLogin(login string) {
	u.LOGIN = login
}

func (u *UserBaseFields) Phone() string {
	return u.PHONE
}

func (u *UserBaseFields) SetPhone(phone string) {
	u.PHONE = phone
}

func (u *UserBaseFields) IsBlocked() bool {
	return u.BLOCKED
}

func (u *UserBaseFields) SetBlocked(val bool) {
	u.BLOCKED = val
}

func (u *UserBaseFields) Email() string {
	return u.EMAIL
}

func (u *UserBaseFields) SetEmail(email string) {
	u.EMAIL = email
}

func (u *UserBaseFields) SetUserFields(ctx op_context.Context, user User) error {
	user.SetEmail(u.Email())
	user.SetPhone(u.Phone())
	user.SetBlocked(u.IsBlocked())
	return nil
}

type UserFieldsWithPassword struct {
	UserBaseFields
	PlainPassword string `gorm:"-:all" json:"password"`
}

func (u *UserFieldsWithPassword) Password() string {
	return u.PlainPassword
}

func (u *UserFieldsWithPassword) SetPassword(password string) {
	u.PlainPassword = password
}

func NewUserFieldsWihPassword() *UserFieldsWithPassword {
	u := &UserFieldsWithPassword{}
	return u
}

type UserBaseDB struct {
	common.ObjectBase
	UserBaseFields
	auth_login_phash.UserBase
	api.ResponseHateous
}

func (u *UserBaseDB) ToCmd(password string) interface{} {
	cmd := &UserFieldsWithPassword{}
	cmd.UserBaseFields = u.UserBaseFields
	cmd.SetPassword(password)
	return cmd
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

type UserFieldsSetter[T User] interface {
	Login() string
	Password() string
	SetUserFields(ctx op_context.Context, user T) error
}

type SetUserFields[UserType interface{}] func(ctx op_context.Context, user UserType) error

type UserFieldsSetterBase[T User] struct {
	UserFieldsWithPassword
}

func (u *UserFieldsSetterBase[T]) SetUserFields(ctx op_context.Context, user T) error {
	user.SetLogin(u.Login())
	user.SetPassword(u.Password())
	return u.UserFieldsWithPassword.SetUserFields(ctx, user)
}

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
