package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCamelToPascalWithRegularCamel(t *testing.T) {
	camelString := "thisIsCamelCase"

	pascalString := CamelToPascal(camelString)

	assert.Equal(t, "ThisIsCamelCase", pascalString)
}

func TestCamelToPascalWithPascal(t *testing.T) {
	fauxCamelString := "ThisIsPascal"
	
	pascalString := CamelToPascal(fauxCamelString)

	assert.Equal(t, )
}