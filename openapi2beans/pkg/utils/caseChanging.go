/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package utils

import (
	"strings"
	"unicode"
)

/*
Camel Case      | camelCase  | camel variant | ^[a-z][a-zA-Z0-9]+([A-Z][a-z0-9]+)+$
Pascal Case     | PascalCase | camel variant | ^[A-Z][a-zA-Z0-9]+([A-Z][a-z0-9]+)+$
Snake Case      | snake_case | snake variant | ^[a-z0-9]+_[a-z0-9]+(_[a-z0-9]+)*$
Screaming Snake | SNAKE_CASE | snake variant | ^[A-Z0-9]+_[A-Z0-9]+(_[A-Z0-9]+)*$
*/
const (
	CAMEL = "camel"
	PASCAL = "pascal"
	SNAKE_VARIANT = "snakeVariant"
)

func StringToCamel(inputString string) string {
	camelString := ""

	stringCase := GetCase(inputString)
	switch stringCase {
	case CAMEL: camelString = inputString
	case PASCAL: camelString = pascalToCamel(inputString)
	case SNAKE_VARIANT: snakeVariantsToCamel(inputString)
	}

	return camelString
}

func StringToPascal(inputString string) string {
	pascalString := ""

	stringCase := GetCase(inputString)
	switch stringCase {
	case CAMEL: pascalString = camelToPascal(inputString)
	case PASCAL: pascalString = inputString
	case SNAKE_VARIANT: pascalString = snakeVariantsToPascal(inputString)
	}

	return pascalString
}

func StringToSnake(inputString string) string {
	snakeString := ""

	stringCase := GetCase(inputString)
	switch stringCase {
	case CAMEL, PASCAL: snakeString = camelVariantsToSnake(inputString)
	case SNAKE_VARIANT: snakeString = snakeVariantsToSnake(inputString)
	}

	return snakeString
}

func StringToScreamingSnake(inputString string) string {
	screamingSnakeString := ""

	stringCase := GetCase(inputString)
	switch stringCase {
	case CAMEL, PASCAL: screamingSnakeString = camelVariantsToScreamingSnake(inputString)
	case SNAKE_VARIANT: screamingSnakeString = snakeVariantsToScreamingSnake(inputString)
	}

	return screamingSnakeString
}

// To camel functions
func pascalToCamel(pascalString string) string {
	return strings.ToLower(string(pascalString[0])) + pascalString[1:]
}

func snakeVariantsToCamel(snakeString string) string {
	var camelString string

	snakeString = strings.ToLower(snakeString)
	splitSnake := strings.Split(snakeString, "_")
	for i, snake := range splitSnake {
		if i != 0 {
			camelString += strings.ToUpper(string(snake[0])) + snake[1:]
		}
	}

	return camelString
}

// To pascal functions
func camelToPascal(camelString string) string {
	return strings.ToUpper(string(camelString[0])) + camelString[1:]
}

func snakeVariantsToPascal(inputString string) string {
	return camelToPascal(snakeVariantsToCamel(inputString))
}

// To snake functions
func camelVariantsToSnake(camelString string) string {
	var snakeString string

	for _, char := range camelString[1:] {
		if unicode.IsUpper(char) {
			snakeString += "_"
		}
		snakeString += string(char)
	}

	return snakeString
}

func snakeVariantsToSnake(inputString string) string {
	return strings.ToLower(inputString)
}

// To screaming snake functions
func camelVariantsToScreamingSnake(camelString string) string {
	return strings.ToUpper(camelVariantsToSnake(camelString))
}

func snakeVariantsToScreamingSnake(inputString string) string {
	return strings.ToUpper(inputString)
}
