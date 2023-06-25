package scanners

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

// BitQuery represents a client for the BitQuery API.
type BitQueryProvider struct {
	// Url is the endpoint of the BitQuery API.
	Url string
	// ApiKey is the key used for authenticating with the BitQuery API.
	ApiKey string
}

// ContractCreationInfo represents the information about a smart contract creation.
type ContractCreationInfo struct {
	// Data is the main data structure that holds the smart contract creation information.
	Data struct {
		// SmartContractCreation holds the details of the smart contract creation.
		SmartContractCreation struct {
			// SmartContractCalls is a slice of smart contract calls made during the creation.
			SmartContractCalls []struct {
				// Transaction holds the details of the transaction used for the smart contract creation.
				Transaction struct {
					// Hash is the hash of the transaction.
					Hash string `json:"hash"`
				} `json:"transaction"`
				// Block holds the details of the block in which the smart contract creation transaction was included.
				Block struct {
					// Height is the height of the block.
					Height int `json:"height"`
				} `json:"block"`
			} `json:"smartContractCalls"`
		} `json:"smartContractCreation"`
	} `json:"data"`
}

// NewBitQuery creates a new BitQuery client with the provided URL and API key.
func NewBitQueryProvider(url, apiKey string) *BitQueryProvider {
	return &BitQueryProvider{
		Url:    url,
		ApiKey: apiKey,
	}
}

// GetContractCreationInfo sends a query to the BitQuery API and returns the contract creation information.
// The query parameter should be a map where the key is the name of the query parameter and the value is the value of the query parameter.
// It returns a pointer to a ContractCreationInfo struct containing the response from the BitQuery API, or an error if there was an issue sending the request or decoding the response.
func (b *BitQueryProvider) GetContractCreationInfo(query map[string]string) (*ContractCreationInfo, error) {
	jsonData, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", b.Url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", b.ApiKey)

	client := &http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info ContractCreationInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}
