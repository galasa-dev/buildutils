/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package generator

import (
	"fmt"
	"testing"

	"github.com/galasa-dev/cli/pkg/files"
	"github.com/stretchr/testify/assert"
)

func getGeneratedCodeFilePathWithPackage(storeFilepath string, packageName string, name string) string {
	return storeFilepath + "/src/main/java/" + packageName + "/" + name + ".java"
}

func assertVariableSetCorrectly(t *testing.T, generatedFile string, description []string, name string, javaExpectedVarType string) {
	assignmentLiteral := `private %s %s;`
	assignment := fmt.Sprintf(assignmentLiteral, javaExpectedVarType, name)
	assert.Contains(t, generatedFile, assignment)

	for _, line := range description {
		assert.Contains(t, generatedFile, "// " + line)
	}
}

func assertVariableMatchesGetter(t *testing.T, generatedFile string, name string, camelName string, javaExpectedVarType string) {
	getterLiteral := `    public %s get%s() {
        return this.%s;
    }`
	getter := fmt.Sprintf(getterLiteral, javaExpectedVarType, camelName, name)
	assert.Contains(t, generatedFile, getter)
}

func assertVariableMatchesSetter(t *testing.T, generatedFile string, name string, camelName string, javaExpectedVarType string) {
	setterLiteral := `    public void set%s(%s %s) {
        this.%s = %s;
    }`
	setter := fmt.Sprintf(setterLiteral, camelName, javaExpectedVarType, name, name, name)
	assert.Contains(t, generatedFile, setter)
}

func assertEnumFilesGeneratedOkWithStringParams(t *testing.T, generatedFile string, name string, values ... string) {
	assert.Contains(t, generatedFile, "package "+ TARGET_JAVA_PACKAGE)
	assert.Contains(t, generatedFile, "public enum " + name)
	for _, value := range values {
		assert.Contains(t, generatedFile, value + ",")
	}
}

func assertConstVarGeneratedOk(t *testing.T, generatedFile string, description []string, name string, javaExpectedVarType string, expectedVal string) {
	assignmentLiteral := `public static final %s %s = %s;`
	assignment := fmt.Sprintf(assignmentLiteral, javaExpectedVarType, name, expectedVal)
	assert.Contains(t, generatedFile, assignment)

	for _, line := range description {
		assert.Contains(t, generatedFile, "// " + line)
	}
}

func TestGenerateFilesProducesFileFromSingleGenericObjectSchema(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	testapiyaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
`
	mockFileSystem.WriteTextFile(apiFilePath, testapiyaml)
	
	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
}

func TestGenerateFilesProducesCorrectClassDescription(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
`
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assert.Contains(t, generatedClassFile, "// A simple example")
}

func TestGenerateFilesProducesCorrectVariableCode(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myStringVar:
          type: string
`
	// When...
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableMatchesSetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableSetCorrectly(t, generatedClassFile, []string{}, "myStringVar", "String")
}

func TestGenerateFilesProducesCorrectVariableDescription(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myStringVar:
          type: string
          description: a test string
`
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableMatchesSetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test string"}, "myStringVar", "String")
}

func TestGenerateFilesProducesMultipleVariables(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myStringVar:
          type: string
          description: a test string
        myIntVar:
          type: integer
          description: a test integer
`
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableMatchesSetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test string"}, "myStringVar", "String")
	assertVariableMatchesGetter(t, generatedClassFile, "myIntVar", "MyIntVar", "int")
	assertVariableMatchesSetter(t, generatedClassFile, "myIntVar", "MyIntVar", "int")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test integer"}, "myIntVar", "int")
}

func TestGenerateFilesProducesVariablesOfAllPrimitiveTypes(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myStringVar:
          type: string
          description: a test string
        myIntVar:
          type: integer
          description: a test integer
        myBoolVar:
          type: boolean
          description: a test boolean
        myDoubleVar:
          type: number
          description: a test double
`
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableMatchesSetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test string"}, "myStringVar", "String")
	assertVariableMatchesGetter(t, generatedClassFile, "myIntVar", "MyIntVar", "int")
	assertVariableMatchesSetter(t, generatedClassFile, "myIntVar", "MyIntVar", "int")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test integer"}, "myIntVar", "int")
	assertVariableMatchesGetter(t, generatedClassFile, "myBoolVar", "MyBoolVar", "boolean")
	assertVariableMatchesSetter(t, generatedClassFile, "myBoolVar", "MyBoolVar", "boolean")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test boolean"}, "myBoolVar", "boolean")
	assertVariableMatchesGetter(t, generatedClassFile, "myDoubleVar", "MyDoubleVar", "double")
	assertVariableMatchesSetter(t, generatedClassFile, "myDoubleVar", "MyDoubleVar", "double")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test double"}, "myDoubleVar", "double")
}

