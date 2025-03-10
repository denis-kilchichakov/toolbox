package sqldb

import (
	"crypto/md5"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const migrationsInitialScript = `
CREATE TABLE IF NOT EXISTS migrations (
    file TEXT NOT NULL,
    md5 TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (md5)
);
`

func (db *SqlDb) RunMigrationsFromEmbed(embedFS embed.FS) error {
	log.Println("Running migrations from embedded files")
	rootFiles, err := embedFS.ReadDir(".")
	if err != nil {
		return err
	}

	var sqlFiles []string
	for _, rootFile := range rootFiles {
		if rootFile.IsDir() {
			dirFiles, err := embedFS.ReadDir(rootFile.Name())
			if err != nil {
				return err
			}
			for _, file := range dirFiles {
				if !file.IsDir() && filepath.Ext(file.Name()) == ".sql" {
					sqlFiles = append(sqlFiles, filepath.Join(rootFile.Name(), file.Name()))
				}
			}
			break
		}
	}

	sort.Strings(sqlFiles)

	return db.runMigrationsCommon(sqlFiles, embedFS.ReadFile)
}

func (db *SqlDb) RunMigrations(migrationsPath string) error {
	log.Println("Running migrations from: ", migrationsPath)
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.sql"))
	if err != nil {
		return err
	}

	sort.Strings(files)

	return db.runMigrationsCommon(files, os.ReadFile)
}

func (db *SqlDb) runMigrationsCommon(files []string, readFileFunc func(string) ([]byte, error)) error {
	db.applyMigration(migrationsInitialScript)

	for _, file := range files {
		fileName := filepath.Base(file)
		contents, err := readFileFunc(file)
		if err != nil {
			return err
		}
		log.Println("Migration applying: ", file)
		nowMd5 := fmt.Sprintf("%x", md5.Sum(contents))
		applied, err := db.checkIfMigrationPreviouslyApplied(nowMd5)
		if err != nil {
			return err
		}
		if !applied {
			err = db.applyMigration(string(contents))
			if err != nil {
				return err
			}
			err = db.saveMigrationInfo(fileName, nowMd5)
			if err != nil {
				return err
			}
		} else {
			log.Println("Migration already applied: ", file)
			continue
		}
		log.Println("Migration applied: ", file)
	}

	return nil
}

func (db *SqlDb) applyMigration(migration string) error {
	_, err := db.Exec(migration)
	if err != nil {
		log.Println("Error applying migration: ", migration)
		return err
	}

	return nil
}

func (db *SqlDb) checkIfMigrationPreviouslyApplied(nowMd5 string) (bool, error) {
	row := db.QueryRow("SELECT file FROM migrations WHERE md5 = $1", nowMd5)
	var file string
	err := row.Scan(&file)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (db *SqlDb) saveMigrationInfo(file string, md5 string) error {
	_, err := db.Exec("INSERT INTO migrations (file, md5, applied_at) VALUES ($1, $2, $3)", file, md5, time.Now())
	if err != nil {
		return err
	}

	return nil
}
