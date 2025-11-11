package storage

import (
	"database/sql"
	"log/slog"
)

type SqliteStorage struct {
	*sql.DB
}

func NewSqliteStorage() (*SqliteStorage, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		slog.Error("Cannot create sqlite DB", "error", err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS kv_store (key INTEGER PRIMARY KEY, value TEXT);`)
	return &SqliteStorage{db}, nil
}

func (s *SqliteStorage) Set(key Key, value Value) error {
	_, err := s.Exec(`INSERT OR REPLACE INTO kv_store (key, value) VALUES (?, ?);`, key, value)
	return err
}

func (s *SqliteStorage) Get(key Key) (Value, error) {
	var value Value
	err := s.QueryRow(`SELECT value FROM kv_store WHERE key = ?;`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", ErrKeyNotFound
	}
	return value, err
}

func (s *SqliteStorage) Delete(key Key) error {
	result, err := s.Exec(`DELETE FROM kv_store WHERE key = ?;`, key)
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
