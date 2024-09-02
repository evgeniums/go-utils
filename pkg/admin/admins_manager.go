package admin

import (
	"github.com/evgeniums/go-utils/pkg/op_context"
	"github.com/evgeniums/go-utils/pkg/user"
)

type Manager struct {
	*user.UsersWithSessionBase[*Admin, *AdminSession, *AdminSessionClient]
}

type AdminControllers = user.UsersWithSessionBaseConfig[*Admin]

func NewManager(controllers ...AdminControllers) *Manager {
	m := &Manager{UsersWithSessionBase: user.NewUsersWithSession(NewAdmin, NewAdminSession, NewAdminSessionClient, NewOplog, controllers...)}
	return m
}

func (m *Manager) AddAdmin(ctx op_context.Context, login string, password string, phone string) (*Admin, error) {
	c := ctx.TraceInMethod("AddAdmin")
	defer ctx.TraceOutMethod()

	admin, err := m.UsersWithSessionBase.Add(ctx, login, password, user.Phone(phone, &Admin{}))
	return admin, c.SetError(err)
}
