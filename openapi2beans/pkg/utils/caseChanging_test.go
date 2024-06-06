package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCamelToPascalWithRegularCamel(t *testing.T) {
	// Given...
	camelString := "thisIsCamel"

	// When...
	pascalString := camelToPascal(camelString)

	// Then...
	assert.Equal(t, "ThisIsCamel", pascalString)
}