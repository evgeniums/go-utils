package tenancy_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

func (t *TenancyClient) Find(ctx op_context.Context, id string, idIsDisplay ...bool) (*multitenancy.TenancyItem, error) {
	return tenancy_manager.FindTenancy(t, ctx, id, idIsDisplay...)
}
