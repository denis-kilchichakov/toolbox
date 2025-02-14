package sqldb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqliteInMemory(t *testing.T) {
	db, err := InitSqlite(":memory:")
	require.NoError(t, err, "InitSqlite failed")
	defer db.Close()

	assert.NotNil(t, db, "Database should not be nil")
}
