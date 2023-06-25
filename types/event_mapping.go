package types

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/google/uuid"
)

type EventMapping struct {
	UUID         uuid.UUID `json:"uuid"`
	ContractUUID uuid.UUID `json:"contract_uuid"`
	EventUUID    uuid.UUID `json:"event_uuid"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewEventMapping(contract *Contract, event *Event) *EventMapping {
	toReturn := EventMapping{
		UUID:         uuid.New(),
		ContractUUID: contract.UUID,
		EventUUID:    event.UUID,
	}

	return &toReturn
}

func (m *EventMapping) MarshalBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(m); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *EventMapping) UnmarshalBytes(data []byte) error {
	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(m)
	if err != nil {
		return err
	}

	return nil
}
