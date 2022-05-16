/*
 * Copyright contributors to the Galasa project
 */

package cmd

import (
	"fmt"
	"io/ioutil"
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

	firstMap  = make(map[string]map[string]interface{})
	secondMap = make(map[string]map[string]interface{})
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
		// #### Galasa projects
		// ##### Dependency chain
		// consolidateSecVulnYamlReports(yamlReport, "cve")

		// Second structure
		// ### Galasa project
		// #### CVEs
		// ##### Dependency chain
		// consolidateSecVulnYamlReports(yamlReport, "projects")

		consolidateSecVulnYamlReports(yamlReport)

	}

	cveStructs := sortCvesSeverity()

	writeMarkdownUsingTemplates(cveStructs)
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

	body, err := ioutil.ReadAll(response.Body)
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

			writeToMap(cve, projectName, depChain, cvssScore)

		}
	}

}

func writeToMap(cve, projectName, depChain string, cvssScore float64) {

	if firstMap[cve] != nil {
		if firstMap[cve]["projects"].(map[string][]string)[projectName] != nil {
			firstMap[cve]["projects"].(map[string][]string)[projectName] = append(firstMap[cve]["projects"].(map[string][]string)[projectName], depChain)
		} else {
			var depChainArray []string
			depChainArray = append(depChainArray, depChain)
			firstMap[cve]["projects"].(map[string][]string)[projectName] = depChainArray
		}
	} else {
		firstMap[cve] = make(map[string]interface{})
		firstMap[cve]["cvssScore"] = cvssScore

		firstMap[cve]["projects"] = make(map[string][]string)
		var depChainArray []string
		depChainArray = append(depChainArray, depChain)
		firstMap[cve]["projects"].(map[string][]string)[projectName] = depChainArray
	}

	// if secondMap[projectName] != nil {
	// 	if secondMap[projectName]["cves"].(map[string][]string)[cve] != nil {
	// 		secondMap[projectName]["cves"].(map[string][]string)[cve] = append(secondMap[projectName]["cves"].(map[string][]string)[cve], depChain)
	// 	} else {
	// 		//
	// 	}
	// } else {
	// 	// cvssScoreMap := make(map[string]interface{})
	// 	// cvssScoreMap["cvssScore"] = cvssScoreMap
	// 	secondMap[projectName]["cves"].(map[string]float64)[cve] = cvssScore

	// 	var depChainArray []string
	// 	depChainArray = append(depChainArray, depChain)

	// }

}

func sortCvesSeverity() []ReportStruct {

	var cvssScores []float64
	// Get list of CvssScores from highest to lowest
	for _, innerMap := range firstMap {
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

		for cve, innerMap := range firstMap {
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

func writeMarkdownUsingTemplates(cveStructs []ReportStruct) {

	// Create file to export Markdown page to
	markdownFile, err := os.Create(fmt.Sprintf("%s/%s", secvulnReportOutput, "galasa-secvuln-report2.md"))
	if err != nil {
		fmt.Printf("Unable to create a file for the Markdown report, %v\n", err)
		panic(err)
	}
	defer markdownFile.Close()

	// Write title to MD page
	_, err = markdownFile.WriteString("# Galasa Security Vulnerability report\n\n## Section 1: CVEs and which Galasa projects they are found in\n\n")
	if err != nil {
		fmt.Printf("Unable to write to the Markdown file, %v\n", err)
		panic(err)
	}

	// Create template for the CVE section
	// TO DO - how to iterate through dep chains?
	// TO DO - don't print Comment or Review Date if there is none
	cveTemplate := "### CVE: {{.Cve}}\n### Severity: {{.Severity}}\n#### Galasa project: {{.GalasaProject}}\n#### Dependency chain(s):\n* {{.DependencyChains}}\n#### Comment: {{.Comment}}\n#### Review Date: {{.ReviewDate}}\n\n"
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

	fmt.Printf("Markdown page exported to %s\n", secvulnReportOutput)
}

func writeMarkdownUsingTemplate() {

	// _, err = markdownFile.WriteString("## Section 2: Galasa projects and CVEs they contain\n\n")
	// if err != nil {
	// 	fmt.Printf("Unable to write to the Markdown file, %v\n", err)
	// 	panic(err)
	// }

	// Create template for the Galasa projects section
	// galasaTemplate := "### Galasa project: {{.GalasaProject}}\n#### CVE: {{.Cve}}\n#### Dependency chain(s):\n* {{.DependencyChains}}\n#### Comment: {{.Comment}}\n#### Review Date: {{.ReviewDate}}\n\n"
	// galasaTmpl, err := template.New("galasaTemplate").Parse(galasaTemplate)
	// if err != nil {
	// 	fmt.Printf("Unable to create the template for the Galasa project section of the Markdown, %v\n", err)
	// 	panic(err)
	// }

	// Galasa project section
	// for galasaProjectName, innerMap := range secondMap {
	// 	for cveName, depChainsArray := range innerMap {
	// 		comment, reviewDate := getAcceptanceData(cveName)
	// 		galasaProjectStruct := ReportStruct{cveName, galasaProjectName, depChainsArray, comment, reviewDate}
	// 		galasaProjectStructs = append(galasaProjectStructs, galasaProjectStruct)
	// 	}
	// }

	// for _, galasaProjectStruct := range markdownReport.GalasaProjectStructs {
	// 	err = galasaTmpl.Execute(markdownFile, galasaProjectStruct)
	// 	if err != nil {
	// 		fmt.Printf("Unable to apply the template to the Galasa project structs, %v\n", err)
	// 		panic(err)
	// 	}
	// }

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
