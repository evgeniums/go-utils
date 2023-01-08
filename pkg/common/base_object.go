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
	id string `gorm:"primary_key"`
}

func (o *IDBase) GetID() string {
	return o.id
}

func (o *IDBase) GenerateID() {
	o.id = utils.GenerateID()
}

func (o *IDBase) SetID(id string) {
	o.id = id
}

type CreatedAt interface {
	InitCreatedAt()
	GetCreatedAt() time.Time
	SetCreatedAt(t time.Time)
}

type CreatedAtBase struct {
	created_at time.Time `gorm:"index"`
}

func (w *CreatedAtBase) InitCreatedAt() {
	w.created_at = time.Now().UTC()
}

func (w *CreatedAtBase) GetCreatedAt() time.Time {
	return w.created_at
}

func (w *CreatedAtBase) SetCreatedAt(t time.Time) {
	w.created_at = t
}

type Object interface {
	ID
	CreatedAt
	InitObject()
}

type ObjectBase struct {
	IDBase
	CreatedAtBase
}

func (o *ObjectBase) InitObject() {
	o.GenerateID()
	o.InitCreatedAt()
}
