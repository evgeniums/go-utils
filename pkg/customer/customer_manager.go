package customer

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_session"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
)

type Manager struct {
	CustomersBase
	auth_session.SessionController
}

type CustomersBase struct {
	user.UsersValidator
	CustomerController
}

func (c *CustomersBase) Construct(customerController CustomerController) {
	c.CustomerController = customerController
}

func (m *CustomersBase) AuthUserManager() auth_session.AuthUserManager {
	return m
}

type ManagerConfig struct {
	CustomerController CustomerController
	SessionController  auth_session.SessionController
}

func NewManager(config ...ManagerConfig) *Manager {

	m := &Manager{}

	if len(config) != 0 {
		m.CustomersBase.Construct(config[0].CustomerController)
		m.SessionController = config[0].SessionController
	}
	if m.CustomerController == nil {
		c := LocalCustomerController()
		m.CustomersBase.Construct(c)
		c.SetUserValidators(m)
	}
	if m.SessionController == nil {
		m.SessionController = auth_session.LocalSessionController()
	}

	m.SetUserBuilder(NewCustomer)
	m.SetSessionBuilder(func() auth_session.Session { return NewCustomerSession() })
	m.SetSessionClientBuilder(func() auth_session.SessionClient { return NewCustomerSessionClient() })
	m.SetOplogBuilder(NewOplog)

	return m
}
