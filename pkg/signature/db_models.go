package signature

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type MessageSignature struct {
	common.ObjectBase
	utils.MonthDataBase
	auth.WithUserBase
	Context    string `gorm:"uniqueIndex"`
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
