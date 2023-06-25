package types

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Method struct {
	UUID            uuid.UUID        `json:"uuid"`
	Name            string           `json:"name"`
	RawName         string           `json:"raw_name"`
	Signature       string           `json:"signature"`
	Hex             string           `json:"hex"`
	Bytes           []byte           `json:"bytes"`
	IsConstant      bool             `json:"is_constant"`
	IsPayable       bool             `json:"is_payable"`
	IsPartial       bool             `json:"is_partial"`
	Type            abi.FunctionType `json:"type"`
	StateMutability string           `json:"state_mutability"`
	Arguments       []MethodArgument `json:"arguments"`
	Returns         []MethodArgument `json:"returns"`
}

type MethodArgument struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Index int    `json:"index"` // Used for partial data to help matching the output in the future
}

func NewFullMethod(method abi.Method) *Method {
	toReturn := Method{
		UUID:            uuid.New(),
		Name:            method.Name,
		RawName:         method.RawName,
		Signature:       method.Sig,
		Hex:             common.Bytes2Hex(method.ID),
		Bytes:           method.ID,
		IsConstant:      method.IsConstant(),
		IsPayable:       method.IsPayable(),
		Type:            method.Type,
		StateMutability: method.StateMutability,
		IsPartial:       false, // This is a fully processed method so it is not partial
	}
	for _, arg := range method.Inputs {
		toReturn.Arguments = append(toReturn.Arguments, MethodArgument{
			Name: arg.Name,
			Type: arg.Type.String(),
		})
	}

	for _, arg := range method.Outputs {
		toReturn.Returns = append(toReturn.Returns, MethodArgument{
			Name: arg.Name,
			Type: arg.Type.String(),
		})
	}

	return &toReturn
}

func NewFourByteMethod(hexSignature string, name string, signature string, arguments []string) (*Method, error) {
	method := strings.TrimLeft(hexSignature, "0x") // We don't want the 0x prefix

	if len(method)%2 != 0 {
		return nil, fmt.Errorf("invalid method length: %d", len(method))
	}

	toReturn := Method{
		UUID:      uuid.New(),
		Name:      name,
		RawName:   name,
		Signature: signature,
		Hex:       method,
		IsPartial: true,
	}

	var signatureBytes []byte

	for i := 0; i < len(method); i += 2 {
		var b byte
		_, err := fmt.Sscanf(method[i:i+2], "%02x", &b)
		if err != nil {
			return nil, err
		}
		signatureBytes = append(signatureBytes, b)
	}
	toReturn.Bytes = signatureBytes

	re := regexp.MustCompile(`[\d\[\]]+`)

	for i := 0; i < len(arguments); i++ {
		if i+1 < len(arguments) {
			name := re.ReplaceAllString(arguments[i], "")

			toReturn.Arguments = append(toReturn.Arguments, MethodArgument{
				Name:  name,
				Type:  arguments[i],
				Index: i,
			})
		}
	}

	return &toReturn, nil
}

func (m *Method) GetArgumentsAsJSON() string {
	return toJSON(m.Arguments)
}

func (m *Method) GetReturnsAsJSON() string {
	return toJSON(m.Returns)
}

func (m *Method) MarshalBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(m); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *Method) UnmarshalBytes(data []byte) error {
	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(m)
	if err != nil {
		return err
	}

	return nil
}

func toJSON(data interface{}) string {
	bytes, err := json.Marshal(data)
	if err != nil {
		zap.L().Error(
			"Error marshalling data to JSON",
			zap.Error(err),
			zap.Any("data", data),
		)
		return "[]"
	}
	return string(bytes)
}
