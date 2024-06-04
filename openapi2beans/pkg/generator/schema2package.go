/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package generator

import (
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
)

func translateSchemaTypesToJavaPackage(schemaTypes map[string]*SchemaType, packageName string) (javaPackage *JavaPackage) {
	javaPackage = NewJavaPackage(packageName)
	for _, schemaType := range schemaTypes {
		description := strings.Split(schemaType.description, "\n")
		if len(description) == 1 && description[0] == "" {
			description = nil
		} else if len(description) > 1 {
			description = description[:len(description)-2]
		}

		if schemaType.ownProperty.IsEnum() {
			enumValues := possibleValuesToEnumValues(schemaType.ownProperty.possibleValues)

			javaEnum := NewJavaEnum(convertToPascalCase(schemaType.name), description, enumValues, javaPackage)
			javaEnum.Sort()

			javaPackage.Enums[convertToPascalCase(schemaType.name)] = javaEnum
		} else {
			dataMembers, requiredMembers, constantDataMembers := retrieveDataMembersFromSchemaType(schemaType)

			javaClass := NewJavaClass(convertToPascalCase(schemaType.name), description, javaPackage, dataMembers, requiredMembers, constantDataMembers)
			javaClass.Sort()
			javaPackage.Classes[convertToPascalCase(schemaType.name)] = javaClass
		}
	}
	return javaPackage
}

func possibleValuesToEnumValues(possibleValues map[string]string) (enumValues []EnumValues) {
	for _, value := range possibleValues {
		var constantFormatName string
		var stringFormat string
		if value != "nil"{
			constantFormatName = strcase.ToScreamingSnake(value)
			stringFormat = value
			enumValue := EnumValues {
				ConstFormatName: constantFormatName,
				StringFormat: stringFormat,
			}
			
			enumValues = append(enumValues, enumValue)
		}
	}
	return enumValues
}

func retrieveDataMembersFromSchemaType(schemaType *SchemaType) (dataMembers []*DataMember, requiredMembers []*RequiredMember, constantDataMembers []*DataMember) {
	for _, property := range schemaType.properties {
		var constVal string
		name := property.name
		description := strings.Split(property.description, "\n")
		if len(description) == 1 && description[0] == "" {
			description = nil
		} else if len(description) > 1 {
			description = description[:len(description)-2]
		}
		if property.IsConstant() {
			posVal := possibleValuesToEnumValues(property.GetPossibleValues())
			name = convertToConstName(name)
			constVal = convertConstValueToJavaReadable(posVal[0].StringFormat, property.typeName)

			constDataMember := DataMember{
				Name:          name,
				CamelCaseName: convertToPascalCase(name),
				MemberType:    propertyToJavaType(property),
				Description:   description,
				ConstantVal:   constVal,
			}

			constantDataMembers = append(constantDataMembers, &constDataMember)

		} else {

			dataMember := DataMember{
				Name:          name,
				CamelCaseName: convertToPascalCase(name),
				MemberType:    propertyToJavaType(property),
				Description:   description,
				ConstantVal:   constVal,
			}
			dataMembers = append(dataMembers, &dataMember)

			if property.IsSetInConstructor() {
				requiredMember := RequiredMember{
					DataMember: &dataMember,
				}
				requiredMembers = append(requiredMembers, &requiredMember)
			}
		}

	}
	
	if len(requiredMembers) > 0 {
		requiredMembers[0].IsFirst = true
	}
	
	return dataMembers, requiredMembers, constantDataMembers
}

func propertyToJavaType(property *Property) string {
	javaType := ""
	if property.IsReferencing() || property.typeName == "object" || property.IsEnum() {
		javaType = property.resolvedType.name
	} else {
		if property.typeName == "string" {
			javaType = "String"
		} else if property.typeName == "integer" {
			javaType = "int"
		} else if property.typeName == "number" {
			javaType = "double"
		} else if property.typeName == "" {
			javaType = "Object"
		} else {
			javaType = property.typeName
		}
	}

	if property.IsCollection() {
		dimensions := property.cardinality.max / MAX_ARRAY_CAPACITY
		for i := 0; i < dimensions; i++ {
			javaType += "[]"
		}
	}

	return javaType
}

// capitilises the first letter of a string e.g. anIntVar -> AnIntVar
// current use cases are converting variable names for use in getters and setters
// e.g. getanIntVar -> getAnIntVar
// and converting enum names to begin with capital letter for java naming conventions
func convertToPascalCase(name string) string {
	initialLetter := name[0]
	camelCaseName := strings.ToUpper(string(initialLetter)) + name[1:]
	return camelCaseName
}

// converts a name from camel/pascal case to uppercase snake case
// e.g. myConstName -> MY_CONST_NAME
func convertToConstName(name string) string {
	var matchFirstCap = regexp.MustCompile("(.[^_])([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	constName := matchFirstCap.ReplaceAllString(name, "${1}_${2}")
	constName = matchAllCap.ReplaceAllString(constName, "${1}_${2}")

	return strings.ToUpper(constName)
}

func convertConstValueToJavaReadable(constVal string, constType string) string {
	if constType == "string" {
		constVal = "\"" + constVal + "\""
	}
	return constVal
}