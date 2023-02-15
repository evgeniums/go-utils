package db_gorm

import (
	"errors"
	"fmt"
	"sync"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"gorm.io/gorm/schema"
)

type FieldDescriptor struct {
	Json       string
	DbTable    string
	DbField    string
	FullDbName string
	Schema     *schema.Field
}

type ModelDescriptor struct {
	Sample     interface{}
	Schema     *schema.Schema
	FieldsJson map[string]*FieldDescriptor
}

func (d *ModelDescriptor) FindJsonField(json string) (*FieldDescriptor, error) {
	f, ok := d.FieldsJson[json]
	if !ok {
		return nil, errors.New("field not found")
	}
	return f, nil
}

func (d *ModelDescriptor) FieldsReady() bool {
	return d.FieldsJson != nil
}

type ModelStore struct {
	mutex       sync.Mutex
	descriptors map[string]*ModelDescriptor
	schemaCache *sync.Map
	schemaNamer schema.Namer
}

var GlobalModelStore *ModelStore

func NewModelStore(global bool) *ModelStore {
	m := &ModelStore{}
	m.descriptors = make(map[string]*ModelDescriptor)
	m.schemaCache = &sync.Map{}
	m.schemaNamer = &schema.NamingStrategy{}
	if global {
		GlobalModelStore = m
		db.SetGlobalModelStore(m)
	}
	return m
}

func (m *ModelStore) RegisterModel(model interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	ms := &ModelDescriptor{Sample: model}
	var err error
	ms.Schema, err = schema.Parse(model, m.schemaCache, m.schemaNamer)
	if err != nil {
		panic(fmt.Sprintf("invalid model: %s", err))
	}
	// ms.FieldsJson = make(map[string]*FieldDescriptor)
}

func (m *ModelStore) FindModel(name string) interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	d, ok := m.descriptors[name]
	if !ok {
		return nil
	}
	return d.Sample
}

func (m *ModelStore) AllModels() []interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	models := make([]interface{}, len(m.descriptors))
	i := 0
	for _, descriptor := range m.descriptors {
		models[i] = descriptor.Sample
		i++
	}
	return models
}

func (m *ModelStore) FindDescriptor(name string) *ModelDescriptor {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	d, ok := m.descriptors[name]
	if !ok {
		return nil
	}
	return d
}

func (m *ModelStore) ParseModelFields(descriptor *ModelDescriptor) error {

	// TODO parse fields

	return nil
}
