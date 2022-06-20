/*
 * Copyright contributors to the Galasa project
 */

package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
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

	fmt.Printf("%v vulnerabilities to report\n", len(cves))

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

	for _, vuln := range yaml.Vulnerabilities {

		cve := vuln.Cve

		// Returns -1 if not found or the index if found, as index is needed to get the existing struct later
		index := cveListedAtTopLevel(cve, cves)

		if index == -1 {

			var projectArray []MdProject
			for _, dirProj := range vuln.DirectProjects {

				// A different version or scope of this project might have been processed already so search by artifact
				artifact, shorterString := getShorterString(dirProj.ProjectName)
				index1 := projectListed(artifact, projectArray)
				if index1 == -1 {
					project := &MdProject{
						Artifact:        artifact, // for searching
						Name:            shorterString,
						DependencyChain: getShortenedDepChain(dirProj.DependencyChain),
					}
					projectArray = append(projectArray, *project)
				}

			}

			comment, reviewDate := getAcceptanceData(cve)

			mdCveStruct := &MdCveStruct{
				Cve:        cve,
				CvssScore:  vuln.CvssScore, // for sorting
				Severity:   getSeverity(vuln.CvssScore),
				Link:       vuln.Reference,
				Comment:    comment,
				ReviewDate: reviewDate,
				Projects:   projectArray,
			}

			cves = append(cves, *mdCveStruct)

		} else if index != -1 {

			cveStruct := cves[index]

			for _, dirProj := range vuln.DirectProjects {

				artifact, shortString := getShorterString(dirProj.ProjectName)
				index1 := projectListed(artifact, cveStruct.Projects)
				if index1 == -1 {

					project := &MdProject{
						Artifact:        artifact, // for searching
						Name:            shortString,
						DependencyChain: getShortenedDepChain(dirProj.DependencyChain),
					}

					cveStruct.Projects = append(cveStruct.Projects, *project)

					cves[index] = cveStruct

				}
			}
		}
	}
}

func consolidateIntoProjectStructs(yaml SecVulnYamlReport) {

	for _, vuln := range yaml.Vulnerabilities {

		cve := vuln.Cve

		for _, dirProj := range vuln.DirectProjects {

			// Returns -1 if not found or the index if found, as index is needed to get the existing struct later
			artifact, shortString := getShorterString(dirProj.ProjectName)
			index := projectListedAtTopLevel(artifact, projects)

			if index == -1 {

				var cveArray []MdCve
				newCveStruct := &MdCve{
					Cve:             cve,
					CvssScore:       vuln.CvssScore, // for sorting
					Severity:        getSeverity(vuln.CvssScore),
					DependencyChain: getShortenedDepChain(dirProj.DependencyChain),
				}
				cveArray = append(cveArray, *newCveStruct)

				var dependents []string
				for _, proj := range dirProj.TransientProjects {
					dependents = append(dependents, proj.ProjectName)
				}

				newProjectStruct := &MdProjectStruct{
					Artifact:   artifact, // for searching
					Name:       shortString,
					Dependents: dependents,
					Cves:       cveArray,
				}

				projects = append(projects, *newProjectStruct)

			} else if index != -1 {

				// This may be a different scope or version of an already processed artifact, so need to merge the structs

				projectStruct := projects[index]

				for _, dep := range dirProj.TransientProjects {
					if depListed(dep.ProjectName, projectStruct.Dependents) == false {
						projectStruct.Dependents = append(projectStruct.Dependents, dep.ProjectName)
					}
				}

				index1 := cveListed(cve, projectStruct.Cves)
				if index1 == -1 {

					newCve := &MdCve{
						Cve:             cve,
						CvssScore:       vuln.CvssScore, // for sorting
						Severity:        getSeverity(vuln.CvssScore),
						DependencyChain: getShortenedDepChain(dirProj.DependencyChain),
					}

					projectStruct.Cves = append(projectStruct.Cves, *newCve)

				}

				projects[index] = projectStruct

			}
		}
	}
}

