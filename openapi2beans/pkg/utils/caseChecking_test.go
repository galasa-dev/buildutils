package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//
// IsSnakeVariant tests
//
func TestIsSnakeVariantWithRegularSnake(t *testing.T) {
	// Given...
	snakeString := "this_is_snake"

	// When...
	isSnake := IsSnakeVariantCase(snakeString)

	// Then...
	assert.True(t, isSnake)
}

func TestIsSnakeVariantWithScreamingSnake(t *testing.T) {
	// Given...
	snakeString := "THIS_IS_SNAKE"

	// When...
	isSnake := IsSnakeVariantCase(snakeString)

	// Then...
	assert.True(t, isSnake)
}

func TestIsSnakeVariantWithSpacedWords(t *testing.T) {
	// Given...
	wordsString := "this is a regular sentence"

	// When...
	isSnake := IsSnakeVariantCase(wordsString)

	// Then...
	assert.False(t, isSnake)
}

func TestIsSnakeVariantWithCamelCase(t *testing.T) {
	// Given...
	camelString := "thisIsCamel"

	// When...
	isSnake := IsSnakeVariantCase(camelString)

	// Then...
	assert.False(t, isSnake)
}

func TestIsSnakeVariantWithPascalCase(t *testing.T) {
	// Given...
	pascalString := "ThisIsPascal"

	// When...
	isSnake := IsSnakeVariantCase(pascalString)

	// Then...
	assert.False(t, isSnake)
}

func TestIsSnakeVariantWithSnakeNumber(t *testing.T) {
	// Given...
	snakeString := "this_is_snake_1"

	// When...
	isSnake := IsSnakeVariantCase(snakeString)

	// Then...
	assert.True(t, isSnake)
}

func TestIsSnakeVariantWithSingleWord(t *testing.T) {
	// Given...
	wordString := "snake"

	// When...
	isSnake := IsSnakeVariantCase(wordString)

	// Then...
	assert.False(t, isSnake)
}

func TestIsSnakeVariantWith2Parts(t *testing.T) {
	// Given...
	snakeString := "snake_case"

	// When...
	isSnake := IsSnakeVariantCase(snakeString)

	// Then...
	assert.True(t, isSnake)
}

//
// IsCamelVariant tests
//

func TestIsCamelVariantWithCamelString(t *testing.T) {
	// Given...
	camelString := "camelString"

	// When...
	isCamelVariant := isCamelVariant(camelString)

	// Then...
	assert.True(t, isCamelVariant)
}

func TestIsCamelVariantWithPascalString(t *testing.T) {
	// Given...
	pascalString := "camelString"

	// When...
	isCamelVariant := isCamelVariant(pascalString)

	// Then...
	assert.True(t, isCamelVariant)
}

func TestIsCamelVariantWithSpace(t *testing.T) {
	// Given...
	camelString := "camel String"

	// When...
	isCamelVariant := isCamelVariant(camelString)

	// Then...
	assert.False(t, isCamelVariant)
}

func TestIsCamelVariantWithComma(t *testing.T) {
	// Given...
	camelString := "camel,String"

	// When...
	isCamelVariant := isCamelVariant(camelString)

	// Then...
	assert.False(t, isCamelVariant)
}

func TestIsCamelVariantWithDot(t *testing.T) {
	// Given...
	camelString := "camel.String"

	// When...
	isCamelVariant := isCamelVariant(camelString)

	// Then...
	assert.False(t, isCamelVariant)
}

func TestIsCamelVariantWithUnderscore(t *testing.T) {
	// Given...
	camelString := "camel_String"

	// When...
	isCamelVariant := isCamelVariant(camelString)

	// Then...
	assert.False(t, isCamelVariant)
}

func TestIsCamelVariantWithNewLine(t *testing.T) {
	// Given...
	camelString := `camel
String`

	// When...
	isCamelVariant := isCamelVariant(camelString)

	// Then...
	assert.False(t, isCamelVariant)
}

func TestIsCamelVariantWithAppostrophe(t *testing.T) {
	// Given...
	camelString := "camel'String"

	// When...
	isCamelVariant := isCamelVariant(camelString)

	// Then...
	assert.False(t, isCamelVariant)
}

func TestIsCamelVariantWithDoubleQuotes(t *testing.T) {
	// Given...
	camelString := "camel\"String"

	// When...
	isCamelVariant := isCamelVariant(camelString)

	// Then...
	assert.False(t, isCamelVariant)
}

func TestIsCamelVariantWithSingleWord(t *testing.T) {
	// Given...
	camelString := "camel"

	// When...
	isCamelVariant := isCamelVariant(camelString)

	// Then...
	assert.False(t, isCamelVariant)
}

func TestIsCamelVariantWithCapAtEnd(t *testing.T) {
	// Given...
	camelString := "camelS"

	// When...
	isCamelVariant := isCamelVariant(camelString)

	// Then...
	assert.False(t, isCamelVariant)
}

func TestIsCamelVariantWithThreeCharacters(t *testing.T) {
	// Given...
	camelString := "cAm"

	// When...
	isCamelVariant := isCamelVariant(camelString)

	// Then...
	assert.True(t, isCamelVariant)
}

func TestIsCamelVariantSingleCharacter(t *testing.T) {
	// Given...
	camelString := "c"

	// When...
	isCamelVariant := isCamelVariant(camelString)

	// Then...
	assert.False(t, isCamelVariant)
}

func TestIsCamelVariantWithTwoCharacters(t *testing.T) {
	// Given...
	camelString := "ca"

	// When...
	isCamelVariant := isCamelVariant(camelString)

	// Then...
	assert.False(t, isCamelVariant)
}

//
// IsCamel tests
//
func TestIsCamelWithCamelString(t *testing.T) {
	// Given...
	camelString := "thisIsCamel"

	// When...
	isCamel := IsCamelCase(camelString)

	// Then...
	assert.True(t, isCamel)
}

func TestIsCamelWithPascalString(t *testing.T) {
	// Given...
	pascalString := "ThisIsCamel"

	// When...
	isCamel := IsCamelCase(pascalString)

	// Then...
	assert.False(t, isCamel)
}

func TestIsCamelWith2WordCamel(t *testing.T) {
	// Given...
	pascalString := "camelCase"

	// When...
	isCamel := IsCamelCase(pascalString)

	// Then...
	assert.True(t, isCamel)
}

func TestIsCamelWithSnakeCase(t *testing.T) {
	// Given...
	snakeString := "snake_case"

	// When...
	isCamel := IsCamelCase(snakeString)

	// Then...
	assert.False(t, isCamel)
}

func TestIsCamelWithScreamingSnakeCase(t *testing.T) {
	// Given...
	snakeString := "SCREAMING_SNAKE"

	// When...
	isCamel := IsCamelCase(snakeString)

	// Then...
	assert.False(t, isCamel)
}

func TestIsCamelWithSpacedWords(t *testing.T) {
	// Given...
	wordsString := "this is a regular sentence"

	// When...
	isCamel := IsCamelCase(wordsString)

	// Then...
	assert.False(t, isCamel)
}

func TestIsCamelWithNumbers(t *testing.T) {
	// Given...
	camelString := "camelString12"

	// When...
	isCamel := IsCamelCase(camelString)

	// Then...
	assert.True(t, isCamel)
}