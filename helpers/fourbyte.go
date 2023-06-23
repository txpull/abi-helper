package helpers

import "strings"

func ExtractFourByteMethodAndArgumentTypes(signature string) (string, []string) {
	// Find the substring between the opening and closing parentheses
	openParenIndex := strings.Index(signature, "(")
	closeParenIndex := strings.LastIndex(signature, ")")

	methodName := signature[:openParenIndex]

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

	return methodName, arguments
}
