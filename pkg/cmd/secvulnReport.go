/*
 * Copyright contributors to the Galasa project
 */

package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"

	"text/template"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	secvulnReportCmd = &cobra.Command{
		Use:   "report",
		Short: "Generate report",
		Long:  "Generate report",
		Run:   secvulnReportExecute,
	}

	secvulnReportExtracts   *[]string
	secvulnReportAcceptance string
	secvulnReportOutput     string

	acceptanceReport AcceptanceYamlReport

	markdownStructs []MarkdownStruct
)

func init() {
	secvulnReportExtracts = secvulnReportCmd.PersistentFlags().StringArray("extract", nil, "Extract yaml files")
	secvulnReportCmd.PersistentFlags().StringVar(&secvulnReportAcceptance, "acceptance", "", "Acceptance yaml URL")
	secvulnReportCmd.PersistentFlags().StringVar(&secvulnReportOutput, "output", "", "Output markdown file")

	secvulnReportCmd.MarkPersistentFlagRequired("extract")
	secvulnReportCmd.MarkPersistentFlagRequired("acceptance")
	secvulnReportCmd.MarkPersistentFlagRequired("output")

	secvulnCmd.AddCommand(secvulnReportCmd)
}

func secvulnReportExecute(cmd *cobra.Command, args []string) {
	fmt.Printf("Galasa Build - Security Vulnerability Report - version %v\n", rootCmd.Version)

	var err error

	var yamlReports []SecVulnYamlReport
	for _, directory := range *secvulnReportExtracts {
		yamlReport, err := unmarshalSecVulnYamlReports(directory)
		if err != nil {
			fmt.Printf("Unable to read and unmarshal the security vulnerability report in directory %s, %v\n", directory, err)
			panic(err)
		}
		yamlReports = append(yamlReports, yamlReport)
	}

	// Get the acceptance report from the project management repo to merge in manager's comments and review dates with the Markdown page
	acceptanceReport, err = getAcceptanceYamlReport()
	if err != nil {
		fmt.Printf("Unable to find the acceptance report at %s/%s\n", secvulnReportAcceptance, "override.yaml")
		panic(err)
	}

	for _, yamlReport := range yamlReports {
		consolidateIntoStructs(yamlReport)
	}

	writeMarkdown()
}

func unmarshalSecVulnYamlReports(directory string) (SecVulnYamlReport, error) {
	var yamlReport SecVulnYamlReport

	yamlFile, err := os.ReadFile(fmt.Sprintf("%s/%s", directory, "galasa-secvuln-report-new.yaml"))
	if err != nil {
		return yamlReport, err
	}

	err = yaml.Unmarshal(yamlFile, &yamlReport)
	if err != nil {
		return yamlReport, err
	}

	return yamlReport, err
}

/* Function is meant to pull 'override.yaml' file from the Galasa project management
repo raw github pages with management's comments and/or review dates for certain vulnerabilities
to show users we are aware of certain vulnerabilities, they are being dealt with, etc
override.yaml has not been written yet so no comments are pulled in currently

Can use commented out function below to pull a dummy report if it is the correct format
cves:
  - cve: CVE-1
    comment: bob
    reviewDate: 01/01/2022
  - cve: CVE-2
    comment: fred
    reviewDate: 01/01/2022
*/
func getAcceptanceYamlReport() (AcceptanceYamlReport, error) {
	var acceptanceReport AcceptanceYamlReport

	url := fmt.Sprintf("%s/%s", secvulnReportAcceptance, "override.yaml")
	response, err := http.Get(url)
	if err != nil {
		return acceptanceReport, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return acceptanceReport, err
	}

	err = yaml.Unmarshal(body, &acceptanceReport)
	if err != nil {
		return acceptanceReport, err
	}

	return acceptanceReport, err
}

// func getAcceptanceYamlReport(directory string) (AcceptanceYamlReport, error) {
// 	var yamlReport AcceptanceYamlReport

// 	yamlFile, err := os.ReadFile(fmt.Sprintf("%s/%s", directory, "override.yaml"))
// 	if err != nil {
// 		return yamlReport, err
// 	}

// 	err = yaml.Unmarshal(yamlFile, &yamlReport)
// 	if err != nil {
// 		return yamlReport, err
// 	}

// 	return yamlReport, err
// }

