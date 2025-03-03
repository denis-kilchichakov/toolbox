package sqldb

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os" // Add the os package
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunMigrations_Success(t *testing.T) {
	// given
	db, err := InitSqlite(":memory:")
	if err != nil {
		t.Fatalf("InitSqlite failed: %v", err)
	}
	defer db.Close()

	script1 := `
	CREATE TABLE IF NOT EXISTS test_migration_1 (
		a TEXT NOT NULL,
		b INT NOT NULL
	);`

	script2 := `
	INSERT INTO test_migration_1 (a, b) VALUES ('foo', 42);
	`

	path := setupMigrationFiles([]string{script1, script2})
	defer removeTempDir(path)

	expectedA := "foo"
	expectedB := 42

	// when
	db.RunMigrations(path)
	// check that second run doesn't produce more inserts
	db.RunMigrations(path)

	// then
	var a string
	var b int
	err = db.QueryRow("SELECT a, b FROM test_migration_1").Scan(&a, &b)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatal("No rows found in table test_migration_1")
		} else {
			t.Fatalf("Failed to query test_migration_1: %v", err)
		}
	}

	assert.Equal(t, expectedA, a)
	assert.Equal(t, expectedB, b)

	// Optionally, you can also check the count of rows to ensure there's exactly one
	var rowCount int
	err = db.QueryRow("SELECT COUNT(*) FROM test_migration_1").Scan(&rowCount)
	if err != nil {
		t.Fatalf("Failed to count rows in test_migration_1: %v", err)
	}
	assert.Equal(t, 1, rowCount, "test_migration_1 does not contain exactly one row")
}

//go:embed "embed_test/*.sql"
var migrations embed.FS

func TestRunMigrationsFromEmbed_Success(t *testing.T) {
	// given
	db, err := InitSqlite(":memory:")
	if err != nil {
		t.Fatalf("InitSqlite failed: %v", err)
	}
	defer db.Close()

	expectedA := "foo"
	expectedB := 42

	// when
	err = db.RunMigrationsFromEmbed(migrations)
	if err != nil {
		t.Fatalf("RunMigrationsFromEmbed failed: %v", err)
	}
	// check that second run doesn't produce more inserts
	err = db.RunMigrationsFromEmbed(migrations)
	if err != nil {
		t.Fatalf("RunMigrationsFromEmbed failed on second run: %v", err)
	}

	// then
	var a string
	var b int
	err = db.QueryRow("SELECT a, b FROM test_migration_1").Scan(&a, &b)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatal("No rows found in table test_migration_1")
		} else {
			t.Fatalf("Failed to query test_migration_1: %v", err)
		}
	}

	assert.Equal(t, expectedA, a)
	assert.Equal(t, expectedB, b)

	// Optionally, you can also check the count of rows to ensure there's exactly one
	var rowCount int
	err = db.QueryRow("SELECT COUNT(*) FROM test_migration_1").Scan(&rowCount)
	if err != nil {
		t.Fatalf("Failed to count rows in test_migration_1: %v", err)
	}
	assert.Equal(t, 1, rowCount, "test_migration_1 does not contain exactly one row")
}

func setupMigrationFiles(files []string) (path string) {
	path = createTempDir()
	for i, file := range files {
		os.WriteFile(fmt.Sprintf("%s/%d.sql", path, i), []byte(file), 0644)
	}
	return
}

func createTempDir() string {
	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func removeTempDir(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Fatal(err)
	}
}
