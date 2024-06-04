/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package generator

import (
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

			javaEnum := NewJavaEnum(strcase.ToCamel(schemaType.name), description, enumValues, javaPackage)
			javaEnum.Sort()

			javaPackage.Enums[strcase.ToCamel(schemaType.name)] = javaEnum
		} else {
			dataMembers, requiredMembers, constantDataMembers, hasSerializedNameDataMember := retrieveDataMembersFromSchemaType(schemaType)

			javaClass := NewJavaClass(strcase.ToCamel(schemaType.name), description, javaPackage, dataMembers, requiredMembers, constantDataMembers, hasSerializedNameDataMember)
			javaClass.Sort()
			javaPackage.Classes[strcase.ToCamel(schemaType.name)] = javaClass
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

func retrieveDataMembersFromSchemaType(schemaType *SchemaType) (dataMembers []*DataMember, requiredMembers []*RequiredMember, constantDataMembers []*DataMember, hasSerializedNameDataMember bool) {
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
			name = strcase.ToScreamingSnake(name)
			constVal = convertConstValueToJavaReadable(posVal[0].StringFormat, property.typeName)

			constDataMember := DataMember{
				Name:          name,
				MemberType:    propertyToJavaType(property),
				Description:   description,
				ConstantVal:   constVal,
			}

			constantDataMembers = append(constantDataMembers, &constDataMember)

		} else {
			var serializedOverrideName string
			pascalCaseName := strcase.ToCamel(name)
			if isSnakeCase(name) {
				serializedOverrideName = name
				name = strcase.ToLowerCamel(name)
				hasSerializedNameDataMember = true
			}
			dataMember := DataMember{
				Name:          name,
				PascalCaseName: pascalCaseName,
				MemberType:    propertyToJavaType(property),
				Description:   description,
				ConstantVal:   constVal,
				SerializedNameOverride: serializedOverrideName,
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
	
	return dataMembers, requiredMembers, constantDataMembers, hasSerializedNameDataMember
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

func convertConstValueToJavaReadable(constVal string, constType string) string {
	if constType == "string" {
		constVal = "\"" + constVal + "\""
	}
	return constVal
}

func isSnakeCase(name string) bool {
	var isSnakeCase bool
	wordArray := strings.Split(name, "_")
	isSnakeCase = len(wordArray) > 1
	return isSnakeCase
}