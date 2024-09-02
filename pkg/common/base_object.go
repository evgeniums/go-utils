package common

import (
	"time"

	"github.com/evgeniums/go-utils/pkg/utils"
)

type WithID interface {
	GetID() string
}

type WithIDStub struct {
}

func (w *WithIDStub) GetID() string {
	return ""
}

type ID interface {
	WithID
	SetID(id string)
	GenerateID()
}

type IDBase struct {
	ID string `gorm:"primary_key" json:"id" display:"ID"`
}

func (o *IDBase) GetID() string {
	return o.ID
}

func (o *IDBase) GenerateID() {
	o.ID = utils.GenerateID()
}

func (o *IDBase) SetID(id string) {
	o.ID = id
}

type CreatedAt interface {
	InitCreatedAt()
	GetCreatedAt() time.Time
	SetCreatedAt(t time.Time)
}

type CreatedAtBase struct {
	CREATED_AT time.Time `gorm:"index;autoCreateTime:false" json:"created_at" display:"Created"`
}

func (w *CreatedAtBase) InitCreatedAt() {
	w.CREATED_AT = time.Now().Truncate(time.Microsecond)
}

func (w *CreatedAtBase) GetCreatedAt() time.Time {
	return w.CREATED_AT
}

func (w *CreatedAtBase) SetCreatedAt(t time.Time) {
	w.CREATED_AT = t
}

type UpdatedAt interface {
	SetUpDatedAt(time.Time)
	GetUpdatedAt() time.Time
}

type UpdatedAtBase struct {
	UPDATED_AT time.Time `gorm:"index;autoUpdateTime:false" json:"updated_at" display:"Updated"`
}

func (w *UpdatedAtBase) SetUpDatedAt(t time.Time) {
	w.UPDATED_AT = t
}

func (w *UpdatedAtBase) GetUpdatedAt() time.Time {
	return w.UPDATED_AT
}

type Object interface {
	ID
	CreatedAt
	UpdatedAt
	InitObject()
}

type ObjectBase struct {
	IDBase
	CreatedAtBase
	UpdatedAtBase
}

func (o *ObjectBase) InitObject() {
	o.GenerateID()
	o.InitCreatedAt()
	o.UPDATED_AT = o.CREATED_AT
}

type ObjectWithMonth interface {
	Object
	utils.MonthData
}

type ObjectWithMonthBase struct {
	ObjectBase
	utils.MonthDataBase
}

func (o *ObjectWithMonthBase) InitObject() {
	o.ObjectBase.InitObject()
	month, _ := utils.MonthFromId(o.GetID())
	o.MonthDataBase.SetMonth(month)
}
