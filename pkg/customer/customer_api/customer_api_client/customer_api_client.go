package customer_api_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api/user_client"
)

type CustomerClient struct {
	*user_client.UserClient[*customer.Customer]
}

func NewCustomerClient(client api_client.Client) *CustomerClient {
	c := &CustomerClient{}
	c.UserClient = user_client.NewUserClient(client, customer.NewCustomer, "customer")
	return c
}
