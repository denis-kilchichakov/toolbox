package sqldb

import "testing"

func TestSqliteInMemory(t *testing.T) {
	db, err := InitSqlite(":memory:")
	if err != nil {
		t.Fatalf("InitSqlite failed: %v", err)
	}
	defer db.Close()
}