func sortCveStructs() {

	// Projects in alphabetical order within each CVE
	for _, cve := range cves {
		sort.Slice(cve.Projects, func(i, j int) bool {
			return cve.Projects[i].Name < cve.Projects[j].Name
		})
	}

	// Highest to lowest Cvss Score
	sort.Slice(cves, func(i, j int) bool {
		return cves[i].CvssScore > cves[j].CvssScore
	})

	// Sort CVEs alphabetically within the same score group
	var sameScore []MdCveStruct
	sameScore = append(sameScore, cves[0])
	for i := 1; i < len(cves); i++ {
		if cves[i].CvssScore == cves[i-1].CvssScore {
			sameScore = append(sameScore, cves[i])
		} else {
			cveScoreGroups = append(cveScoreGroups, sameScore)
			sameScore = nil
			sameScore = append(sameScore, cves[i])
		}
		if i == len(cves)-1 {
			cveScoreGroups = append(cveScoreGroups, sameScore)
		}
		continue
	}

	for _, scoreGroup := range cveScoreGroups {
		sort.Slice(scoreGroup, func(i, j int) bool {
			return scoreGroup[i].Cve < scoreGroup[j].Cve
		})
	}

}

func sortProjectStructs() {

	// Projects in alphabetical order
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	// CVEs within each project by severity order
	for _, proj := range projects {
		sort.Slice(proj.Cves, func(i, j int) bool {
			return proj.Cves[i].CvssScore > proj.Cves[j].CvssScore
		})
	}

}

func depListed(targetString string, array []string) bool {
	for _, el := range array {
		if targetString == el {
			return true
		}
	}
	return false
}

func formSummarySection() {
	for _, cve := range cves {
		cveSum := &CveSummary{
			Cve:      cve.Cve,
			Link:     cve.Link,
			Severity: getSeverity(cve.CvssScore),
			Amount:   len(cve.Projects),
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

	cveTemplate := "\n### {{.Cve}}\n\nSeverity: **{{.Severity}}**\n\n{{if .ReviewDate}}Acceptance: To be reviewed {{.ReviewDate}}\n\n{{end}}{{if .Comment}}Acceptance: {{.Comment}}\n\n{{end}}[Link]({{.Link}})\n\nGalasa Projects/Images directly affected:\n\n"
	cveTmpl, err := template.New("cveTemplate").Parse(cveTemplate)
	if err != nil {
		fmt.Printf("Unable to create the template for the CVE main section of the Markdown, %v\n", err)
		panic(err)
	}

	cveProjTemplate := "- {{.Name}}\n{{ range .DependencyChain}}  - {{.}}\n{{end}}\n"
	cveProjTmpl, err := template.New("cveProjTemplate").Parse(cveProjTemplate)
	if err != nil {
		fmt.Printf("Unable to create the template for the CVE projects section of the Markdown, %v\n", err)
		panic(err)
	}

	for _, cveScoreGroup := range cveScoreGroups {
		for _, cve := range cveScoreGroup {
			err = cveTmpl.Execute(markdownFile, cve)
			for _, proj := range cve.Projects {
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
		_, shortString := getShorterString(submatch)
		shorterDepChain = append(shorterDepChain, shortString)
	}
	return shorterDepChain
}

func getShorterString(fullString string) (string, string) {
	regex := "[a-zA-Z0-9._-]+"
	re := regexp.MustCompile(regex)
	submatches := re.FindAllString(fullString, -1)

	group := submatches[0]
	artifact := submatches[1]
	version := submatches[3]
	result := fmt.Sprintf("%s:%s:%s", group, artifact, version)

	return artifact, result
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

func projectListed(targetProject string, array []MdProject) int {
	for index, project := range array {
		if project.Artifact == targetProject {
			return index
		}
	}
	return -1
}

func projectListedAtTopLevel(targetProject string, array []MdProjectStruct) int {
	for index, project := range array {
		if project.Artifact == targetProject {
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
