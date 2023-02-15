package db

type ModelStore interface {
	RegisterModel(model interface{})
	FindModel(name string) interface{}
	AllModels() []interface{}
}

var globalModelStore ModelStore

func SetGlobalModelStore(m ModelStore) {
	globalModelStore = m
}

func GlobalModelStore() ModelStore {
	return globalModelStore
}
