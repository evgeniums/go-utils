package signature

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
)

type MessageSignature struct {
	common.ObjectWithMonth
	auth.WithUserBase
	Context    string `gorm:"index;index:,unique,composite:u_month"`
	Operation  string `gorm:"index"`
	Algorithm  string `gorm:"index"`
	Message    string
	Signature  string
	ExtraData  string `gorm:"index"`
	PubKeyHash string `gorm:"index"`
}

type OpLogPubKey struct {
	user.OpLogUser
	KeyId   string `gorm:"index" json:"key_id"`
	KeyHash string `gorm:"index" json:"key_hash"`
}

func DbModels() []interface{} {
	return []interface{}{&MessageSignature{}, &OpLogPubKey{}}
}
