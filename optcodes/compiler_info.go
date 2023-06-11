package optcodes

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

var debugPrefixes = map[string]string{
	"0x6060604052": "0.5.x and earlier",
	"0x60606040526040805193826020810185905260": "0.5.x and earlier",
	"0x6080604052":                         "0.6.x",
	"0x6080604052600436106100365763":       "0.7.x",
	"0x608060405260043610610036576000":     "0.8.x",
	"0x6080604052600436106100365760003560": "0.8.x",
	"0x608060405260008054600160a060020a03199081163390836080845292909216608482015291829082902080547ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff16906020019091905050336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507fa26469706673582212208d3c2d42a16a8f6d4f1e832a1b6d1963ed77f7989a0147e84ef2e73c4b6d073d64736f6c63430007060033": "0.8.x",
	"0x6080604052348015600f57600080fd5b506004361060285760003560e01c806360fe47b114602d575b600080fd5b60336035565b005b600054600160a060020a031681565b600054600160a060020a031633600160a060020a0316ff5b56fea26469706673582212200aaedce947f8a6df36f1cddc6e43a76a2921332b7a58165f8dc1b0c6f39e393a64736f6c63430006010033":                                                                                                                  "0.8.x",
	"0x6080604052348015600f57600080fd5b506004361060285760003560e01c806360fe47b114602d575b600080fd5b60336035565b005b600054600160a060020a031681565b600054600160a060020a031633600160a060020a0316ff5b56fea26469706673582212204d1b94f98be3b25c78141ea2858132015c58646a3ccf532da9fc0081cc47ac5d64736f6c63430007060033":                                                                                                                  "0.8.x",
	// Add more prefixes and versions as needed
}

// DebugInfo represents the extracted debug information.
type DebugInfo struct {
	SourceCode       string `json:"sourceCode"`
	SourceMapping    string `json:"sourceMapping"`
	ContractName     string `json:"contractName"`
	CompilerVersion  string `json:"compilerVersion"`
	Optimization     bool   `json:"optimization"`
	OptimizationRuns int    `json:"optimizationRuns"`
}

// ExtractDebugInfo extracts debug information from the bytecode.
func (d *Decompiler) ExtractDebugInfo() (*DebugInfo, error) {
	// Extract the debug information from the bytecode
	debugInfoBytes, err := d.extractDebugInfoFromBytecode()
	if err != nil {
		return nil, err
	}

	// Unmarshal the debug information JSON
	debugInfo := &DebugInfo{}
	err = json.Unmarshal(debugInfoBytes, debugInfo)
	if err != nil {
		return nil, err
	}

	return debugInfo, nil
}

// extractDebugInfoFromBytecode extracts the debug information from the bytecode.
func (d *Decompiler) extractDebugInfoFromBytecode() ([]byte, error) {
	// Extract the section containing debug information
	sectionStart, _ := d.findSectionStart()
	if sectionStart == -1 {
		return nil, fmt.Errorf("debug information section not found")
	}

	// Extract the debug information from the bytecode
	debugInfo := d.bytecode[sectionStart:]

	return debugInfo, nil
}

// findSectionStart finds the start of the section with the given prefix.
func (d *Decompiler) findSectionStart() (int, string) {
	for prefix, version := range debugPrefixes {
		if strings.HasPrefix(common.Bytes2Hex(d.bytecode), prefix) {
			return len(prefix), version
		}
	}
	return -1, ""
}

// PopulateDebugInfo populates the debug information into the provided struct.
func (d *Decompiler) PopulateDebugInfo(debugInfo *DebugInfo) error {
	// Extract the debug information from the bytecode
	extractedDebugInfo, err := d.ExtractDebugInfo()
	if err != nil {
		return err
	}

	// Populate the provided struct with the extracted debug information
	debugInfo.SourceCode = extractedDebugInfo.SourceCode
	debugInfo.SourceMapping = extractedDebugInfo.SourceMapping
	debugInfo.ContractName = extractedDebugInfo.ContractName
	debugInfo.CompilerVersion = extractedDebugInfo.CompilerVersion
	debugInfo.Optimization = extractedDebugInfo.Optimization
	debugInfo.OptimizationRuns = extractedDebugInfo.OptimizationRuns

	return nil
}