func TestGenerateFilesProcessesRequiredVariable(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      required: [myStringVar]
      properties:
        myStringVar:
          type: string
          description: a test string
`
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableMatchesSetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test string"}, "myStringVar", "String")
	assert.Contains(t, generatedClassFile, `    public MyBeanName (String myStringVar) {
        this.myStringVar = myStringVar;
    }`)
}

func TestGenerateFilesProducesMultipleRequiredVariables(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      required: [myStringVar, myIntVar]
      description: A simple example
      properties:
        myStringVar:
          type: string
          description: a test string
        myIntVar:
          type: integer
          description: a test integer
`
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableMatchesSetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test string"}, "myStringVar", "String")
	assertVariableMatchesGetter(t, generatedClassFile, "myIntVar", "MyIntVar", "int")
	assertVariableMatchesSetter(t, generatedClassFile, "myIntVar", "MyIntVar", "int")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test integer"}, "myIntVar", "int")
	assert.Contains(t, generatedClassFile, `    public MyBeanName (int myIntVar, String myStringVar) {
        this.myIntVar = myIntVar;
        this.myStringVar = myStringVar;
    }`)
}

func TestGenerateFilesProducesMultipleVariablesWithMixedRequiredStatus(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      required: [myStringVar]
      description: A simple example
      properties:
        myStringVar:
          type: string
          description: a test string
        myIntVar:
          type: integer
          description: a test integer
`
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableMatchesSetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test string"}, "myStringVar", "String")
	assertVariableMatchesGetter(t, generatedClassFile, "myIntVar", "MyIntVar", "int")
	assertVariableMatchesSetter(t, generatedClassFile, "myIntVar", "MyIntVar", "int")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test integer"}, "myIntVar", "int")
	assert.Contains(t, generatedClassFile, `    public MyBeanName (String myStringVar) {
        this.myStringVar = myStringVar;
    }`)
}

func TestGenerateFilesProducesArray(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myArrayVar:
          type: array
          description: a test array
          items:
            type: string
`
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myArrayVar", "MyArrayVar", "String[]")
	assertVariableMatchesSetter(t, generatedClassFile, "myArrayVar", "MyArrayVar", "String[]")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test array"}, "myArrayVar", "String[]")
}

func TestGenerateFilesProduces2DArrayFromNestedArrayStructure(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myArrayVar:
          type: array
          description: a test 2d array
          items:
            type: array
            items:
              type: string
`
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myArrayVar", "MyArrayVar", "String[][]")
	assertVariableMatchesSetter(t, generatedClassFile, "myArrayVar", "MyArrayVar", "String[][]")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test 2d array"}, "myArrayVar", "String[][]")
}

func TestGenerateFilesProduces3DArrayFromNestedArrayStructure(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myArrayVar:
          type: array
          description: a test 3d array
          items:
            type: array
            items:
              type: array
              items:
                type: string
`
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myArrayVar", "MyArrayVar", "String[][][]")
	assertVariableMatchesSetter(t, generatedClassFile, "myArrayVar", "MyArrayVar", "String[][][]")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test 3d array"}, "myArrayVar", "String[][][]")
}

