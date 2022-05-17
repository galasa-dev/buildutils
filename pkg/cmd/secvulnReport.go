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

	cveMap     = make(map[string]map[string]interface{})
	projectMap = make(map[string]map[string]map[string]interface{})
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

	// Unmarshal all security vulnerability reports to be translated to markdown page
	var yamlReports []YamlReport
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

		// As there may be multiple Yaml reports, they must all be consolidated into maps
		// so duplicate information is not put onto the Markdown page

		// There are two different maps for the two hierarchical structures of the Markdown page:

		// First structure
		// ### CVE
		// ### Severity
		// #### Galasa project that contains the CVE
		// * Dependency chain(s)
		// #### Comment
		// #### Review date

		// Second structure
		// ### Galasa project
		// #### CVE it contains
		// #### Severity
		// * Dependency chain(s)
		// #### Comment
		// #### Review date

		consolidateSecVulnYamlReports(yamlReport)

	}

	cveStructs := sortCves()

	galasaStructs := sortCvesWithinProjects()

	writeMarkdownUsingTemplates(cveStructs, galasaStructs)
}

func unmarshalSecVulnYamlReports(directory string) (YamlReport, error) {
	var yamlReport YamlReport

	yamlFile, err := os.ReadFile(fmt.Sprintf("%s/%s", directory, "galasa-secvuln-report.yaml"))
	if err != nil {
		return yamlReport, err
	}

	err = yaml.Unmarshal(yamlFile, &yamlReport)
	if err != nil {
		return yamlReport, err
	}

	return yamlReport, err
}

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

func consolidateSecVulnYamlReports(yamlReport YamlReport) {

	for _, vulnerability := range yamlReport.Vulnerabilities {
		for _, project := range vulnerability.Projects {

			cve := vulnerability.Cve
			cvssScore := vulnerability.CvssScore
			projectName := project.Project
			depChain := project.DependencyChain

			writeToCveMap(cve, projectName, depChain, cvssScore)

			writeToProjectMap(cve, projectName, depChain, cvssScore)

		}
	}

}

func writeToCveMap(cve, projectName, depChain string, cvssScore float64) {

	if cveMap[cve] != nil {
		if cveMap[cve]["projects"].(map[string][]string)[projectName] != nil {
			cveMap[cve]["projects"].(map[string][]string)[projectName] = append(cveMap[cve]["projects"].(map[string][]string)[projectName], depChain)
		} else {
			var depChainArray []string
			depChainArray = append(depChainArray, depChain)
			cveMap[cve]["projects"].(map[string][]string)[projectName] = depChainArray
		}
	} else {
		cveMap[cve] = make(map[string]interface{})
		cveMap[cve]["cvssScore"] = cvssScore

		cveMap[cve]["projects"] = make(map[string][]string)
		var depChainArray []string
		depChainArray = append(depChainArray, depChain)
		cveMap[cve]["projects"].(map[string][]string)[projectName] = depChainArray
	}

}

func writeToProjectMap(cve, projectName, depChain string, cvssScore float64) {

	if projectMap[projectName] != nil {
		if projectMap[projectName][cve] != nil {
			projectMap[projectName][cve]["depChain"] = append(projectMap[projectName][cve]["depChain"].([]string), depChain)
		} else {
			projectMap[projectName][cve] = make(map[string]interface{})

			projectMap[projectName][cve]["cvssScore"] = cvssScore

			var depChainArray []string
			depChainArray = append(depChainArray, depChain)
			projectMap[projectName][cve]["depChain"] = depChainArray
		}
	} else {
		projectMap[projectName] = make(map[string]map[string]interface{})

		projectMap[projectName][cve] = make(map[string]interface{})

		projectMap[projectName][cve]["cvssScore"] = cvssScore

		var depChainArray []string
		depChainArray = append(depChainArray, depChain)
		projectMap[projectName][cve]["depChain"] = depChainArray
	}
}

func sortCves() []ReportStruct {

	var cvssScores []float64
	// Get list of CvssScores from highest to lowest
	for _, innerMap := range cveMap {
		cvssScores = append(cvssScores, innerMap["cvssScore"].(float64))
	}
	sort.Float64s(cvssScores)
	sort.Sort(sort.Reverse(sort.Float64Slice(cvssScores)))

	var cveStructs []ReportStruct

	// Write to Markdown page in order of CVE highest to lowest
	index := 0
	for index < len(cvssScores) {

		// Start at the highest CVSS Score
		highestCvss := cvssScores[index]

		for cve, innerMap := range cveMap {
			if innerMap["cvssScore"] == highestCvss {
				severity := getSeverity(innerMap["cvssScore"].(float64))
				if severity == "" {
					fmt.Printf("Unable to get severity level from Cvss Score")
					panic(nil)
				}
				comment, reviewDate := getAcceptanceData(cve)
				projects := innerMap["projects"]
				for project, depChainsArray := range projects.(map[string][]string) {
					cveStruct := ReportStruct{cve, severity, project, depChainsArray, comment, reviewDate}
					cveStructs = append(cveStructs, cveStruct)
				}
				// Move to next highest
				index++
			}
		}
	}

	return cveStructs
}

