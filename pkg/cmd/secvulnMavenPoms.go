package cmd

import "encoding/xml"

// main pom 
type Pom struct {
	XMLName           xml.Name     `xml:"project"`

	Xmlns             string       `xml:"xmlns,attr"`
	XmlnsXsi          string       `xml:"xmlns:xsi,attr"`
	XsiSchemaLocation string       `xml:"xsi:schemaLocation,attr"`

	GroupId           string       `xml:"groupId"`
	ArtifactId        string       `xml:"artifactId"`
	Version           string       `xml:"version"`
	Packaging         string       `xml:"packaging"`

	Dependencies      Dependencies `xml:"dependencies"`
}

// Stripped down psuedo maven project poms
type NewPom struct {
	XMLName           xml.Name     `xml:"project"`

	Xmlns             string       `xml:"xmlns,attr"`
	XmlnsXsi          string       `xml:"xmlns:xsi,attr"`
	XsiSchemaLocation string       `xml:"xsi:schemaLocation,attr"`

	GroupId           string       `xml:"groupId"`
	ArtifactId        string       `xml:"artifactId"`
	Version           string       `xml:"version"`
	Packaging         string       `xml:"packaging"`

	Parent            Parent       `xml:"parent"`

	Dependencies      Dependencies `xml:"dependencies"`
}

// Security scanning parent pom
type SecurityScanningPom struct {
	XMLName           xml.Name     `xml:"project"`

	Xmlns             string       `xml:"xmlns,attr"`
	XmlnsXsi          string       `xml:"xmlns:xsi,attr"`
	XsiSchemaLocation string       `xml:"xsi:schemaLocation,attr"`

	GroupId           string       `xml:"groupId"`
	ArtifactId        string       `xml:"artifactId"`
	Version           string       `xml:"version"`
	Packaging         string       `xml:"packaging"`

	Modules Modules `xml:"modules"`
}

// Elements
type Parent struct {
	XMLName xml.Name `xml:"parent"`

	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

type Dependencies struct {
	XMLName      xml.Name     `xml:"dependencies"`

	Dependencies []Dependency `xml:"dependency"`
}
type Dependency struct {
	XMLName    xml.Name `xml:"dependency"`

	GroupId    string   `xml:"groupId"`
	ArtifactId string   `xml:"artifactId"`
	Version    string   `xml:"version"`
}

type Modules struct {
	XMLName xml.Name `xml:"modules"`

	Module []string `xml:"module"`
}
