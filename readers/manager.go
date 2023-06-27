package readers

import "context"

// Manager is a struct that manages multiple Reader instances.
type Manager struct {
	// ctx is the context in which the Manager operates.
	ctx context.Context
	// readers is a map that stores Reader instances by their names.
	readers map[string]Reader
	// priorityReader is the name of the Reader that has priority over others.
	priorityReader string
}

// ManagerOption is a function that applies a certain configuration to a Manager instance.
type ManagerOption func(*Manager)

// WithPriorityReader is a ManagerOption that sets the priority reader of a Manager.
func WithPriorityReader(reader string) ManagerOption {
	return func(m *Manager) {
		m.priorityReader = reader
	}
}

// WithReader is a ManagerOption that adds a new Reader to a Manager.
func WithReader(name string, reader Reader) ManagerOption {
	return func(m *Manager) {
		m.readers[name] = reader
	}
}

// NewManager creates a new Manager instance with the provided context and options.
func NewManager(ctx context.Context, opts ...ManagerOption) (*Manager, error) {
	manager := &Manager{ctx: ctx, readers: make(map[string]Reader)}

	for _, opt := range opts {
		opt(manager)
	}

	return manager, nil
}

// AddReader adds a new Reader to the Manager.
func (m *Manager) AddReader(name string, reader Reader) {
	m.readers[name] = reader
}

// SetPriorityReader sets the priority reader of the Manager.
func (m *Manager) SetPriorityReader(name string) error {
	if _, ok := m.readers[name]; !ok {
		return ErrReaderNotFound
	}

	m.priorityReader = name
	return nil
}

// GetPriorityReader returns the priority reader of the Manager.
func (m *Manager) GetPriorityReader() Reader {
	if m.priorityReader == "" {
		return nil
	}

	return m.readers[m.priorityReader]
}

// GetReaders returns all the readers of the Manager.
func (m *Manager) GetReaders() map[string]Reader {
	return m.readers
}

// GetReaderByName returns a Reader by its name.
func (m *Manager) GetReaderByName(name string) (Reader, error) {
	if reader, ok := m.readers[name]; ok {
		return reader, nil
	}

	return nil, ErrReaderNotFound
}

// GetSortedReaders returns all the readers of the Manager, with the priority reader being the first in the list.
func (m *Manager) GetSortedReaders() []Reader {
	readers := make([]Reader, 0, len(m.readers))

	// If a priority reader is set, add it to the slice first
	if m.priorityReader != "" {
		readers = append(readers, m.readers[m.priorityReader])
	}

	// Add the rest of the readers to the slice
	for name, reader := range m.readers {
		if name != m.priorityReader {
			readers = append(readers, reader)
		}
	}

	return readers
}
