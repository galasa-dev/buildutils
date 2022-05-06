/*
 * Copyright contributors to the Galasa project
 */

package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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

	firstMap  = make(map[string]map[string][]string)
	secondMap = make(map[string]map[string][]string)

	acceptanceReport AcceptanceYamlReport
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
		consolidateSecVulnYamlReports(yamlReport, "cve")

		// Second structure
		// ### Galasa project
		// #### CVEs
		// ##### Dependency chain
		consolidateSecVulnYamlReports(yamlReport, "project")

	}

	// Write the Markdown page
	markdownPage := "# Galasa Security Vulnerability report\n\n"
	markdownPage += writeMarkdown()

	// Write the Markdown to a file and export
	exportMarkdownPage(markdownPage)
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

func consolidateSecVulnYamlReports(yamlReport YamlReport, topLevelObject string) {
	for _, vulnerability := range yamlReport.Vulnerabilities {
		for _, project := range vulnerability.Projects {

			cve := vulnerability.Cve
			projectName := project.Project
			depChain := project.DependencyChain

			if topLevelObject == "cve" {
				writeToMap(firstMap, cve, projectName, depChain)
			} else if topLevelObject == "project" {
				writeToMap(secondMap, projectName, cve, depChain)
			}
		}
	}
}

func writeToMap(whicheverMap map[string]map[string][]string, topLevelObject, secondLevelObject, depChain string) {
	if whicheverMap[topLevelObject] != nil {
		if whicheverMap[topLevelObject][secondLevelObject] != nil {
			whicheverMap[topLevelObject][secondLevelObject] = append(whicheverMap[topLevelObject][secondLevelObject], depChain)
		} else {
			var depChainArray []string
			depChainArray = append(depChainArray, depChain)
			whicheverMap[topLevelObject][secondLevelObject] = depChainArray
		}
	} else {
		var depChainArray []string
		depChainArray = append(depChainArray, depChain)

		projectMap := make(map[string][]string)
		projectMap[secondLevelObject] = depChainArray

		whicheverMap[topLevelObject] = projectMap
	}
}

func getAcceptanceData(cve string) (string, string) {
	for _, value := range acceptanceReport.Cves {
		if value.Cve == cve {
			return value.Comment, value.ReviewDate
		}
	}
	return "", ""
}

func writeMarkdown() string {

	markdown := "## Section 1: CVEs and which Galasa projects they are found in\n\n"

	for cveName, projectMap := range firstMap {
		for galasaProject, depChainsArray := range projectMap {
			markdown += fmt.Sprintf("### CVE: %s ", cveName)
			markdown += fmt.Sprintf("Galasa project: %s ", galasaProject)
			markdown += "Dependency chain(s): "
			for _, depChain := range depChainsArray {
				markdown += fmt.Sprintf("%s ", depChain)
			}
			comment, reviewDate := getAcceptanceData(cveName)
			if comment != "" {
				markdown += fmt.Sprintf("\n")
				markdown += fmt.Sprintf("#### Manager's comment: %s", comment)
			}
			if reviewDate != "" {
				markdown += fmt.Sprintf("\n")
				markdown += fmt.Sprintf("#### Review date: %s", reviewDate)
			}
			markdown += fmt.Sprintf("\n\n")
		}
		markdown += fmt.Sprintf("\n\n")
	}

	markdown += "## Section 2: Galasa projects and CVEs they contain\n\n"

	for galasaProject, cveMap := range secondMap {
		for cveName, depChainsArray := range cveMap {
			markdown += fmt.Sprintf("### Galasa project: %s ", galasaProject)
			markdown += fmt.Sprintf("CVE: %s ", cveName)
			markdown += "Dependency chain(s): "
			for _, depChain := range depChainsArray {
				markdown += fmt.Sprintf("%s ", depChain)
			}
			comment, reviewDate := getAcceptanceData(cveName)
			if comment != "" {
				markdown += fmt.Sprintf("\n")
				markdown += fmt.Sprintf("#### Manager's comment: %s", comment)
			}
			if reviewDate != "" {
				markdown += fmt.Sprintf("\n")
				markdown += fmt.Sprintf("#### Review date: %s", reviewDate)
			}
			markdown += fmt.Sprintf("\n\n")
		}
		markdown += fmt.Sprintf("\n\n")
	}

	return markdown
}

func exportMarkdownPage(markdownPage string) {
	markdownFile, err := os.Create(fmt.Sprintf("%s/%s", secvulnReportOutput, "galasa-secvuln-report.md"))
	if err != nil {
		fmt.Printf("Unable to create a file for the markdown, %v\n", err)
	}

	defer markdownFile.Close()

	_, err = markdownFile.WriteString(markdownPage)
	if err != nil {
		fmt.Printf("Unable to write the markdown to the markdown file, %v\n", err)
	}

	fmt.Printf("Markdown page exported to %s\n", secvulnReportOutput)
}
