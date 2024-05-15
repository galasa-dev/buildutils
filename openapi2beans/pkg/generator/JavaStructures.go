/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package generator

import (
	"sort"
	"strings"
)

type JavaPackage struct {
	Name            string
	Classes         map[string]*JavaClass
	Enums           map[string]*JavaEnum
}

func NewJavaPackage(name string) *JavaPackage {
	javaPackage := JavaPackage{
		Name:            name,
		Classes:         make(map[string]*JavaClass),
		Enums:           make(map[string]*JavaEnum),
	}
	return &javaPackage
}

type JavaClass struct {
	Name                string
	Description         []string
	JavaPackage         *JavaPackage
	DataMembers         []*DataMember
	RequiredMembers     []*RequiredMember
	ConstantDataMembers []*DataMember
}

func NewJavaClass(name string, description []string, javaPackage *JavaPackage, dataMembers []*DataMember, requiredMembers []*RequiredMember, constantDataMembers []*DataMember) *JavaClass {
	javaClass := JavaClass{
		Name:                name,
		Description:         description,
		JavaPackage:         javaPackage,
		DataMembers:         dataMembers,
		RequiredMembers:     requiredMembers,
		ConstantDataMembers: constantDataMembers,
	}
	return &javaClass
}

// sorts DataMembers, RequiredMembers, and ConstantDataMembers.
// order is:
// boolean > int > double > String > other
func (class JavaClass) Sort() {
	sort.SliceStable(class.DataMembers, func(i int, j int) bool { return isDataMemberLessThanComparison(class.DataMembers[i], class.DataMembers[j]) })
	sort.SliceStable(class.ConstantDataMembers, func(i int, j int) bool { return isDataMemberLessThanComparison(class.ConstantDataMembers[i], class.ConstantDataMembers[j]) })
	if class.RequiredMembers != nil {
		class.RequiredMembers[0].IsFirst = false
	}
	sort.SliceStable(class.RequiredMembers, func(i int, j int) bool { return isDataMemberLessThanComparison(class.RequiredMembers[i].DataMember, class.RequiredMembers[j].DataMember) })
	if class.RequiredMembers != nil {
		class.RequiredMembers[0].IsFirst = true
	}
}

type DataMember struct {
	Name          string
	CamelCaseName string
	MemberType    string
	Description   []string
	Required      bool
	ConstantVal   string
}

func (dataMember DataMember) IsConstant() bool {
	return dataMember.ConstantVal != ""
}

type RequiredMember struct {
	IsFirst    bool
	DataMember *DataMember
}

type JavaEnum struct {
	Name        string
	Description []string
	EnumValues  []string
	JavaPackage *JavaPackage
}

func NewJavaEnum(name string, description []string, enumValues []string, javaPackage *JavaPackage) *JavaEnum {
	javaEnum := JavaEnum{
		Name:        name,
		Description: description,
		EnumValues:  enumValues,
		JavaPackage: javaPackage,
	}
	return &javaEnum
}

func (enum JavaEnum) Sort() {
	sort.SliceStable(enum.EnumValues, func(i, j int) bool { return enum.EnumValues[i] < enum.EnumValues[j] })
}




// function used for sorting; groups variables by type and then alphabetically
// order of variables is:
// boolean > int > double > String > other
func isDataMemberLessThanComparison(dataMember *DataMember, comparisonMember *DataMember) bool {
	less := true
	switch memberType := dataMember.MemberType; {
	case strings.Contains(memberType, "boolean"):
		switch comparisonMemberTpye := comparisonMember.MemberType; {
		case strings.Contains(comparisonMemberTpye, "boolean"):
			less = dataMember.Name > comparisonMember.Name
		default:
			less = true
		}
	case strings.Contains(memberType, "int"):
		switch comparisonMember.MemberType {
		case "boolean":
			less = false
		case "int":
			less = dataMember.Name > comparisonMember.Name
		default:
			less = true
		}
	case strings.Contains(memberType, "double"):
		switch comparisonMemberType := comparisonMember.MemberType; {
		case strings.Contains(comparisonMemberType, "boolean"), strings.Contains(comparisonMemberType, "int"):
			less = false
		case strings.Contains(comparisonMemberType, "double"):
			less = dataMember.Name > comparisonMember.Name
		default:
			less = true
		}
	case strings.Contains(memberType, "String"):
		switch comparisonMemberType := comparisonMember.MemberType; {
		case strings.Contains(comparisonMemberType, "boolean"), strings.Contains(comparisonMemberType, "int"), strings.Contains(comparisonMemberType, "double"):
			less = false
		case strings.Contains(comparisonMemberType, "String"):
			less = dataMember.Name > comparisonMember.Name
		default:
			less = true
		}
	default:
		if dataMember.MemberType == comparisonMember.MemberType {
			less = dataMember.Name > comparisonMember.Name
		} else {
			less = dataMember.MemberType > comparisonMember.MemberType
		}
	}
	return less
}
