package signature

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
)

type MessageSignature struct {
	common.ObjectBase
	auth.WithUserBase
	Context    string `gorm:"uniqueIndex"`
	Operation  string `gorm:"index"`
	Algorithm  string `gorm:"index"`
	Message    string
	Signature  string
	ExtraData  string `gorm:"index"`
	PubKeyHash string `gorm:"index"`
}

func DbModels() []interface{} {
	return []interface{}{&MessageSignature{}}
}
