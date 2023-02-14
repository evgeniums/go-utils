package db

import (
	"sync"

	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type JoinQuery interface {
	Join(ctx logger.WithLogger, filter *Filter, dest interface{}) (int64, error)
}

type Joiner interface {
	Join(model interface{}, field string, fieldsModel ...interface{}) JoinBegin
}

type JoinBegin interface {
	On(model interface{}, field string, fieldsModel ...interface{}) JoinEnd
}

type JoinEnd interface {
	Joiner
	Destination(dst interface{}) (JoinQuery, error)
}

type JoinTableData struct {
	Model       interface{}
	FieldsModel interface{}
}

type JoinTableBase struct {
	JoinTableData
}

func (j *JoinTableBase) Model() interface{} {
	return j.Model
}

func (j *JoinTableBase) FieldsModel() interface{} {
	return j.FieldsModel
}

type JoinPairData struct {
	LeftField  string
	RightField string
}

type JoinPairBase struct {
	JoinPairData
}

func (j *JoinPairBase) LeftField() string {
	return j.JoinPairData.LeftField
}

func (j *JoinPairBase) RightField() string {
	return j.JoinPairData.RightField
}

type JoinQueryData struct {
	Destination interface{}
}

type JoinQueryBase struct {
	JoinQueryData
}

func (j *JoinQueryBase) Destination() interface{} {
	return j.JoinQueryData.Destination
}

func (j *JoinQueryBase) SetDestination(destination interface{}) {
	j.JoinQueryData.Destination = destination
}

type JoinQueryBuilder = func() JoinQuery

type JoinQueryConfig struct {
	Builder JoinQueryBuilder
	Name    string
	Nocache bool
}

func NewJoin(builder JoinQueryBuilder, name string, nocache ...bool) *JoinQueryConfig {
	return &JoinQueryConfig{Builder: builder, Name: name, Nocache: utils.OptionalArg(false, nocache...)}
}

type JoinQueries struct {
	mutex sync.Mutex
	cache map[string]JoinQuery
}

func NewJoinQueries() *JoinQueries {
	j := &JoinQueries{}
	j.cache = make(map[string]JoinQuery)
	return j
}

func (j *JoinQueries) FindOrCreate(config *JoinQueryConfig) JoinQuery {
	j.mutex.Lock()
	q, ok := j.cache[config.Name]
	j.mutex.Unlock()
	if !ok || config.Nocache {
		q = config.Builder()
	}
	if !config.Nocache {
		j.mutex.Lock()
		j.cache[config.Name] = q
		j.mutex.Unlock()
	}
	return q
}
