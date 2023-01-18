package auth_token

import (
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/crypt_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type AuthTokenSession struct {
	common.ObjectBase
	UpdatedAt   time.Time `gorm:"index"`
	UserId      string    `gorm:"index"`
	UserDisplay string    `gorm:"index"`
	UserLogin   string    `gorm:"index"`
	Valid       bool      `gorm:"index"`
	Expiration  time.Time `gorm:"index"`
}

type AuthTokenClient struct {
	common.ObjectBase
	UpdatedAt   time.Time `gorm:"index"`
	SessionId   string    `gorm:"index;index:,unique,composite:client_session"`
	UserId      string    `gorm:"index"`
	UserDisplay string    `gorm:"index"`
	ClientIp    string    `gorm:"index"`
	UserAgent   string
	ClientHash  string `gorm:"index;index:,unique,composite:client_session"`
}

func CreateSession(ctx auth.AuthContext, expiration time.Time) (*AuthTokenSession, error) {

	c := ctx.TraceInMethod("auth_token.CreateSession")
	defer ctx.TraceOutMethod()

	session := &AuthTokenSession{}
	session.InitObject()
	session.UpdatedAt = session.GetCreatedAt()
	session.UserId = ctx.AuthUser().GetID()
	session.UserDisplay = ctx.AuthUser().Display()
	session.UserLogin = ctx.AuthUser().Login()
	session.Valid = true
	session.Expiration = expiration

	err := ctx.DB().Create(ctx, session)
	if err != nil {
		return nil, c.SetError(err)
	}

	return session, nil
}

func FindSession(ctx op_context.Context, sessionId string) (*AuthTokenSession, error) {

	c := ctx.TraceInMethod("auth_token.FindSession")
	defer ctx.TraceOutMethod()

	session := &AuthTokenSession{}
	_, err := ctx.DB().FindByField(ctx, "id", sessionId, session)
	if err != nil {
		return nil, c.SetError(err)
	}

	return session, nil
}

func UpdateSessionClient(ctx auth.AuthContext) error {

	// setup
	c := ctx.TraceInMethod("auth_token.UpdateSessionClient")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// extract client parameters
	clientIp := ctx.GetRequestClientIp()
	userAgent := ctx.GetRequestUserAgent()
	h := crypt_utils.NewHash()
	clientHash := h.CalcStrStr(clientIp, userAgent)

	// find client in database
	tryUpdate := true
	client := &AuthTokenClient{}
	fields := db.Fields{"session_id": ctx.AuthUser().GetSessionId(), "client_hash": clientHash}
	notfound, err := ctx.DB().FindByFields(ctx, fields, client)
	if !db.CheckFoundNoError(notfound, &err) {
		if err != nil {
			c.SetMessage("failed to find client in database")
			return err
		}

		// create new client
		tryUpdate = false
		client.InitObject()
		client.ClientIp = clientIp
		client.ClientHash = clientHash
		client.SessionId = ctx.AuthUser().GetSessionId()
		client.UpdatedAt = client.GetCreatedAt()
		client.UserId = ctx.AuthUser().GetID()
		client.UserDisplay = ctx.AuthUser().Display()
		client.UserAgent = userAgent
		err1 := ctx.DB().Create(ctx, client)
		if err1 != nil {
			c.Logger().Error("failed to create session client in database", err1)
			tryUpdate = true
		}
	}

	// update client
	if tryUpdate {
		err = ctx.DB().Update(ctx, client, db.Fields{"updated_at": time.Now()})
		if err != nil {
			c.SetMessage("failed to update client in database")
			return err
		}
	}

	ctx.AuthUser().SetClientId(client.GetID())
	ctx.SetLoggerField("client", client.GetID())
	return nil
}

func UpdateSessionExpiration(ctx auth.AuthContext, session *AuthTokenSession) error {

	c := ctx.TraceInMethod("auth_token.UpdateSessionExpiration")
	defer ctx.TraceOutMethod()

	err := ctx.DB().Update(ctx, session, db.Fields{"expiration": session.Expiration})
	if err != nil {
		return c.SetError(err)
	}
	return nil
}

func InvalidateSession(ctx op_context.Context, userId string, sessionId string) error {

	c := ctx.TraceInMethod("auth_token.InvalidateSession")
	defer ctx.TraceOutMethod()

	err := ctx.DB().UpdateWithFilter(ctx, &AuthTokenSession{}, db.Fields{"id": sessionId, "user_id": userId}, db.Fields{"valid": false, "updated_at": time.Now()})
	if err != nil {
		return c.SetError(err)
	}
	return nil

}

func InvalidateUserSessions(ctx op_context.Context, userId string) error {
	c := ctx.TraceInMethod("auth_token.InvalidateUserSessions")
	defer ctx.TraceOutMethod()

	err := ctx.DB().UpdateWithFilter(ctx, &AuthTokenSession{}, db.Fields{"user_id": userId}, db.Fields{"valid": false, "updated_at": time.Now()})
	if err != nil {
		return c.SetError(err)
	}
	return nil
}

func InvalidateAllSessions(ctx op_context.Context) error {
	c := ctx.TraceInMethod("auth_token.InvalidateAllSessions")
	defer ctx.TraceOutMethod()

	err := ctx.DB().UpdateAll(ctx, &AuthTokenSession{}, db.Fields{"valid": false, "updated_at": time.Now()})
	if err != nil {
		return c.SetError(err)
	}
	return nil
}