func TestGenerateFilesProducesMultipleClassFilesFromNestedObject(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myNestedObject:
          type: object
`
	// When...
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myNestedObject", "MyNestedObject", "MyBeanNameMyNestedObject")
	assertVariableMatchesSetter(t, generatedClassFile, "myNestedObject", "MyNestedObject", "MyBeanNameMyNestedObject")
	assertVariableSetCorrectly(t, generatedClassFile, []string{}, "myNestedObject", "MyBeanNameMyNestedObject")
	generatedNestedClassFile := openGeneratedFile(t, mockFileSystem, getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, "MyBeanNameMyNestedObject"))
	assertClassFileGeneratedOk(t, generatedNestedClassFile, "MyBeanNameMyNestedObject")
}

func TestGenerateFilesProducesClassWithVariableOfTypeReferencedObject(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myReferencingProperty:
          $ref: '#/components/schemas/MyReferencedObject'
    MyReferencedObject:
      type: object
`
	// When...
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myReferencingProperty", "MyReferencingProperty", "MyReferencedObject")
	assertVariableMatchesSetter(t, generatedClassFile, "myReferencingProperty", "MyReferencingProperty", "MyReferencedObject")
	assertVariableSetCorrectly(t, generatedClassFile, []string{}, "myReferencingProperty", "MyReferencedObject")
	generatedNestedClassFile := openGeneratedFile(t, mockFileSystem, getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, "MyReferencedObject"))
	assertClassFileGeneratedOk(t, generatedNestedClassFile, "MyReferencedObject")
}

func TestGenerateFilesProducesArrayWithReferenceToObject(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myArrayVar:
          type: array
          description: a test array
          items:
            $ref: '#/components/schemas/MyReferencedObject'
    MyReferencedObject:
      type: object
`
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myArrayVar", "MyArrayVar", "MyReferencedObject[]")
	assertVariableMatchesSetter(t, generatedClassFile, "myArrayVar", "MyArrayVar", "MyReferencedObject[]")
	assertVariableSetCorrectly(t, generatedClassFile, []string{"a test array"}, "myArrayVar", "MyReferencedObject[]")
	generatedNestedClassFile := openGeneratedFile(t, mockFileSystem, getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, "MyReferencedObject"))
	assertClassFileGeneratedOk(t, generatedNestedClassFile, "MyReferencedObject")
}

func TestGenerateFilesProducesEnumAndClass(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myEnum:
          type: string
          enum: [randValue1, randValue2]
`
	// When...
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myEnum", "MyEnum", "MyBeanNameMyEnum")
	assertVariableMatchesSetter(t, generatedClassFile, "myEnum", "MyEnum", "MyBeanNameMyEnum")
	assertVariableSetCorrectly(t, generatedClassFile, []string{}, "myEnum", "MyBeanNameMyEnum")
	assert.Contains(t, generatedClassFile, `    public MyBeanName (MyBeanNameMyEnum myEnum) {
        this.myEnum = myEnum;
    }`)
	generatedEnumFile := openGeneratedFile(t, mockFileSystem, getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, "MyEnum"))
	assertEnumFilesGeneratedOkWithStringParams(t, generatedEnumFile, "MyEnum", "randValue1", "randValue2")
}

func TestGenerateFilesProducesEnumWithNilValueIsntSetInConstructor(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myEnum:
          type: string
          enum: [randValue1, nil]
`
	// When...
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myEnum", "MyEnum", "MyBeanNameMyEnum")
	assertVariableMatchesSetter(t, generatedClassFile, "myEnum", "MyEnum", "MyBeanNameMyEnum")
	assertVariableSetCorrectly(t, generatedClassFile, []string{}, "myEnum", "MyBeanNameMyEnum")
	assert.NotContains(t, generatedClassFile, `    public MyBeanName (MyBeanNameMyEnum myEnum) {
        this.myEnum = myEnum;
    }`)
	generatedEnumFile := openGeneratedFile(t, mockFileSystem, getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, "MyEnum"))
	assertEnumFilesGeneratedOkWithStringParams(t, generatedEnumFile, "MyEnum", "randValue1", "nil")
}

func TestGenerateFilesProducesConstantCorrectly(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myConstVar:
          type: string
          description: a test constant
          enum: [constVal]
`
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertConstVarGeneratedOk(t, generatedClassFile, []string{"a test constant"}, "MY_CONST_VAR", "String", "\"constVal\"")
}

