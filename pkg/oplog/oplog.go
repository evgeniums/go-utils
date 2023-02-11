package oplog

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
)

type Oplog interface {
	common.Object
	Operation() string
	SetOperation(val string)
	Context() string
	SetContext(val string)
	ContextName() string
	SetContextName(val string)
	OriginType() string
	SetOriginType(val string)
	OriginName() string
	SetOriginName(val string)
	OriginSource() string
	SetOriginSource(val string)
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

func (o *OplogBase) SetOperation(val string) {
	o.OplogHolder.Operation = val
}

func (o *OplogBase) Context() string {
	return o.OplogHolder.Context
}

func (o *OplogBase) SetContext(val string) {
	o.OplogHolder.Context = val
}

func (o *OplogBase) ContextName() string {
	return o.OplogHolder.ContextName
}

func (o *OplogBase) SetContextName(val string) {
	o.OplogHolder.ContextName = val
}

func (o *OplogBase) OriginType() string {
	return o.OplogHolder.OriginType
}

func (o *OplogBase) SetOriginType(val string) {
	o.OplogHolder.OriginType = val
}

func (o *OplogBase) OriginName() string {
	return o.OplogHolder.OriginName
}

func (o *OplogBase) SetOriginName(val string) {
	o.OplogHolder.OriginName = val
}

func (o *OplogBase) OriginSource() string {
	return o.OplogHolder.OriginSource
}

func (o *OplogBase) SetOriginSource(val string) {
	o.OplogHolder.OriginSource = val
}

type OplogController interface {
	Write(o Oplog) error
	Read(filter *db.Filter, docs interface{}) error
}
