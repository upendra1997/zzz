package storage

import (
	"database/sql"
	"log/slog"

	_ "github.com/glebarez/go-sqlite"
)

type SqliteStorage struct {
	*sql.DB
}

type SqliteStorageTransaction struct {
	*sql.Tx
	db *SqliteStorage
}

func NewSqliteStorage(filePath string) *SqliteStorage {
	if filePath == "" {
		filePath = ":memory:"
	}
	slog.Error(filePath)
	db, err := sql.Open("sqlite", filePath)
	if err != nil {
		slog.Error("Cannot create sqlite DB", "error", err)
		return nil
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS kv_store (key INTEGER PRIMARY KEY, value TEXT);`)
	return &SqliteStorage{db}
}

func (db *SqliteStorage) Get(key Key) (Value, error) {
	var value Value
	err := db.QueryRow(`SELECT value FROM kv_store WHERE key = ?;`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", ErrKeyNotFound
	}
	return value, err
}

func (db *SqliteStorage) Begin() StorageTransaction {
	tx, err := db.DB.Begin()
	if err != nil {
		slog.Error("Cannot begin transaction", "error", err)
		return nil
	}
	return &SqliteStorageTransaction{tx, db}
}

func (tx *SqliteStorageTransaction) Set(key Key, value Value) error {
	_, err := tx.Exec(`INSERT OR REPLACE INTO kv_store (key, value) VALUES (?, ?);`, key, value)
	return err
}

func (tx *SqliteStorageTransaction) Delete(key Key) error {
	result, err := tx.Exec(`DELETE FROM kv_store WHERE key = ?;`, key)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrKeyNotFound
	}
	return nil
}

func (tx *SqliteStorageTransaction) Get(key Key) (Value, error) {
	var value Value
	err := tx.QueryRow(`SELECT value FROM kv_store WHERE key = ?;`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", ErrKeyNotFound
	}
	return value, err
}
