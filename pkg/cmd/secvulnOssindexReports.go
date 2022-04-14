/*
 * Copyright contributors to the Galasa project
 */

package cmd

type YamlReport struct {
	Vulnerabilities []Vulnerability `yaml:"vulnerabilities"`
}

type Vulnerability struct {
	Cve      string    `yaml:"cve"`
	Projects []Project `yaml:"projects"`
}

type Project struct {
	Project string `yaml:"name"`
	// DependencyType string `yaml:"dependencyType"`
	// DependencyChain *DependencyChain // Pointer to DependencyChain as not needed if direct dependency
}

type DependencyChain struct {
}
