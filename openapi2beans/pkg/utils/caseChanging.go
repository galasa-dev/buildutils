/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package utils

import (
	"regexp"
	"strings"
)

/*
Camel Case      | camelCase  | camel variant
Pascal Case     | PascalCase | camel variant
Snake Case      | snake_case | snake variant
Screaming Snake | SNAKE_CASE | snake variant
*/

func CamelToPacal(camelString string) string {
	return strings.ToUpper(string(camelString[0])) + camelString[1:]
}

func CamelVariantsToSnake(camelString string) string {
	var snakeString string
	var matchFirstCap = regexp.MustCompile("(.[^_])([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snakeString = matchFirstCap.ReplaceAllString(camelString, "${1}_${2}")
	snakeString = matchAllCap.ReplaceAllString(snakeString, "${1}_${2}")

	return snakeString
}

func CamelVariantsToScreamingSnake(camelString string) string {
	return strings.ToUpper(CamelVariantsToSnake(camelString))
}

func SnakeVariantsToCamel(snakeString string) string {
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