/*
 * Copyright contributors to the Galasa project
 */

package cmd

type MdCveStruct struct {
	Cve        string
	CvssScore  float64 // for sorting
	Severity   string
	Link       string
	Comment    string
	ReviewDate string
	Projects   []MdProject
}

type MdProject struct {
	Artifact        string
	Name            string
	DependencyChain []string
}

type MdProjectStruct struct {
	Artifact   string
	Name       string
	Dependents []string
	Cves       []MdCve
}

type MdCve struct {
	Cve             string
	CvssScore       float64 // for sorting
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
