package transactions

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// Function to extract argument types from a function signature
func extractArgumentTypes(signature string) map[int64]string {
	// Find the substring between the opening and closing parentheses
	openParenIndex := strings.Index(signature, "(")
	closeParenIndex := strings.LastIndex(signature, ")")

	// Extract the string inside the outer parentheses
	argumentsStr := signature[openParenIndex+1 : closeParenIndex]

	// Split the arguments by comma
	arguments := strings.Split(argumentsStr, ",")

	// Ensure that any other argument in the list does not contain any type of the
	// parenthesies
	for i := range arguments {
		arguments[i] = strings.TrimSpace(strings.Trim(arguments[i], "("))
		arguments[i] = strings.TrimSpace(strings.Trim(arguments[i], ")"))
	}

	toReturn := map[int64]string{}

	index := int64(0)
	for _, arg := range arguments {
		toReturn[index] = arg
		index += 32
	}

	return toReturn
}

// Function to extract the value of an argument based on its type and byte position
func extractValue(argType string, startIndex int64, dataBytes []byte) interface{} {
	if len(dataBytes) < int(startIndex+32) {
		return nil
	}

	endIndex := startIndex + 32

	// Extract the hex string representing the argument value
	argValueHex := common.Bytes2Hex(dataBytes[startIndex:endIndex])

	// Remove leading zeros from the hex string
	argValueHex = strings.TrimLeft(argValueHex, "0")

	//fmt.Printf("Arg Val Index: %d - Arg: %s - Hex: %s \n", startIndex, argType, argValueHex)

	switch argType {
	case "uint128":
		// Convert the hex string to a big.Int
		argValue := new(big.Int)
		argValue.SetString(argValueHex, 16)
		return argValue

	case "uint256":
		// Convert the hex string to a little-endian byte slice
		argValueBytes := common.Hex2Bytes(argValueHex)

		// Create a new big.Int and set its value from the byte slice
		return new(big.Int).SetBytes(argValueBytes)

	case "uint256[]":
		// Extract the length of the uint256[] array
		length := new(big.Int)
		length.SetString(argValueHex, 16)

		// Initialize a slice to store the uint256 values
		uint256Array := make([]*big.Int, length.Int64())

		// Iterate over the indices of the uint256[] array
		for i := int64(0); i < length.Int64(); i++ {
			// Calculate the start and end indices for each uint256 value
			start := startIndex + (i * 32)
			end := start + 32

			// Check if the indices are within the valid range
			if end > int64(len(dataBytes)) {
				break
			}

			// Convert the hex string to a big-endian byte slice
			uint256ValueBytes := dataBytes[start:end]

			// Create a new big.Int and set its value from the byte slice
			uint256Value := new(big.Int).SetBytes(uint256ValueBytes)

			// Assign the uint256 value to the uint256Array
			uint256Array[i] = uint256Value
		}

		// FIXME: This is a very horrible hack to return just list without bunch of nil big.Int
		toReturn := []*big.Int{}

		for _, uint256 := range uint256Array {
			if uint256 != nil {
				toReturn = append(toReturn, uint256)
			}
		}

		return toReturn

	case "address":
		return common.HexToAddress(argValueHex)

	case "bytes":
		// Convert the hex string to a byte slice
		return common.Hex2Bytes(argValueHex)

	case "bytes[]":
		// Extract the length of the bytes[] array
		length := new(big.Int)
		length.SetString(argValueHex, 16)

		// Initialize a slice to store the bytes values
		bytesArray := make([][]byte, length.Int64())

		// Iterate over the indices of the bytes[] array
		for i := int64(0); i < length.Int64(); i++ {
			// Calculate the start and end indices for each bytes value
			start := startIndex + (i * 32)
			end := start + 32

			// Check if the indices are within the valid range
			if end > int64(len(dataBytes)*1) {
				return nil
			}

			// Extract the hex string representing the bytes value
			bytesValueHex := string(dataBytes[start:end])

			// Convert the hex string to a byte slice
			bytesValue, _ := hex.DecodeString(bytesValueHex)

			// Append the bytes value to the bytesArray
			bytesArray[i] = bytesValue
		}

		return bytesArray
	default:
		return nil
	}
}
