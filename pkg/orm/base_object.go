package orm

import (
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type WithId struct {
	Id string `json:"id" gorm:"primary_key"`
}

func (o *WithId) ID() string {
	return o.Id
}

func (o *WithId) GenerateId() {
	o.Id = utils.GenerateID()
}

type WithIdInterface interface {
	ID() string
	GenerateId()
}

type WithCreatedAt struct {
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

func (w *WithCreatedAt) InitCreatedAt() {
	w.CreatedAt = time.Now().UTC()
}

func (w *WithCreatedAt) GetCreatedAt() time.Time {
	return w.CreatedAt
}

func (w *WithCreatedAt) SetCreatedAt(t time.Time) {
	w.CreatedAt = t
}

type WithCreatedAtInterface interface {
	InitCreatedAt()
	GetCreatedAt() time.Time
	SetCreatedAt(t time.Time)
}

type BaseInterface interface {
	WithIdInterface
	WithCreatedAtInterface
	Init()
}

type BaseObject struct {
	WithId
	WithCreatedAt
}

func (o *BaseObject) Init() {
	o.GenerateId()
	o.InitCreatedAt()
}
