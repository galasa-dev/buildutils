package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)



func TestStringToCamel(t *testing.T) {
	// with pascal
	pascalString := "PascalCaseString"
	camelString := StringToCamel(pascalString)
	assert.Equal(t, "pascalCaseString", camelString)

	// with snake
	snakeString := "snake_case_string"
	camelString = StringToCamel(snakeString)
	assert.Equal(t, "snakeCaseString", camelString)

	// with screaming snake
	snakeString = "SNAKE_CASE_STRING"
	camelString = StringToCamel(snakeString)
	assert.Equal(t, "snakeCaseString", camelString)
}