package readers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_GetReaders(t *testing.T) {
	ctx := context.TODO()
	manager := &Manager{
		ctx:     ctx,
		readers: make(map[string]Reader),
	}

	reader1 := &MockReader{}
	reader2 := &MockReader{}

	manager.AddReader("reader1", reader1)
	manager.AddReader("reader2", reader2)

	readers := manager.GetReaders()
	if len(readers) != 2 {
		t.Errorf("Expected reader count: 2, got: %d", len(readers))
	}

	if readers["reader1"] != reader1 || readers["reader2"] != reader2 {
		t.Error("Readers not retrieved correctly")
	}
}

func TestNewManager(t *testing.T) {
	ctx := context.TODO()

	// Test Manager creation with WithPriorityReader and WithReader options
	reader1 := &MockReader{}
	reader2 := &MockReader{}

	manager, err := NewManager(ctx,
		WithPriorityReader("reader1"),
		WithReader("reader1", reader1),
		WithReader("reader2", reader2),
	)

	if err != nil {
		t.Errorf("Failed to create Manager: %s", err)
	}

	if manager.ctx != ctx {
		t.Error("Context not set correctly")
	}

	if manager.priorityReader != "reader1" {
		t.Error("Priority reader not set correctly")
	}

	if len(manager.readers) != 2 {
		t.Errorf("Expected reader count: 2, got: %d", len(manager.readers))
	}

	if manager.readers["reader1"] != reader1 || manager.readers["reader2"] != reader2 {
		t.Error("Readers not added correctly")
	}
}

func TestManager_SetPriorityReader(t *testing.T) {
	ctx := context.TODO()
	manager := &Manager{
		ctx:     ctx,
		readers: make(map[string]Reader),
	}

	reader := &MockReader{}
	manager.AddReader("mock", reader)

	err := manager.SetPriorityReader("mock")
	if err != nil {
		t.Errorf("Error setting priority reader: %s", err)
	}

	priorityReader := manager.GetPriorityReader()
	if priorityReader != reader {
		t.Error("Priority reader not set correctly")
	}

	err = manager.SetPriorityReader("nonexistent")
	if !errors.Is(err, ErrReaderNotFound) {
		t.Error("Expected ErrReaderNotFound error, but got different error or nil")
	}
}

func TestManager_GetSortedReaders(t *testing.T) {
	ctx := context.TODO()
	manager := &Manager{
		ctx:     ctx,
		readers: make(map[string]Reader),
	}

	reader1 := &MockReader{}
	reader2 := &MockReader{}
	reader3 := &MockReader{}

	manager.AddReader("reader1", reader1)
	manager.AddReader("reader2", reader2)
	manager.AddReader("reader3", reader3)

	err := manager.SetPriorityReader("reader2")
	assert.Error(t, err)

	sortedReaders := manager.GetSortedReaders()
	if len(sortedReaders) != 3 {
		t.Errorf("Expected reader count: 3, got: %d", len(sortedReaders))
	}

	if sortedReaders[0] != reader2 {
		t.Error("Priority reader not placed at the beginning")
	}

	if sortedReaders[1] != reader1 || sortedReaders[2] != reader3 {
		t.Error("Readers not sorted correctly")
	}
}

func TestManager_AddReader(t *testing.T) {
	ctx := context.TODO()
	manager := &Manager{ctx: ctx, readers: make(map[string]Reader)}

	reader := &MockReader{}
	manager.AddReader("mock", reader)

	if len(manager.readers) != 1 {
		t.Errorf("Expected reader count: 1, got: %d", len(manager.readers))
	}

	if manager.readers["mock"] != reader {
		t.Error("Reader was not added correctly")
	}

	// Adding duplicate reader
	manager.AddReader("mock", reader)

	if len(manager.readers) != 1 {
		t.Errorf("Expected reader count: 1 after adding duplicate, got: %d", len(manager.readers))
	}
}

func TestManager_GetReaderByName(t *testing.T) {
	ctx := context.TODO()
	manager := &Manager{
		ctx:     ctx,
		readers: make(map[string]Reader),
	}

	reader := &MockReader{}
	manager.AddReader("mock", reader)

	foundReader, err := manager.GetReaderByName("mock")
	if err != nil {
		t.Errorf("Error getting reader: %s", err)
	}

	if foundReader != reader {
		t.Error("Reader not retrieved correctly by name")
	}

	_, err = manager.GetReaderByName("nonexistent")
	if !errors.Is(err, ErrReaderNotFound) {
		t.Error("Expected ErrReaderNotFound error, but got different error or nil")
	}
}
