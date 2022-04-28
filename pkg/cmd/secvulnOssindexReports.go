/*
 * Copyright contributors to the Galasa project
 */

package cmd

type YamlReport struct {
	Vulnerabilities []Vulnerability `yaml:"cves"`
}

type Vulnerability struct {
	Cve      string    `yaml:"cve"`
	Projects []Project `yaml:"projects"`
}

type Project struct {
	Project         string `yaml:"name"`
	DependencyType  string `yaml:"dependencyType"`
	DependencyChain string `yaml:"dependencyChain"`
}
