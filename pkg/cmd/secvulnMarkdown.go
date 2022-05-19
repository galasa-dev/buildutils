/*
 * Copyright contributors to the Galasa project
 */

package cmd

type MarkdownStruct struct {
	Cve            string
	CvssScore      float64 // for sorting
	Severity       string
	DirectProjects []DirectProject
	Comment        string
	ReviewDate     string
}
