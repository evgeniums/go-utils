package user_manager

import (
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/crypt_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type Session interface {
	common.Object

	SetUserId(userId string)
	SetUserDisplay(userDisplay string)
	SetUserLogin(userLogin string)
	SetValid(valid bool)
	SetExpiration(exp time.Time)

	GetUserId() string
	GetUserDisplay() string
	GetUserLogin() string
	IsValid() bool
	GetExpiration() time.Time
}

type SessionBase struct {
	common.ObjectBase
	UserId      string    `gorm:"index"`
	UserDisplay string    `gorm:"index"`
	UserLogin   string    `gorm:"index"`
	Valid       bool      `gorm:"index"`
	Expiration  time.Time `gorm:"index"`
}

func (s *SessionBase) SetUserId(userId string) {
	s.UserId = userId
}

func (s *SessionBase) GetUserId() string {
	return s.UserId
}

func (s *SessionBase) SetUserDisplay(userDisplay string) {
	s.UserDisplay = userDisplay
}

func (s *SessionBase) GetUserDisplay() string {
	return s.UserDisplay
}

func (s *SessionBase) SetUserLogin(userLogin string) {
	s.UserLogin = userLogin
}

func (s *SessionBase) GetUserLogin() string {
	return s.UserLogin
}

func (s *SessionBase) SetValid(valid bool) {
	s.Valid = valid
}

func (s *SessionBase) IsValid() bool {
	return s.Valid
}

func (s *SessionBase) SetExpiration(exp time.Time) {
	s.Expiration = exp
}

func (s *SessionBase) GetExpiration() time.Time {
	return s.Expiration
}

type SessionClient interface {
	common.Object
	SetSessionId(userId string)
	SetUserId(userId string)
	SetUserDisplay(userDisplay string)
	SetClientIp(clientIp string)
	SetUserAgent(userAgent string)
	SetClientHash(clientHash string)

	GetSessionId() string
	GetUserId() string
	GetUserDisplay() string
	GetClientIp() string
	GetUserAgent() string
	GetClientHash() string
}

type SessionClientBase struct {
	common.ObjectBase
	SessionId   string `gorm:"index;index:,unique,composite:client_session"`
	UserId      string `gorm:"index"`
	UserDisplay string `gorm:"index"`
	ClientIp    string `gorm:"index"`
	UserAgent   string
	ClientHash  string `gorm:"index;index:,unique,composite:client_session"`
}

func (s *SessionClientBase) SetSessionId(sessionId string) {
	s.SessionId = sessionId
}

func (s *SessionClientBase) GetSessionId() string {
	return s.SessionId
}

func (s *SessionClientBase) SetUserId(val string) {
	s.UserId = val
}

func (s *SessionClientBase) GetUserId() string {
	return s.UserId
}

func (s *SessionClientBase) SetUserDisplay(val string) {
	s.UserDisplay = val
}

func (s *SessionClientBase) GetUserDisplay() string {
	return s.UserDisplay
}

func (s *SessionClientBase) SetClientIp(val string) {
	s.ClientIp = val
}

func (s *SessionClientBase) GetClientIp() string {
	return s.ClientIp
}

func (s *SessionClientBase) SetUserAgent(val string) {
	s.UserAgent = val
}

func (s *SessionClientBase) GetUserAgent() string {
	return s.UserAgent
}

func (s *SessionClientBase) SetClientHash(val string) {
	s.ClientHash = val
}

func (s *SessionClientBase) GetClientHash() string {
	return s.ClientHash
}

type SessionManager interface {
	CreateSession(ctx auth.AuthContext, expiration time.Time) (Session, error)
	FindSession(ctx op_context.Context, sessionId string) (Session, error)
	UpdateSessionClient(ctx auth.AuthContext) error
	UpdateSessionExpiration(ctx auth.AuthContext, session Session) error
	InvalidateSession(ctx op_context.Context, userId string, sessionId string) error
	InvalidateUserSessions(ctx op_context.Context, userId string) error
	InvalidateAllSessions(ctx op_context.Context) error
}

type WithSessionManager interface {
	WithUserManager
	SessionManager() SessionManager
}

type SessionManagerBase struct {
	MakeSession       func() Session
	MakeSessionClient func() SessionClient
}

func NewSessionManager(MakeSession func() Session, MakeSessionClient func() SessionClient) *SessionManagerBase {
	return &SessionManagerBase{MakeSession, MakeSessionClient}
}

func (s *SessionManagerBase) CreateSession(ctx auth.AuthContext, expiration time.Time) (Session, error) {

	c := ctx.TraceInMethod("auth_token.CreateSession")
	defer ctx.TraceOutMethod()

	session := s.MakeSession()
	session.InitObject()
	session.SetUserId(ctx.AuthUser().GetID())
	session.SetUserDisplay(ctx.AuthUser().Display())
	session.SetUserLogin(ctx.AuthUser().Login())
	session.SetValid(true)
	session.SetExpiration(expiration)

	err := ctx.DB().Create(ctx, session)
	if err != nil {
		return nil, c.SetError(err)
	}

	return session, nil
}

func (s *SessionManagerBase) FindSession(ctx op_context.Context, sessionId string) (Session, error) {

	c := ctx.TraceInMethod("auth_token.FindSession")
	defer ctx.TraceOutMethod()

	session := s.MakeSession()
	_, err := ctx.DB().FindByField(ctx, "id", sessionId, session)
	if err != nil {
		return nil, c.SetError(err)
	}

	return session, nil
}

func (s *SessionManagerBase) UpdateSessionClient(ctx auth.AuthContext) error {

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
	client := s.MakeSessionClient()
	fields := db.Fields{"session_id": ctx.GetSessionId(), "client_hash": clientHash}
	notfound, err := ctx.DB().FindByFields(ctx, fields, client)
	if !db.CheckFoundNoError(notfound, &err) {
		if err != nil {
			c.SetMessage("failed to find client in database")
			return err
		}

		// create new client
		tryUpdate = false
		client.InitObject()
		client.SetClientIp(clientIp)
		client.SetClientHash(clientHash)
		client.SetSessionId(ctx.GetSessionId())
		client.SetUserId(ctx.AuthUser().GetID())
		client.SetUserDisplay(ctx.AuthUser().Display())
		client.SetUserAgent(userAgent)
		err1 := ctx.DB().Create(ctx, client)
		if err1 != nil {
			c.Logger().Error("failed to create session client in database", err1)
			tryUpdate = true
		}
	}

	// update client
	if tryUpdate {
		err = db.Update(ctx.DB(), ctx, client, db.Fields{"updated_at": time.Now()})
		if err != nil {
			c.SetMessage("failed to update client in database")
			return err
		}
	}

	ctx.SetClientId(client.GetID())
	ctx.SetLoggerField("client", client.GetID())
	return nil
}

func (s *SessionManagerBase) UpdateSessionExpiration(ctx auth.AuthContext, session Session) error {

	c := ctx.TraceInMethod("auth_token.UpdateSessionExpiration")
	defer ctx.TraceOutMethod()

	err := db.Update(ctx.DB(), ctx, session, db.Fields{"expiration": session.GetExpiration()})
	if err != nil {
		return c.SetError(err)
	}
	return nil
}

func (s *SessionManagerBase) InvalidateSession(ctx op_context.Context, userId string, sessionId string) error {

	c := ctx.TraceInMethod("auth_token.InvalidateSession")
	defer ctx.TraceOutMethod()

	err := ctx.DB().Update(ctx, s.MakeSession(), db.Fields{"id": sessionId, "user_id": userId}, db.Fields{"valid": false, "updated_at": time.Now()})
	if err != nil {
		return c.SetError(err)
	}
	return nil

}

func (s *SessionManagerBase) InvalidateUserSessions(ctx op_context.Context, userId string) error {
	c := ctx.TraceInMethod("auth_token.InvalidateUserSessions")
	defer ctx.TraceOutMethod()

	err := ctx.DB().Update(ctx, s.MakeSession(), db.Fields{"user_id": userId}, db.Fields{"valid": false, "updated_at": time.Now()})
	if err != nil {
		return c.SetError(err)
	}
	return nil
}

func (s *SessionManagerBase) InvalidateAllSessions(ctx op_context.Context) error {
	c := ctx.TraceInMethod("auth_token.InvalidateAllSessions")
	defer ctx.TraceOutMethod()

	err := ctx.DB().UpdateAll(ctx, s.MakeSession(), db.Fields{"valid": false, "updated_at": time.Now()})
	if err != nil {
		return c.SetError(err)
	}
	return nil
}
