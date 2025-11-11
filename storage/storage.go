package storage

import "errors"

type Key = uint64
type Value = string

var ErrKeyNotFound = errors.New("key not found")

type Storage interface {
	Get(key Key) (Value, error)
	Begin() StorageTransaction
}

type StorageTransaction interface {
	Commit() error
	Rollback() error
	Set(key Key, value Value) error
	Get(key Key) (Value, error)
	Delete(key Key) error
}
