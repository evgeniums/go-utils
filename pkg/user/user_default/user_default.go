package user_default

import "github.com/evgeniums/go-backend-helpers/pkg/user"

type User struct {
	user.UserBase
}

type Users struct {
	user.Users[*User]
}

func NewUser() *User {
	return &User{}
}

func NewUsers() *Users {
	m := &Users{}
	m.MakeUser = NewUser
	return m
}
