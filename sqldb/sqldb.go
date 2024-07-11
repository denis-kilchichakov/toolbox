package sqldb

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type SqlDb struct {
	*sql.DB
}

func InitSqlite(dbPath string) (*SqlDb, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &SqlDb{
		db,
	}, nil
}
