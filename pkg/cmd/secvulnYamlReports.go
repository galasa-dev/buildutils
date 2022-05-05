/*
 * Copyright contributors to the Galasa project
 */

package cmd

// Security vulnerability report
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

// Acceptance report
type AcceptanceYamlReport struct {
	Cves []Cve `yaml:"cves"`
}

type Cve struct {
	Cve        string `yaml:"cve"`
	Comment    string `yaml:"comment"`
	ReviewDate string `yaml:"reviewDate"`
}
