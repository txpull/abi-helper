package scanners

import (
	"bytes"
	"encoding/gob"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type BscScanResponse struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Result  []Result `json:"result"`
}

type BscScanErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

// BscScanResult represents the individual contract result in the BscScan response.
type Result struct {
	Name                 string `json:"ContractName"`
	CompilerVersion      string `json:"CompilerVersion"`
	OptimizationUsed     string `json:"OptimizationUsed"`
	Runs                 string `json:"Runs"`
	ConstructorArguments string `json:"ConstructorArguments"`
	EVMVersion           string `json:"EvmVersion"`
	Library              string `json:"Library"`
	LicenseType          string `json:"LicenseType"`
	Proxy                string `json:"Proxy"`
	Implementation       string `json:"Implementation"`
	SourceCode           string `json:"SourceCode"`
	ABI                  string `json:"ABI"`
	SwarmSource          string `json:"SwarmSource"`
}

func (r *Result) MarshalBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r *Result) UnmarshalBytes(data []byte) error {
	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(r)
	if err != nil {
		return err
	}

	return nil
}

// Unmarshal the ABI from JSON
func (r *Result) UnmarshalABI() (*abi.ABI, error) {
	parsedAbi, err := abi.JSON(strings.NewReader(r.ABI))
	if err != nil {
		return nil, err
	}

	return &parsedAbi, nil
}
