package scanners

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBscScanProvider_ScanContract(t *testing.T) {
	// Create a mock HTTP server for BscScan API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with a sample BscScan response for a known contract address
		if r.URL.Query().Get("address") == "0x123456789abcdef" {
			response := `{"status": "1", "message": "OK", "result": [{"ContractName": "MyContract", "CompilerVersion": "0.8.9"}]}`
			fmt.Fprintln(w, response)
			return
		}

		// Respond with a sample BscScan response for a not known contract address
		if r.URL.Query().Get("address") == "0x123456789abcdefg" {
			response := `{"status": "1", "message": "OK", "result": [{"ContractName": "", "ABI": "Contract source code not verified"}]}`
			fmt.Fprintln(w, response)
			return
		}

		// Respond with a non-successful response for invalid contract addresses
		response := `{"status":"0","message":"NOTOK","result":"Invalid Address format"}`
		fmt.Fprintln(w, response)
	}))
	defer server.Close()

	// Create a BscScan provider with the mock server URL
	bscScanProvider := NewBscScanProvider(server.URL, "")

	// Define test cases for different contract addresses
	testCases := []struct {
		contractAddress   string
		expectedName      string
		expectedCompiler  string
		expectError       bool
		expectedErrorText string
	}{
		{
			contractAddress:   "0x123456789abcdef",
			expectedName:      "MyContract",
			expectedCompiler:  "0.8.9",
			expectError:       false,
			expectedErrorText: "",
		},
		{
			contractAddress:   "0x123456789abcdefg",
			expectedName:      "MyContract",
			expectedCompiler:  "0.8.9",
			expectError:       true,
			expectedErrorText: "contract not found",
		},
		{
			contractAddress:   "0xabcdef123456789",
			expectedName:      "",
			expectedCompiler:  "",
			expectError:       true,
			expectedErrorText: "Invalid Address format",
		},
	}

	// Run sub-tests for each test case
	for _, tc := range testCases {
		t.Run(tc.contractAddress, func(t *testing.T) {
			// Perform the contract scan
			result, err := bscScanProvider.ScanContract(tc.contractAddress)
			if tc.expectError {
				// Verify that an error occurred as expected
				if err == nil {
					t.Error("expected error, got nil")
				} else if err.Error() != tc.expectedErrorText {
					t.Errorf("unexpected error message; got %s, want %s", err.Error(), tc.expectedErrorText)
				}
			} else {
				// Verify the contract information
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				t.Logf("Got result: %+v - addr: %s", result, tc.contractAddress)

				if result.Name != tc.expectedName {
					t.Errorf("unexpected contract name; got %s, want %s", result.Name, tc.expectedName)
				}
				if result.CompilerVersion != tc.expectedCompiler {
					t.Errorf("unexpected compiler version; got %s, want %s", result.CompilerVersion, tc.expectedCompiler)
				}
			}
		})
	}
}
