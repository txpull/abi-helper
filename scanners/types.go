package providers

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
	ContractName         string `json:"ContractName"`
	CompilerVersion      string `json:"CompilerVersion"`
	OptimizationUsed     string `json:"OptimizationUsed"`
	Runs                 int    `json:"Runs"`
	ConstructorArguments string `json:"ConstructorArguments"`
	EVMVersion           string `json:"EvmVersion"`
	Library              string `json:"Library"`
	LicenseType          string `json:"LicenseType"`
	Proxy                int    `json:"Proxy"`
	Implementation       string `json:"Implementation"`
	SourceCode           string `json:"SourceCode"`
	ABI                  string `json:"ABI"`
	SwarmSource          string `json:"SwarmSource"`
}