/* As there may be multiple security vulnerabilty reports in yaml, they all must be consolidated
into structs so the same information is not duplicated when the Markdown page is written
*/
func consolidateIntoStructs(yaml SecVulnYamlReport) {

	// Iterate through each CVE entry in the Yaml
	for _, vuln := range yaml.Vulnerabilities {

		cve := vuln.Cve
		cvssScore := vuln.CvssScore
		severity := getSeverity(cvssScore)
		// Comment and review date can be "" if the vulnerability is new
		comment, reviewDate := getAcceptanceData(cve)

		if containsCve(cve, markdownStructs) == false {
			// This CVE does not yet have a struct (therefore neither do it's direct projects and transient projects) so make one
			var directProjs []DirectProject
			for _, directProj := range vuln.DirectProjects {
				projectName := directProj.ProjectName
				depChain := directProj.DependencyChain
				var transProjs []TransientProject
				for _, transientProj := range directProj.TransientProjects {
					transientProjName := transientProj.ProjectName
					transientDepChain := transientProj.DependencyChain
					transientProject := TransientProject{transientProjName, transientDepChain}
					transProjs = append(transProjs, transientProject)
				}
				directProjectStruct := DirectProject{projectName, depChain, transProjs}
				directProjs = append(directProjs, directProjectStruct)
			}
			markdownStruct := MarkdownStruct{cve, cvssScore, severity, directProjs, comment, reviewDate}
			markdownStructs = append(markdownStructs, markdownStruct)

		} else if containsCve(cve, markdownStructs) == true {
			// Already a struct for this CVE

			// Find existing Markdown struct for this CVE, index is needed later
			for index, markdownStruct := range markdownStructs {
				if markdownStruct.Cve == cve {

					// Now iterate through the directly affected Galasa projects listed under this CVE in the Yaml
					for _, directProj := range vuln.DirectProjects {
						directProjectName := directProj.ProjectName

						if containsDirectProj(directProjectName, markdownStruct.DirectProjects) == false {
							// This directly affected Galasa project is not listed in this CVE's struct so add it
							depChain := directProj.DependencyChain
							var transientProjs []TransientProject
							for _, transientProj := range directProj.TransientProjects {
								transientProjName := transientProj.ProjectName
								transientDepChain := transientProj.DependencyChain
								transientProject := TransientProject{transientProjName, transientDepChain}
								transientProjs = append(transientProjs, transientProject)
							}
							directProjectStruct := DirectProject{directProjectName, depChain, transientProjs}
							// Remove old list of directly affected projects and replace with new one
							newArray := append(markdownStruct.DirectProjects, directProjectStruct)
							newStruct := MarkdownStruct{cve, cvssScore, severity, newArray, comment, reviewDate}
							markdownStructs = removeMD(markdownStructs, index)
							markdownStructs = append(markdownStructs, newStruct)

						} else if containsDirectProj(directProjectName, markdownStruct.DirectProjects) == true {
							// Already a struct for this CVE which lists this directly affected Galasa project
							// but need to check all of the indirectly affected projects are listed too

							// Find existing Direct Project struct, index is needed later
							for index2, directProject := range markdownStruct.DirectProjects {
								if directProject.ProjectName == directProjectName {

									// Iterate through the transient projects from Yaml report
									for _, transientProj := range directProj.TransientProjects {
										transientProjName := transientProj.ProjectName

										if containsTransientProj(transientProjName, directProject.TransientProjects) == false {
											// This transient project is not in the struct so add it
											transientDepChain := transientProj.DependencyChain
											transientProjectStruct := TransientProject{transientProjName, transientDepChain}
											newArray := append(directProject.TransientProjects, transientProjectStruct)
											newStruct := DirectProject{directProject.ProjectName, directProject.DependencyChain, newArray}
											markdownStruct.DirectProjects = removePrj(markdownStruct.DirectProjects, index2)
											markdownStruct.DirectProjects = append(markdownStruct.DirectProjects, newStruct)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

}

func writeMarkdown() {

	// Create file to export Markdown page
	markdownFile, err := os.Create(fmt.Sprintf("%s/%s", secvulnReportOutput, "galasa-secvuln-report-new2.md"))
	if err != nil {
		fmt.Printf("Unable to create a file for the Markdown report, %v\n", err)
		panic(err)
	}
	defer markdownFile.Close()

	_, err = markdownFile.WriteString("# Galasa Security Vulnerability report")
	if err != nil {
		fmt.Printf("Unable to write to the Markdown file, %v\n", err)
		panic(err)
	}

	cveHeaderTemplate := "\n\n\n## CVE: {{.Cve}}\n### Severity: {{.Severity}}{{if .Comment}}\n#### Comment: {{ .Comment }}{{end}}{{if .ReviewDate }}\n#### Review date: {{ .ReviewDate }}{{end}}"
	cveHeaderTmpl, err := template.New("cveHeaderTemplate").Parse(cveHeaderTemplate)
	if err != nil {
		fmt.Printf("Unable to create the template for the CVE Header section of the Markdown, %v\n", err)
		panic(err)
	}

	cveTemplate := "\n\n### Directly affected Galasa artifact: {{.ProjectName}}\n#### Dependency chain:\n* {{ .DependencyChain }}{{if .TransientProjects}}\n> > #### Indirectly affected artifacts that use {{.ProjectName}}:\n{{end}}"
	cveTmpl, err := template.New("cveTemplate").Parse(cveTemplate)
	if err != nil {
		fmt.Printf("Unable to create the template for the direct projects section of the Markdown, %v\n", err)
		panic(err)
	}

	affectedProjTemplate := "> > * {{.ProjectName}}\n"
	affectedTmpl, err := template.New("affectedProjTemplate").Parse(affectedProjTemplate)
	if err != nil {
		fmt.Printf("Unable to create the template for the affected projects section of the Markdown, %v\n", err)
		panic(err)
	}

	// Write to Markdown page in order of critical to low CVEs
	sortedCvss := sortCvss()

	index := 0
	for index < len(sortedCvss) {
		highestCvss := sortedCvss[index]

		for _, markdownStruct := range markdownStructs {
			if markdownStruct.CvssScore == highestCvss {
				err = cveHeaderTmpl.Execute(markdownFile, markdownStruct)
				if err != nil {
					fmt.Printf("Unable to apply the template to the first structs, %v\n", err)
					panic(err)
				}
				for _, directProjects := range markdownStruct.DirectProjects {
					err = cveTmpl.Execute(markdownFile, directProjects)
					if err != nil {
						fmt.Printf("Unable to apply the template to the 2nd structs, %v\n", err)
						panic(err)
					}
					for _, transientProjects := range directProjects.TransientProjects {
						err = affectedTmpl.Execute(markdownFile, transientProjects)
						if err != nil {
							fmt.Printf("Unable to apply the template to the 3rd structs, %v\n", err)
							panic(err)
						}
					}
				}
				index++
			}
		}

	}

	fmt.Printf("Markdown page exported to %s\n", secvulnReportOutput)

}

func sortCvss() []float64 {
	var cvssScores []float64
	for _, mdstruct := range markdownStructs {
		cvssScores = append(cvssScores, mdstruct.CvssScore)
	}
	sort.Float64s(cvssScores)
	sort.Sort(sort.Reverse(sort.Float64Slice(cvssScores)))
	return cvssScores
}

func getAcceptanceData(cve string) (string, string) {
	for _, value := range acceptanceReport.Cves {
		if value.Cve == cve {
			return value.Comment, value.ReviewDate
		}
	}
	return "", ""
}

func getSeverity(cvssScore float64) string {
	if cvssScore >= 0 && cvssScore < 0.1 {
		return "none"
	} else if cvssScore >= 0.1 && cvssScore <= 3.9 {
		return "low"
	} else if cvssScore >= 4.0 && cvssScore <= 6.9 {
		return "medium"
	} else if cvssScore >= 7.0 && cvssScore <= 8.9 {
		return "high"
	} else if cvssScore >= 9.0 && cvssScore <= 10.0 {
		return "critical"
	} else {
		return ""
	}
}

func removeMD(s []MarkdownStruct, i int) []MarkdownStruct {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func removePrj(s []DirectProject, i int) []DirectProject {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func containsCve(cve string, array []MarkdownStruct) bool {
	for _, element := range array {
		if cve == element.Cve {
			return true
		}
	}
	return false
}

func containsDirectProj(projName string, array []DirectProject) bool {
	for _, element := range array {
		if projName == element.ProjectName {
			return true
		}
	}
	return false
}

func containsTransientProj(projName string, array []TransientProject) bool {
	for _, element := range array {
		if projName == element.ProjectName {
			return true
		}
	}
	return false
}