func sortCvesWithinProjects() []ReportStruct {

	var galasaStructs []ReportStruct

	for project, innerMap := range projectMap {

		// Same as with the CVEs but inside each Galasa project

		var cvssScores []float64
		// Get list of CvssScores from highest to lowest
		for _, innerInnerMap := range innerMap {
			cvssScores = append(cvssScores, innerInnerMap["cvssScore"].(float64))
		}
		sort.Float64s(cvssScores)
		sort.Sort(sort.Reverse(sort.Float64Slice(cvssScores)))

		index := 0
		for index < len(cvssScores) {

			// Start at the highest CVSS Score
			highestCvss := cvssScores[index]

			for cve, innerInnerMap := range innerMap {
				if innerInnerMap["cvssScore"] == highestCvss {
					severity := getSeverity(innerInnerMap["cvssScore"].(float64))
					if severity == "" {
						fmt.Printf("Unable to get severity level from Cvss Score")
						panic(nil)
					}
					comment, reviewDate := getAcceptanceData(cve)
					depChainsArray := innerInnerMap["depChain"].([]string)
					cveStruct := ReportStruct{cve, severity, project, depChainsArray, comment, reviewDate}
					galasaStructs = append(galasaStructs, cveStruct)

					// Move to next highest
					index++
				}
			}

		}

	}

	return galasaStructs
}

func writeMarkdownUsingTemplates(cveStructs, galasaStructs []ReportStruct) {

	// Create file to export Markdown page
	markdownFile, err := os.Create(fmt.Sprintf("%s/%s", secvulnReportOutput, "galasa-secvuln-report2.md"))
	if err != nil {
		fmt.Printf("Unable to create a file for the Markdown report, %v\n", err)
		panic(err)
	}
	defer markdownFile.Close()

	// Write Title and Section 1 Header to MD page
	_, err = markdownFile.WriteString("# Galasa Security Vulnerability report\n\n## Section 1: CVEs and which Galasa projects they are found in\n\n")
	if err != nil {
		fmt.Printf("Unable to write to the Markdown file, %v\n", err)
		panic(err)
	}

	// Create template for the CVE section
	cveTemplate := "### CVE: {{.Cve}}\n### Severity: {{.Severity}}\n#### Galasa project: {{.GalasaProject}}\n#### Dependency chain(s):\n{{ range .DependencyChains }}* {{.}}\n{{end}}{{if .Comment}}#### Comment: {{ .Comment }}\n{{end}}{{if .ReviewDate}}#### Review Date: {{ .ReviewDate }}{{end}}\n\n"
	cveTmpl, err := template.New("cveTemplate").Parse(cveTemplate)
	if err != nil {
		fmt.Printf("Unable to create the template for the CVE section of the Markdown, %v\n", err)
		panic(err)
	}

	for _, cveStruct := range cveStructs {
		err = cveTmpl.Execute(markdownFile, cveStruct)
		if err != nil {
			fmt.Printf("Unable to apply the template to the CVE structs, %v\n", err)
			panic(err)
		}
	}

	// Write Section 2 Header to MD page
	_, err = markdownFile.WriteString("## Section 2: Galasa projects and CVEs they contain\n\n")
	if err != nil {
		fmt.Printf("Unable to write to the Markdown file, %v\n", err)
		panic(err)
	}

	// Create template for the Galasa projects section
	galasaTemplate := "### Galasa project: {{.GalasaProject}}\n#### CVE: {{.Cve}}\n#### Severity: {{.Severity}}\n#### Dependency chain(s):\n{{ range .DependencyChains }}* {{.}}\n{{end}}{{if .Comment}}#### Comment: {{ .Comment }}\n{{end}}{{if .ReviewDate}}#### Review Date: {{ .ReviewDate }}{{end}}\n\n"
	galasaTmpl, err := template.New("galasaTemplate").Parse(galasaTemplate)
	if err != nil {
		fmt.Printf("Unable to create the template for the Galasa project section of the Markdown, %v\n", err)
		panic(err)
	}

	for _, galasaStruct := range galasaStructs {
		err = galasaTmpl.Execute(markdownFile, galasaStruct)
		if err != nil {
			fmt.Printf("Unable to apply the template to the Galasa project structs, %v\n", err)
			panic(err)
		}
	}

	fmt.Printf("Markdown page exported to %s\n", secvulnReportOutput)
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
