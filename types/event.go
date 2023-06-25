package types

import (
	"bytes"
	"encoding/gob"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
)

type Event struct {
	UUID        uuid.UUID       `json:"uuid"`
	Name        string          `json:"name"`
	RawName     string          `json:"raw_name"`
	Signature   string          `json:"signature"`
	Hash        common.Hash     `json:"bytes"`
	IsAnonymous bool            `json:"is_anonymous"`
	IsPartial   bool            `json:"is_partial"`
	Arguments   []EventArgument `json:"arguments"`
}

type EventArgument struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Indexed bool   `json:"indexed"`
}

func NewFullEvent(event abi.Event) *Event {
	toReturn := Event{
		UUID:        uuid.New(),
		Name:        event.Name,
		RawName:     event.RawName,
		Signature:   event.Sig,
		Hash:        event.ID,
		IsAnonymous: event.Anonymous,
		IsPartial:   false, // This is a fully processed event so it is not partial
	}

	for _, arg := range event.Inputs {
		toReturn.Arguments = append(toReturn.Arguments, EventArgument{
			Name:    arg.Name,
			Type:    arg.Type.String(),
			Indexed: arg.Indexed,
		})
	}

	return &toReturn
}

func (m *Event) GetArgumentsAsJSON() string {
	return toJSON(m.Arguments)
}

func (m *Event) MarshalBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(m); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *Event) UnmarshalBytes(data []byte) error {
	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(m)
	if err != nil {
		return err
	}

	return nil
}
