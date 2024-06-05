package utils

import (
	"strings"
	"unicode"
)


func GetCase(inputString string) string {
	var outputCase string
	if IsCamelCase(inputString) {
		outputCase = CAMEL
	} else if IsPascalCase(inputString) {
		outputCase = PASCAL
	} else if IsSnakeVariantCase(inputString) {
		IsSnakeVariantCase(inputString)
	}
	return outputCase
}

func IsSnakeVariantCase(inputString string) bool {
	var isSnakeCase bool
	wordArray := strings.Split(inputString, "_")
	isSnakeCase = len(wordArray) > 1
	return isSnakeCase
}

func IsCamelCase(inputString string) bool {
	isCamelCase := isCamelVariant(inputString)
	if unicode.IsLower(rune(inputString[0])) && !unicode.IsNumber(rune(inputString[0])) {
		isCamelCase = true
	}
	return isCamelCase
}

func IsPascalCase(inputString string) bool {
	isPascalCase := isCamelVariant(inputString)
	if unicode.IsUpper(rune(inputString[0])) && !unicode.IsNumber(rune(inputString[0])) {
		isPascalCase = true
	}
	return isPascalCase
}

func isCamelVariant(inputString string) (bool) {
	isCamelVariant := false
	if !strings.ContainsAny(inputString, " ,\n_.'\"!@#£$%^&*()_-=+[]{}:;\\|`~/?<>") {
		for i, char := range inputString[1:] {
			if unicode.IsUpper(char) && i != len(inputString)-2{
				isCamelVariant = true
			}
		}
	}
	return isCamelVariant
}