/*
 * Copyright contributors to the Galasa project
 */

package cmd

type MdCveStruct struct {
	Cve                 string
	CvssScore           float64
	Severity            string
	Link                string
	Comment             string
	ReviewDate          string
	VulnerableArtifacts []MdVulnArtifact
}

type MdVulnArtifact struct {
	VulnName string
	Projects []MdProject
}

type MdProject struct {
	Name            string
	DependencyChain []string
}

type MdProjectStruct struct {
	Name       string
	Dependents []string
	Cves       []MdCve
}

type MdCve struct {
	Cve             string
	CvssScore       float64
	Severity        string
	DependencyChain []string
}

// For Summary section of Markdown
type CveSummary struct {
	Cve      string
	Link     string
	Severity string
	Amount   int
}

type ProjSummary struct {
	Project    string
	High       int
	Other      int
	Dependents int
}