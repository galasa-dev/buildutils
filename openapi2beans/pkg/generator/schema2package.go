/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package generator

import (
	"github.com/iancoleman/strcase"
)

func translateSchemaTypesToJavaPackage(schemaTypes map[string]*SchemaType, packageName string) (javaPackage *JavaPackage) {
	javaPackage = NewJavaPackage(packageName)
	for _, schemaType := range schemaTypes {

		if schemaType.ownProperty.IsEnum() {
			enumValues := mapValuesToArray(schemaType.ownProperty.possibleValues)
			javaEnum := NewJavaEnum(strcase.ToCamel(schemaType.name), schemaType.description, enumValues, javaPackage)

			javaPackage.Enums[javaEnum.Name] = javaEnum
		} else {
			dataMembers, requiredMembers, constantDataMembers, hasSerializedNameDataMember := retrieveDataMembersFromSchemaType(schemaType)
			javaClass := NewJavaClass(strcase.ToCamel(schemaType.name), schemaType.description, javaPackage, dataMembers, requiredMembers, constantDataMembers, hasSerializedNameDataMember)
			
			javaPackage.Classes[javaClass.Name] = javaClass
		}
	}
	return javaPackage
}

func mapValuesToArray(inputMap map[string]string) (outputValueArray []string) {
	for _, value := range inputMap {
		outputValueArray = append(outputValueArray, value)
	}
	return outputValueArray
}

func retrieveDataMembersFromSchemaType(schemaType *SchemaType) (dataMembers []*DataMember, requiredMembers []*RequiredMember, constantDataMembers []*DataMember, hasSerializedNameDataMember bool) {
	for _, property := range schemaType.properties {
		name := property.name
		dataMember := NewDataMember(name, propertyToJavaType(property), property.description)

		if property.IsSetInConstructor() {
			requiredMember := RequiredMember{DataMember: dataMember}
			requiredMembers = append(requiredMembers, &requiredMember)
		}
		if property.IsConstant() {
			constVals := mapValuesToArray(property.GetPossibleValues())
			dataMember.Name = strcase.ToScreamingSnake(name)
			dataMember.ConstantVal = convertConstValueToJavaReadable(constVals[0], property.typeName)

			constantDataMembers = append(constantDataMembers, dataMember)
		} else {
			dataMembers = append(dataMembers, dataMember)
		}
		if dataMember.SerializedNameOverride != "" {
			hasSerializedNameDataMember = true
		}
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
