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
	"strings"
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

	cves     []MdCveStruct
	projects []MdProjectStruct

	cveScoreGroups [][]MdCveStruct

	cveSummary     []CveSummary
	projectSummary []ProjSummary
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
		fmt.Printf("Unable to find the acceptance report at %s\n", secvulnReportAcceptance)
		panic(err)
	}

	for _, yamlReport := range yamlReports {

		consolidateIntoCveStructs(yamlReport)
		consolidateIntoProjectStructs(yamlReport)

	}

	fmt.Printf("%v CVEs to report\n", len(cves))
	fmt.Printf("%v vulnerable Galasa projects to report\n", len(projects))

	sortCveStructs()
	sortProjectStructs()

	formSummarySection()

	writeMarkdown()
}

func unmarshalSecVulnYamlReports(directory string) (SecVulnYamlReport, error) {
	var yamlReport SecVulnYamlReport

	yamlFile, err := os.ReadFile(directory)
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
    comment: To be reviewed
	reviewDate: 01/01/2023
  - cve: CVE-2
    comment: Not applicable
*/
func getAcceptanceYamlReport() (AcceptanceYamlReport, error) {
	var acceptanceReport AcceptanceYamlReport

	response, err := http.Get(secvulnReportAcceptance)
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

// func getAcceptanceYamlReport() (AcceptanceYamlReport, error) {
// 	var yamlReport AcceptanceYamlReport

// 	yamlFile, err := os.ReadFile(secvulnReportAcceptance)
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
func consolidateIntoCveStructs(yaml SecVulnYamlReport) {

	for _, yamlVuln := range yaml.Vulnerabilities {

		cve := yamlVuln.Cve

		// Returns -1 if not found or the index if found, as index is needed to get the existing struct later
		index := cveListedAtTopLevel(cve, cves)
		if index == -1 {

			// This CVE does not have a struct

			var mdVulnArtifacts []MdVulnArtifact
			mdVulnArtifacts = updateVulnStructs(yamlVuln, mdVulnArtifacts)

			comment, reviewDate := getAcceptanceData(cve)

			mdCveStruct := &MdCveStruct{
				Cve:                 cve,
				CvssScore:           yamlVuln.CvssScore,
				Severity:            getSeverity(yamlVuln.CvssScore),
				Link:                yamlVuln.Reference,
				Comment:             comment,
				ReviewDate:          reviewDate,
				VulnerableArtifacts: mdVulnArtifacts,
			}

			cves = append(cves, *mdCveStruct)

		} else if index != -1 {

			// This CVE has a struct already

			existingCveStruct := cves[index]

			existingCveStruct.VulnerableArtifacts = updateVulnStructs(yamlVuln, existingCveStruct.VulnerableArtifacts)

		}

	}

}

func updateVulnStructs(yamlVuln Vulnerability, vulnStructs []MdVulnArtifact) []MdVulnArtifact {

	for _, yamlArtifact := range yamlVuln.VulnerableArtifacts {

		yamlVulnGroupArtifactVersion := getGroupArtifactVersion(yamlArtifact.VulnerableArtifact)

		index1 := vulnerabilityListed(yamlVulnGroupArtifactVersion, vulnStructs)
		if index1 == -1 {

			// This vulnerability isn't listed

			var mdProjects []MdProject

			mdProjects = updateProjectStruct(yamlArtifact, mdProjects)

			mdVulnArtifact := &MdVulnArtifact{
				VulnName: yamlVulnGroupArtifactVersion,
				Projects: mdProjects,
			}

			vulnStructs = append(vulnStructs, *mdVulnArtifact)

		} else if index1 != -1 {

			// This vulnerability is already listed

			existingVulnArtifactStruct := vulnStructs[index1]

			existingVulnArtifactStruct.Projects = updateProjectStruct(yamlArtifact, existingVulnArtifactStruct.Projects)

			vulnStructs[index1] = existingVulnArtifactStruct

		}

	}

	return vulnStructs

}

func updateProjectStruct(yamlArtifact VulnerableArtifact, projectStructs []MdProject) []MdProject {

	for _, yamlProject := range yamlArtifact.DirectProjects {

		galasaGroupArtifact := getGroupAndArtifact(yamlProject.ProjectName)

		index2 := projectListed(galasaGroupArtifact, projectStructs)
		if index2 == -1 {

			// This project isn't listed

			mdProject := &MdProject{
				Name:            galasaGroupArtifact,
				DependencyChain: getShortenedDepChain(yamlProject.DependencyChain),
			}

			projectStructs = append(projectStructs, *mdProject)

		}

	}

	return projectStructs

}

func consolidateIntoProjectStructs(yaml SecVulnYamlReport) {

	for _, yamlVuln := range yaml.Vulnerabilities {

		cve := yamlVuln.Cve

		for _, yamlArtifact := range yamlVuln.VulnerableArtifacts {

			for _, directProject := range yamlArtifact.DirectProjects {

				galasaGroupArtifact := getGroupAndArtifact(directProject.ProjectName)

				index := projectListedAtTopLevel(galasaGroupArtifact, projects)
				if index == -1 {

					// Project doesn't have a struct

					var mdCves []MdCve

					mdCve := &MdCve{
						Cve:             cve,
						CvssScore:       yamlVuln.CvssScore,
						Severity:        getSeverity(yamlVuln.CvssScore),
						DependencyChain: getShortenedDepChain(directProject.DependencyChain),
					}

					mdCves = append(mdCves, *mdCve)

					var dependents []string
					for _, proj := range directProject.TransientProjects {
						dependents = append(dependents, proj.ProjectName)
					}

					mdProjectStruct := &MdProjectStruct{
						Name:       galasaGroupArtifact,
						Dependents: dependents,
						Cves:       mdCves,
					}

					projects = append(projects, *mdProjectStruct)

				} else if index != -1 {

					// This project already has a struct
					// A different scope and/or version may have been processed so need to make sure all info is listed

					existingProjectStruct := projects[index]

					for _, dep := range directProject.TransientProjects {
						if arrayContainsString(dep.ProjectName, existingProjectStruct.Dependents) == false {
							existingProjectStruct.Dependents = append(existingProjectStruct.Dependents, dep.ProjectName)
						}
					}

					index1 := cveListed(cve, existingProjectStruct.Cves)

					if index1 == -1 {

						// This CVE is not listed

						mdCve := &MdCve{
							Cve:             cve,
							CvssScore:       yamlVuln.CvssScore,
							Severity:        getSeverity(yamlVuln.CvssScore),
							DependencyChain: getShortenedDepChain(directProject.DependencyChain),
						}

						existingProjectStruct.Cves = append(existingProjectStruct.Cves, *mdCve)

					}

					projects[index] = existingProjectStruct

				}

			}

		}

	}

}

func sortCveStructs() {

	// Highest to lowest Cvss Score
	sort.Slice(cves, func(i, j int) bool {
		return cves[i].CvssScore > cves[j].CvssScore
	})

	// If Cvss Score is the same, then CVEs in alphabetical order
	newCves := cves
	cves = nil
	var scoreGroup []MdCveStruct
	scoreGroup = append(scoreGroup, newCves[0])
	for i := 1; i <= len(newCves); i++ {
		if i == len(newCves) {
			sort.Slice(scoreGroup, func(x, y int) bool {
				return scoreGroup[x].Cve < scoreGroup[y].Cve
			})
			cves = append(cves, scoreGroup...)
			break
		}
		if newCves[i].CvssScore == newCves[i-1].CvssScore {
			scoreGroup = append(scoreGroup, newCves[i])
		} else if newCves[i].CvssScore != newCves[i-1].CvssScore {
			sort.Slice(scoreGroup, func(x, y int) bool {
				return scoreGroup[x].Cve < scoreGroup[y].Cve
			})
			cves = append(cves, scoreGroup...)
			scoreGroup = nil
			scoreGroup = append(scoreGroup, newCves[i])
		}
	}

	// Vulnerable artifacts in alphabetical order within each CVE
	for _, cve := range cves {
		sort.Slice(cve.VulnerableArtifacts, func(i, j int) bool {
			return cve.VulnerableArtifacts[i].VulnName < cve.VulnerableArtifacts[j].VulnName
		})

		// Projects in alphabetical order within each vulnerable artifact
		for _, vulnArtifact := range cve.VulnerableArtifacts {
			sort.Slice(vulnArtifact.Projects, func(i, j int) bool {
				return vulnArtifact.Projects[i].Name < vulnArtifact.Projects[j].Name
			})
		}
	}

}

func sortProjectStructs() {

	// Projects in alphabetical order
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	// CVEs within each project by order of Cvss Score
	for _, proj := range projects {
		sort.Slice(proj.Cves, func(i, j int) bool {
			return proj.Cves[i].CvssScore > proj.Cves[j].CvssScore
		})
	}

	newProjects := projects
	projects = nil
	// If Cvss Score is the same, then CVEs in alphabetical order
	for _, project := range newProjects {
		if len(project.Cves) > 1 {

			var cves []MdCve
			var scoreGroup []MdCve
			scoreGroup = append(scoreGroup, project.Cves[0])
			for i := 1; i <= len(project.Cves); i++ {
				if i == len(project.Cves) {
					sort.Slice(scoreGroup, func(x, y int) bool {
						return scoreGroup[x].Cve < scoreGroup[y].Cve
					})
					cves = append(cves, scoreGroup...)
					break
				}
				if project.Cves[i].CvssScore == project.Cves[i-1].CvssScore {
					scoreGroup = append(scoreGroup, project.Cves[i])
				} else if project.Cves[i].CvssScore != project.Cves[i-1].CvssScore {
					sort.Slice(scoreGroup, func(x, y int) bool {
						return scoreGroup[x].Cve < scoreGroup[y].Cve
					})
					cves = append(cves, scoreGroup...)
					scoreGroup = nil
					scoreGroup = append(scoreGroup, project.Cves[i])
				}
			}
			project.Cves = cves
			projects = append(projects, project)

		}

	}

}

func formSummarySection() {
	for _, cve := range cves {
		var allProjs []string
		for _, vuln := range cve.VulnerableArtifacts {
			for _, proj := range vuln.Projects {
				allProjs = append(allProjs, proj.Name)
			}
		}
		allProjs = removeDuplicates(allProjs)
		cveSum := &CveSummary{
			Cve:      cve.Cve,
			Link:     cve.Link,
			Severity: getSeverity(cve.CvssScore),
			Amount:   len(allProjs),
		}
		cveSummary = append(cveSummary, *cveSum)
	}

	for _, proj := range projects {
		highCount := 0
		otherCount := 0
		for _, cve := range proj.Cves {
			if cve.Severity == "Critical" || cve.Severity == "High" {
				highCount++
			} else {
				otherCount++
			}
		}
		projSum := &ProjSummary{
			Project:    proj.Name,
			High:       highCount,
			Other:      otherCount,
			Dependents: len(proj.Dependents),
		}
		projectSummary = append(projectSummary, *projSum)
	}
}

func writeMarkdown() {

	// Create file to export Markdown page
	markdownFile, err := os.Create(secvulnReportOutput)
	if err != nil {
		fmt.Printf("Unable to create a file for the Markdown report, %v\n", err)
		panic(err)
	}
	defer markdownFile.Close()

	_, err = markdownFile.WriteString("# Security vulnerabilites in deployed Galasa\n\n## Vulnerabilities\n\n### Summary\n\n")
	if err != nil {
		fmt.Printf("Unable to write to the Markdown file, %v\n", err)
		panic(err)
	}

	// Section 1
	cveSummaryTemplate := "- [{{.Cve}}]({{.Link}}) - {{.Severity}} - {{if ( gt .Amount 1)}}{{.Amount}} projects{{else}}1 project{{end}}\n"
	cveSummaryTmpl, err := template.New("cveSummaryTemplate").Parse(cveSummaryTemplate)
	if err != nil {
		fmt.Printf("Unable to create the template for the CVE summary section of the Markdown, %v\n", err)
		panic(err)
	}

	for _, cveSum := range cveSummary {
		err = cveSummaryTmpl.Execute(markdownFile, cveSum)
	}

	cveTemplate := "\n### {{.Cve}}\n\nSeverity: **{{.Severity}}**\n\n{{if .ReviewDate}}Acceptance: To be reviewed {{.ReviewDate}}\n\n{{end}}{{if .Comment}}Acceptance: {{.Comment}}\n\n{{end}}[Link]({{.Link}})\n\nVulnerable artifacts:\n\n"
	cveTmpl, err := template.New("cveTemplate").Parse(cveTemplate)
	if err != nil {
		fmt.Printf("Unable to create the template for the CVE main section of the Markdown, %v\n", err)
		panic(err)
	}

	cveArtifactTemplate := "{{.VulnName}}\n\nGalasa Projects/Images directly affected:\n\n"
	cveArtifactTmpl, err := template.New("cveArtifactTemplate").Parse(cveArtifactTemplate)
	if err != nil {
		fmt.Printf("Unable to create the template for the CVE artifact section of the Markdown, %v\n", err)
		panic(err)
	}

	cveProjTemplate := "- {{.Name}}\n{{ range .DependencyChain}}  - {{.}}\n{{end}}\n"
	cveProjTmpl, err := template.New("cveProjTemplate").Parse(cveProjTemplate)
	if err != nil {
		fmt.Printf("Unable to create the template for the CVE projects section of the Markdown, %v\n", err)
		panic(err)
	}

	for _, cve := range cves {
		err = cveTmpl.Execute(markdownFile, cve)
		for _, vuln := range cve.VulnerableArtifacts {
			err = cveArtifactTmpl.Execute(markdownFile, vuln)
			for _, proj := range vuln.Projects {
				err = cveProjTmpl.Execute(markdownFile, proj)
			}
		}
	}

	// Section 2
	_, err = markdownFile.WriteString("\n\n\n## Galasa Projects/Images\n\n### Summary\n\n")
	if err != nil {
		fmt.Printf("Unable to write to the Markdown file, %v\n", err)
		panic(err)
	}

	projSummaryTemplate := "- {{.Project}} - {{.High}} High+{{if (gt .Other 0)}}, {{.Other}} Other{{end}}{{if (gt .Dependents 0)}}, {{.Dependents}} dependents{{end}}\n"
	projSummaryTmpl, err := template.New("projSummaryTemplate").Parse(projSummaryTemplate)
	if err != nil {
		fmt.Printf("Unable to create the template for the Project summary section of the Markdown, %v\n", err)
		panic(err)
	}

	for _, projSum := range projectSummary {
		err = projSummaryTmpl.Execute(markdownFile, projSum)
	}

	projectTemplate := "\n### {{.Name}}\n\n"
	projectTmpl, err := template.New("projectTemplate").Parse(projectTemplate)
	if err != nil {
		fmt.Printf("Unable to create the template for the Project name section of the Markdown, %v\n", err)
		panic(err)
	}

	projCvesTemplate := "- {{.Cve}} - **{{.Severity}}**\n{{ range .DependencyChain}}  - {{.}}\n{{end}}\n"
	projCvesTmpl, err := template.New("projCvesTemplate").Parse(projCvesTemplate)
	if err != nil {
		fmt.Printf("Unable to create the template for the Project CVEs section of the Markdown, %v\n", err)
		panic(err)
	}

	for _, proj := range projects {
		err = projectTmpl.Execute(markdownFile, proj)
		for _, cve := range proj.Cves {
			err = projCvesTmpl.Execute(markdownFile, cve)
		}
	}

	fmt.Printf("Markdown page exported to %s\n", secvulnReportOutput)

}

func getShortenedDepChain(depChain string) []string {
	chain := strings.Split(depChain, " -> ")
	var shorterDepChain []string
	for _, submatch := range chain {
		groupArtifactVersion := getGroupArtifactVersion(submatch)
		shorterDepChain = append(shorterDepChain, groupArtifactVersion)
	}
	return shorterDepChain
}

func getGroupArtifactVersion(fullString string) string {
	submatches := getRegexSubmatches(fullString)

	group := submatches[0]
	artifact := submatches[1]
	version := submatches[3]

	return fmt.Sprintf("%s:%s:%s", group, artifact, version)
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
		return "None"
	} else if cvssScore >= 0.1 && cvssScore <= 3.9 {
		return "Low"
	} else if cvssScore >= 4.0 && cvssScore <= 6.9 {
		return "Medium"
	} else if cvssScore >= 7.0 && cvssScore <= 8.9 {
		return "High"
	} else if cvssScore >= 9.0 && cvssScore <= 10.0 {
		return "Critical"
	} else {
		return ""
	}
}

func cveListedAtTopLevel(targetCve string, array []MdCveStruct) int {
	for index, cve := range array {
		if cve.Cve == targetCve {
			return index
		}
	}
	return -1
}

func vulnerabilityListed(targetVuln string, array []MdVulnArtifact) int {
	for index, vuln := range array {
		if vuln.VulnName == targetVuln {
			return index
		}
	}
	return -1
}

func projectListed(targetProject string, array []MdProject) int {
	for index, project := range array {
		if project.Name == targetProject {
			return index
		}
	}
	return -1
}

func projectListedAtTopLevel(targetProject string, array []MdProjectStruct) int {
	for index, project := range array {
		if project.Name == targetProject {
			return index
		}
	}
	return -1
}

func cveListed(targetCve string, array []MdCve) int {
	for index, cve := range array {
		if cve.Cve == targetCve {
			return index
		}
	}
	return -1
}