func TestGenerateFilesProducesClassWithReferencedStringProperty(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myStringVar:
          $ref: '#/components/schemas/myReferencedProperty'
    myReferencedProperty:
      type: string
`
	// When...
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableMatchesSetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableSetCorrectly(t, generatedClassFile, []string{}, "myStringVar", "String")
}

func TestGenerateFilesProducesClassWithReferencedArrayProperty(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      properties:
        myReferencingArrayProp:
          $ref: '#/components/schemas/myReferencedArrayProperty'
    myReferencedArrayProperty:
      type: array
      items:
        type: string
`
	// When...
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myReferencingArrayProp", "MyReferencingArrayProp", "String[]")
	assertVariableMatchesSetter(t, generatedClassFile, "myReferencingArrayProp", "MyReferencingArrayProp", "String[]")
	assertVariableSetCorrectly(t, generatedClassFile, []string{}, "myReferencingArrayProp", "String[]")
}

func TestGenerateFilesProducesAcceptibleCodeUsingAllPreviousTestsAipYaml(t *testing.T) {
	// Given...
	packageName := "generated"
	mockFileSystem := files.NewMockFileSystem()
	storeFilepath := "dev/wyvinar"
	apiFilePath := "test-resources/single-bean.yaml"
	objectName := "MyBeanName"
	generatedCodeFilePath := getGeneratedCodeFilePathWithPackage(storeFilepath, packageName, objectName)
	apiYaml := `openapi: 3.0.3
components:
  schemas:
    MyBeanName:
      type: object
      description: A simple example
      required: [myNestedObject, myDoubleVar, my2DArrayVar, myEnum]
      properties:
        myStringVar:
          type: string
          description: a test string
        myIntVar:
          type: integer
          description: a test integer
        myBoolVar:
          type: boolean
          description: a test boolean
        myDoubleVar:
          type: number
          description: a test double
        myArrayVar:
          type: array
          description: a test array
          items:
            type: string
        my2DArrayVar:
          type: array
          description: a test 2d array
          items:
            type: array
            items:
              type: string
        my3DArrayVar:
          type: array
          description: a test 3d array
          items:
            type: array
            items:
              type: array
              items:
                type: string
        myNestedObject:
          type: object
        myObjectReferencingProperty:
          $ref: '#/components/schemas/MyReferencedObject'
        myAnyOfReferencingArrayVar:
          type: array
          description: a test array
          items:
            anyOf:
            - $ref: '#/components/schemas/MyReferencedObject'
        myEnum:
          type: string
          enum: [randValue1, nil]
        myConstVar:
          type: string
          description: a test constant
          enum: [constVal]
        myReferencingArrayProp:
          $ref: '#/components/schemas/myReferencedArrayProperty'
        myReferencingStringProperty:
          $ref: '#/components/schemas/myReferencedStringProperty'
    MyReferencedObject:
      type: object
    myReferencedStringProperty:
      type: string
    myReferencedArrayProperty:
      type: array
      items:
        type: string
`
	// When...
	mockFileSystem.WriteTextFile(apiFilePath, apiYaml)

	// When...
	err := GenerateFiles(mockFileSystem, storeFilepath, apiFilePath, packageName)

	// Then...
	assert.Nil(t, err)
	generatedClassFile := openGeneratedFile(t, mockFileSystem, generatedCodeFilePath)
	assertClassFileGeneratedOk(t, generatedClassFile, objectName)
	assertVariableMatchesGetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableMatchesSetter(t, generatedClassFile, "myStringVar", "MyStringVar", "String")
	assertVariableSetCorrectly(t, generatedClassFile, []string{}, "myStringVar", "String")
}