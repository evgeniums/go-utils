package auth

import (
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
)

type User interface {
	GetID() string
	Display() string
	Login() string
	IsBlocked() bool
}

type UserBase struct {
	UserId      string `gorm:"index"`
	UserLogin   string `gorm:"index"`
	UserDisplay string `gorm:"index"`
	UserBlocked bool   `gorm:"index"`
}

func (u *UserBase) GetID() string {
	return u.UserId
}

func (u *UserBase) Display() string {
	return u.UserDisplay
}

func (u *UserBase) Login() string {
	return u.UserLogin
}

func (u *UserBase) IsBlocked() bool {
	return u.UserBlocked
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
	IsLoggedIn() bool
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

func (u *SessionBase) IsLoggedIn() bool {
	return u.session != ""
}

type WithAuthUser interface {
	AuthUser() User
	SetAuthUser(user User)
}

type UserContext interface {
	multitenancy.TenancyContext
	WithAuthUser
}

func Tenancy(ctx UserContext) string {
	return multitenancy.ContextTenancy(ctx)
}

type UserContextBase struct {
	multitenancy.TenancyContextBase
	User User
}

func (u *UserContextBase) AuthUser() User {
	return u.User
}

func (u *UserContextBase) SetAuthUser(user User) {
	u.User = user
}
