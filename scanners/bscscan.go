package scanners

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Default BscScan API URL
const BSCSCAN_API_URL = "https://api.bscscan.com/api"

// BscScanProvider represents the BscScan scanner provider.
type BscScanProvider struct {
	url    string
	apiKey string
}

// ScanContract scans the contract using the BscScan provider.
func (p *BscScanProvider) ScanContract(contractAddress string) (*Result, error) {
	// Construct the BscScan API URL
	url := fmt.Sprintf("%s?module=contract&action=getsourcecode&address=%s&apikey=%s", p.url, contractAddress, p.apiKey)

	// Send HTTP GET request to the API URL
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %s", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %s", err)
	}

	// BSCSCAN returns response as string or as [](objects...) and therefore lets do a hack
	// to make sure all is working properly.
	// DO NOT like it at all...

	if strings.Contains(string(body), "NOTOK") {
		// Unmarshal the response body into BscScanErrorResponse struct
		var bscScanResponse BscScanErrorResponse
		if err := json.Unmarshal(body, &bscScanResponse); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error response: %s", err)
		}

		return nil, errors.New(bscScanResponse.Result)
	}

	// Unmarshal the response body into BscScanResponse struct
	var bscScanResponse BscScanResponse
	if err := json.Unmarshal(body, &bscScanResponse); err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("failed to unmarshal BscScan response: %s", err)
	}

	toReturn := bscScanResponse.Result[0]

	if toReturn.ABI == "Contract source code not verified" {
		return nil, fmt.Errorf("contract not found")
	}

	return &toReturn, nil
}

// NewBscScanProvider creates a new instance of BscScanProvider with the provided API key and API URL.
func NewBscScanProvider(url, apiKey string) *BscScanProvider {
	return &BscScanProvider{
		apiKey: apiKey,
		url:    url,
	}
}
