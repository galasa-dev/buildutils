/*
 * Copyright contributors to the Galasa project
 */

package cmd

import "encoding/xml"

type Pom struct {
	XMLName xml.Name `xml:"project"`

	ModelVersion string `xml:"modelVersion"`

	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
	Packaging  string `xml:"packaging"`

	Parent *Parent

	Dependencies *Dependencies

	Modules *Modules

	Build *Build
}
type Parent struct {
	XMLName xml.Name `xml:"parent"`

	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

type Dependencies struct {
	XMLName xml.Name `xml:"dependencies"`

	Dependencies []Dependency `xml:"dependency"`
}
type Dependency struct {
	XMLName xml.Name `xml:"dependency"`

	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

type Modules struct {
	XMLName xml.Name `xml:"modules"`

	Module []string `xml:"module"`
}

type Build struct {
	XMLName xml.Name `xml:"build"`

	Plugins Plugins `xml:"plugins"`
}
type Plugins struct {
	XMLName xml.Name `xml:"plugins"`

	Plugin Plugin `xml:"plugin"`
}
type Plugin struct {
	XMLName xml.Name `xml:"plugin"`

	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`

	Executions Executions `xml:"executions"`

	Configuration Configuration `xml:"configuration"`
}
type Executions struct {
	XMLName xml.Name `xml:"executions"`

	Execution Execution `xml:"execution"`
}
type Execution struct {
	XMLName xml.Name `xml:"execution"`
	Id      string   `xml:"id"`
	Phase   string   `xml:"phase"`

	Goals Goals `xml:"goals"`
}
type Goals struct {
	XMLName xml.Name `xml:"goals"`

	Goal string `xml:"goal"`
}
type Configuration struct {
	XMLName xml.Name `xml:"configuration"`

	ReportFile string `xml:"reportFile"`

	Fail string `xml:"fail"`
}
