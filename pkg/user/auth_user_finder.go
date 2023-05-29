package user

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type AuthUserFinderBase struct {
	crud.WithCRUDBase
	userBuilder func() User
}

func (a *AuthUserFinderBase) FindAuthUser(ctx op_context.Context, login string) (auth.User, error) {
	user := a.userBuilder()
	found, err := FindByLogin(a.CRUD(), ctx, login, user)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return user, nil
}

func NewAuthUserFinder(userBuilder func() User, cruds ...crud.CRUD) *AuthUserFinderBase {
	a := &AuthUserFinderBase{userBuilder: userBuilder}
	a.Construct(cruds...)
	return a
}
