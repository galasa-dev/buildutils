package cmd

import "encoding/xml"

type Pom struct {
	XMLName xml.Name `xml:"project"`

	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
	Packaging  string `xml:"packaging"`

	Parent *Parent

	Dependencies *Dependencies

	Modules *Modules
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
