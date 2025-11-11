package storage

import "errors"

type Key = uint64
type Value = string

var ErrKeyNotFound = errors.New("key not found")

type Storage interface {
	Set(key Key, value Value) error
	Get(key Key) (Value, error)
	Delete(key Key) error
}
