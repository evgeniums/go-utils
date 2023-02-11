package oplog

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
)

type Oplog interface {
	common.Object
	Operation() string
	Context() string
	ContextName() string
	OriginType() string
	OriginName() string
	OriginSource() string
}

type OplogHolder struct {
	Operation    string `gorm:"index" json:"operation"`
	Context      string `gorm:"index" json:"context"`
	ContextName  string `gorm:"index" json:"context_name"`
	OriginType   string `gorm:"index" json:"origin_type"`
	OriginName   string `gorm:"index" json:"origin_name"`
	OriginSource string `gorm:"index" json:"origin_source"`
}

type OplogBase struct {
	common.ObjectBase
	OplogHolder
}

func (o *OplogBase) Operation() string {
	return o.OplogHolder.Operation
}

func (o *OplogBase) Context() string {
	return o.OplogHolder.Context
}

func (o *OplogBase) ContextName() string {
	return o.OplogHolder.ContextName
}

func (o *OplogBase) OriginType() string {
	return o.OplogHolder.OriginType
}

func (o *OplogBase) OriginName() string {
	return o.OplogHolder.OriginName
}

func (o *OplogBase) OriginSource() string {
	return o.OplogHolder.OriginSource
}

type OplogController interface {
	Write(o Oplog) error
	Read(filter *db.Filter, docs interface{}) error
}
