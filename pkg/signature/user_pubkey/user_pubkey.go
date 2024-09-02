package user_pubkey

import (
	"github.com/evgeniums/go-utils/pkg/auth"
	"github.com/evgeniums/go-utils/pkg/common"
	"github.com/evgeniums/go-utils/pkg/signature"
)

type PubkeyI interface {
	PubKey() string
	PubKeyHash() string
	SetPubKey(key string)
	SetPubKeyHash(hash string)
}

type UserPubkeyI interface {
	common.Object
	common.WithActive
	PubkeyI
	PubKeyOwner() string
	SetPubKeyOwner(hash string)
}

type PubkeyData struct {
	PublicKey string `json:"public_key" validate:"required" vmessage:"Public key must be set in request." display:"Key"`
}

type PubkeyEssentials struct {
	PubkeyData
	PublicKeyHash  string `json:"public_key_hash" gorm:"index;index:,unique,composite:u" display:"Hash"`
	PublicKeyOwner string `json:"public_key_owner" gorm:"index;index:,unique,composite:u" display:"Owner ID"`
}

type UserPubkey struct {
	common.ObjectBase
	common.WithActiveBase
	PubkeyEssentials
}

func (u *UserPubkey) PubKey() string {
	return u.PublicKey
}

func (u *UserPubkey) SetPubKey(key string) {
	u.PublicKey = key
}

func (u *UserPubkey) PubKeyHash() string {
	return u.PublicKeyHash
}

func (u *UserPubkey) SetPubKeyHash(hash string) {
	u.PublicKeyHash = hash
}

func (u *UserPubkey) PubKeyOwner() string {
	return u.PublicKeyOwner
}

func (u *UserPubkey) SetPubKeyOwner(owner string) {
	u.PublicKeyOwner = owner
}

func NewOplog() *signature.OpLogPubKey {
	return &signature.OpLogPubKey{}
}

func FindUserPubKey[T UserPubkeyI](ctrl PubkeyController[T], ctx auth.AuthContext) (signature.UserWithPubkey, error) {

	pubKey, err := ctrl.FindActivePubKey(ctx, ctx.AuthUser().GetID())
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}
