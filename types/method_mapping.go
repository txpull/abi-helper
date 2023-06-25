package types

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/google/uuid"
)

type MethodMapping struct {
	UUID         uuid.UUID `json:"uuid"`
	ContractUUID uuid.UUID `json:"contract_uuid"`
	MethodUUID   uuid.UUID `json:"method_uuid"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewMethodMapping(contract *Contract, method *Method) *MethodMapping {
	toReturn := MethodMapping{
		UUID:         uuid.New(),
		ContractUUID: contract.UUID,
		MethodUUID:   method.UUID,
	}

	return &toReturn
}

func (m *MethodMapping) MarshalBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(m); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *MethodMapping) UnmarshalBytes(data []byte) error {
	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(m)
	if err != nil {
		return err
	}

	return nil
}
