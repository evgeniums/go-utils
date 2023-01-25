package common

import (
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type ID interface {
	GetID() string
	SetID(id string)
	GenerateID()
}

type IDBase struct {
	ID string `gorm:"primary_key" json:"id"`
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
	CREATED_AT time.Time `gorm:"index" json:"created_at"`
}

func (w *CreatedAtBase) InitCreatedAt() {
	w.CREATED_AT = time.Now().UTC()
}

func (w *CreatedAtBase) GetCreatedAt() time.Time {
	return w.CREATED_AT
}

func (w *CreatedAtBase) SetCreatedAt(t time.Time) {
	w.CREATED_AT = t
}

type UpdatedAt interface {
	GetUpdatedAt() time.Time
}

type UpdatedAtBase struct {
	UPDATED_AT time.Time `gorm:"index" json:"updated_at"`
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
