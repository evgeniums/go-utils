package db_test

import (
	"sync"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/db/db_gorm"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"
)

func TestModelDescriptor(t *testing.T) {

	cache := &sync.Map{}
	namer := &schema.NamingStrategy{}

	descr := db_gorm.NewModelDescriptor(&pool.PoolServiceBinding{}, cache, namer)
	require.NotNil(t, descr)

	err := descr.ParseFields()
	assert.NoError(t, err)
}
