package cmd

import "encoding/xml"

type Pom struct {
	XMLName xml.Name `xml:"project"`

	Xmlns             string `xml:"xmlns,attr"`
	XmlnsXsi          string `xml:"http://maven.apache.org/POM/4.0.0 xsi,attr"`
	XsiSchemaLocation string `xml:"http://www.w3.org/2001/XMLSchema-instance schemaLocation,attr"`

	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
	Packaging  string `xml:"packaging"`

	Parent Parent `xml:"parent,omitempty"`

	Dependencies Dependencies `xml:"dependencies,omitempty"`

	Modules Modules `xml:"modules,omitempty"`
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
