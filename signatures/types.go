// Package signatures provides functionality for working with signature data.
package signatures

import (
	"bytes"
	"encoding/gob"
	"regexp"
	"time"
)

type Signature struct {
	ID              uint64        `json:"constant"`
	Text            string        `json:"text_signature"`
	Hex             string        `json:"hex_signature"`
	Inputs          []InputOutput `json:"inputs"`
	Name            string        `json:"name"`
	Outputs         []InputOutput `json:"outputs"`
	Payable         bool          `json:"payable"`
	StateMutability string        `json:"stateMutability"`
	CreatedAt       time.Time     `json:"created_at"`
}

type InputOutput struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Index int    `json:"type"`
}

func (r *Signature) MarshalBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r *Signature) UnmarshalBytes(data []byte) error {
	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(r)
	if err != nil {
		return err
	}

	return nil
}

func NewSignature(
	ID uint64,
	TextSignature string,
	HexSignature string,
	Inputs []InputOutput,
	Name string,
	Outputs []InputOutput,
	Payable bool,
	StateMutability string,
	CreatedAt time.Time,
) *Signature {
	return &Signature{
		ID:              ID,
		Text:            TextSignature,
		Hex:             HexSignature,
		Inputs:          Inputs,
		Name:            Name,
		Outputs:         Outputs,
		Payable:         Payable,
		StateMutability: StateMutability,
		CreatedAt:       CreatedAt,
	}
}

func ParseSignatureInputFromArray(dataArray []string) []InputOutput {
	var result []InputOutput

	for i := 0; i < len(dataArray); i++ {
		if i+1 < len(dataArray) {
			name := stripNumbersAndBrackets(dataArray[i])
			inputOutput := InputOutput{
				Name:  name,
				Type:  dataArray[i],
				Index: i,
			}
			result = append(result, inputOutput)
		}
	}

	return result
}

func stripNumbersAndBrackets(input string) string {
	re := regexp.MustCompile(`[\d\[\]]+`)
	return re.ReplaceAllString(input, "")
}
