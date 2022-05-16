/*
 * Copyright contributors to the Galasa project
 */

package cmd

type ReportStruct struct {
	Cve              string
	Severity         string
	GalasaProject    string
	DependencyChains []string
	Comment          string
	ReviewDate       string
}
