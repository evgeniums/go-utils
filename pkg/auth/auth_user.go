package auth

import (
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type User interface {
	GetID() string
	Display() string
	Login() string
	IsBlocked() bool
}

type WithUser interface {
	SetUser(user User)
	GetUserId() string
	GetUserDisplay() string
	GetUserLogin() string
}

type WithUserBase struct {
	UserId      string `gorm:"index"`
	UserLogin   string `gorm:"index"`
	UserDisplay string `gorm:"index"`
}

func (w *WithUserBase) SetUser(user User) {
	w.UserId = user.GetID()
	w.UserLogin = user.Login()
	w.UserDisplay = user.Display()
}

func (s *WithUserBase) GetUserId() string {
	return s.UserId
}

func (s *WithUserBase) GetUserDisplay() string {
	return s.UserDisplay
}

func (s *WithUserBase) GetUserLogin() string {
	return s.UserLogin
}

type Session interface {
	GetSessionId() string
	SetSessionId(id string)
	GetClientId() string
	SetClientId(id string)
}

type SessionBase struct {
	session string
	client  string
}

func (u *SessionBase) GetSessionId() string {
	return u.session
}

func (u *SessionBase) SetSessionId(id string) {
	u.session = id
}

func (u *SessionBase) GetClientId() string {
	return u.client
}

func (u *SessionBase) SetClientId(id string) {
	u.client = id
}

type WithAuthUser interface {
	AuthUser() User
	SetAuthUser(user User)
}

type UserContext interface {
	op_context.Context
	multitenancy.WithTenancy
	WithAuthUser
}

func Tenancy(ctx UserContext) string {
	if ctx.GetTenancy() == nil {
		return ""
	}
	return ctx.GetTenancy().GetID()
}

type UserContextBase struct {
	op_context.ContextBase
	User    User
	Tenancy multitenancy.Tenancy
}

func (u *UserContextBase) AuthUser() User {
	return u.User
}

func (u *UserContextBase) SetAuthUser(user User) {
	u.User = user
}

func (u *UserContextBase) GetTenancy() multitenancy.Tenancy {
	return u.Tenancy
}

func (u *UserContextBase) SetTenancy(tenancy multitenancy.Tenancy) {
	u.Tenancy = tenancy
	u.SetLoggerField("tenancy", tenancy.Name())
	if tenancy.Cache() != nil {
		u.SetCache(tenancy.Cache())
	}
}
