package transpiler

type stack[T any] []T

func (s *stack[T]) push(item T) {
	*s = append(*s, item)
}

func (s *stack[T]) pop() (item T) {
	if len(*s) == 0 {
		return
	}
	item = (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return item
}

func (s *stack[T]) peek() (item T) {
	if len(*s) == 0 {
		return
	}
	return (*s)[len(*s)-1]
}

type OrderedMap[T any] struct {
	keys   []string
	values map[string]T
}

func NewOrderedMap[T any]() *OrderedMap[T] {
	return &OrderedMap[T]{
		keys:   []string{},
		values: map[string]T{},
	}
}

func (m *OrderedMap[T]) Len() int {
	return len(m.keys)
}

func (m *OrderedMap[T]) Set(key string, value T) {
	if _, ok := m.values[key]; ok {
		m.values[key] = value
		return
	}
	m.keys = append(m.keys, key)
	m.values[key] = value
}

func (m *OrderedMap[T]) Get(key string) (value T, ok bool) {
	value, ok = m.values[key]
	return
}

func (m *OrderedMap[T]) Delete(key string) {
	delete(m.values, key)
	for i, k := range m.keys {
		if k == key {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			return
		}
	}
}

func (m *OrderedMap[T]) Range(f func(key string, value T) (bool, error)) error {
	for _, key := range m.keys {
		if ok, err := f(key, m.values[key]); err != nil {
			return err
		} else if !ok {
			break
		}
	}

	return nil
}
