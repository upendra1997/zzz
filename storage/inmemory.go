package storage

type InMemoryStorage struct {
	data map[Key]Value
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[Key]Value),
	}
}

func (s *InMemoryStorage) Set(key Key, value Value) error {
	s.data[key] = value
	return nil
}

func (s *InMemoryStorage) Get(key Key) (Value, error) {
	value, exists := s.data[key]
	if !exists {
		return "", ErrKeyNotFound
	}
	return value, nil
}

func (s *InMemoryStorage) Delete(key Key) error {
	_, exists := s.data[key]
	if !exists {
		return ErrKeyNotFound
	}
	delete(s.data, key)
	return nil
}
