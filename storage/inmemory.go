package storage

import "sync"

type InMemoryStorage struct {
	data map[Key]Value
}

type InMemoryStorageTransaction struct {
	*InMemoryStorage
	lock         sync.RWMutex
	transactions map[Key]*Value
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[Key]Value),
	}
}

func (store *InMemoryStorage) Get(key Key) (Value, error) {
	value, exists := store.data[key]
	if !exists {
		return "", ErrKeyNotFound
	}
	return value, nil
}

func (store *InMemoryStorage) Begin() StorageTransaction {
	return &InMemoryStorageTransaction{
		InMemoryStorage: store,
		transactions:    make(map[Key]*Value),
	}
}

func (tx *InMemoryStorageTransaction) Set(key Key, value Value) error {
	tx.lock.Lock()
	defer tx.lock.Unlock()
	tx.transactions[key] = &value
	return nil
}

func (tx *InMemoryStorageTransaction) Get(key Key) (Value, error) {
	tx.lock.RLock()
	defer tx.lock.RUnlock()
	value, exists := tx.transactions[key]
	if !exists || value == nil {
		originalValue, ok := tx.data[key]
		if !ok {
			return "", ErrKeyNotFound
		}
		return originalValue, nil
	}
	return *value, nil
}

func (tx *InMemoryStorageTransaction) Delete(key Key) error {
	_, exists := tx.data[key]
	if !exists {
		return ErrKeyNotFound
	}
	tx.transactions[key] = nil
	return nil
}

func (tx *InMemoryStorageTransaction) Commit() error {
	tx.lock.Lock()
	defer tx.lock.Unlock()
	for key, value := range tx.transactions {
		if value == nil {
			delete(tx.data, key)
		} else {
			tx.data[key] = *value
		}
	}
	return nil
}

func (tx *InMemoryStorageTransaction) Rollback() error {
	clear(tx.transactions)
	return nil
}
