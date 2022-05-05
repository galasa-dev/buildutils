//
// Copyright contributors to the Galasa project
//

package cmd

import (
	"fmt"
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

	// Unmarshal all security vulnerability reports to be translated to markdown page
	var yamlReports []YamlReport
	for _, directory := range *secvulnReportExtracts {
		yamlReports = append(yamlReports, unmarshalSecVulnYamlReports(directory))
	}

	// Unmarshal the acceptance report to merge in manager's comments and review dates with the markdown page
	acceptanceReport = unmarshalAcceptanceYamlReport()

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

	// Write two pieces of Markdown based on the two structures
	firstMarkdown := writeFirstMarkdownSection()

	secondMarkdown := writeSecondMarkdownSection()

	// Concatenate both pieces to form the full Markdown page
	fullMarkdownPage := "# Galasa Security Vulnerability report\n\n"
	fullMarkdownPage += firstMarkdown
	fullMarkdownPage += secondMarkdown

	// Write the Markdown to a file and export
	exportMarkdownPage(fullMarkdownPage)
}

func unmarshalSecVulnYamlReports(directory string) YamlReport {
	var yamlReport YamlReport

	yamlFile, err := os.ReadFile(fmt.Sprintf("%s/%s", directory, "galasa-secvuln-report.yaml"))
	if err != nil {
		fmt.Printf("Unable to read the security vulnerability report in directory %s, %v\n", directory, err)
	}

	err = yaml.Unmarshal(yamlFile, &yamlReport)
	if err != nil {
		fmt.Printf("Unable to unmarshal the security vulnerability report in directory %s, %v\n", directory, err)
	}

	return yamlReport
}

func unmarshalAcceptanceYamlReport() AcceptanceYamlReport {
	var accYamlReport AcceptanceYamlReport

	yamlFile, err := os.ReadFile(fmt.Sprintf("%s/%s", secvulnReportAcceptance, "acceptance-report.yaml"))
	if err != nil {
		fmt.Printf("Unable to read the acceptance report in directory %s, %v\n", secvulnReportAcceptance, err)
	}

	err = yaml.Unmarshal(yamlFile, &accYamlReport)
	if err != nil {
		fmt.Printf("Unable to unmarshal the acceptance report in directory %s, %v\n", secvulnReportAcceptance, err)
	}

	return accYamlReport
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

func writeFirstMarkdownSection() string {

	markdown := "## Section 1: CVEs and which Galasa projects they are found in\n\n"

	for cveName, projectMap := range firstMap {
		markdown += fmt.Sprintf("%s %s\n\n", "###", cveName)
		comment, reviewDate := getAcceptanceData(cveName)
		if comment != "" {
			markdown += fmt.Sprintf("%s Manager's comment: %s\n\n", "####", comment)
		}
		if reviewDate != "" {
			markdown += fmt.Sprintf("%s Review date: %s\n\n", "####", reviewDate)
		}
		markdown += "#### CVE found in the following Galasa projects:\n\n"
		for galasaProject, depChainsArray := range projectMap {
			markdown += fmt.Sprintf("%s %s\n\n", "####", galasaProject)
			if len(depChainsArray) > 1 {
				markdown += fmt.Sprintf("%s\n\n", "##### Dependency chains:")
			} else {
				markdown += fmt.Sprintf("%s\n\n", "##### Dependency chain:")
			}
			for _, depChain := range depChainsArray {
				markdown += fmt.Sprintf("%s %s\n\n", "#####", depChain)
			}
			markdown += fmt.Sprintf("\n")
		}
		markdown += fmt.Sprintf("\n\n")
	}

	return markdown
}

func writeSecondMarkdownSection() string {

	markdown := "## Section 2: Galasa projects and CVEs they contain\n\n"

	for galasaProject, cveMap := range secondMap {
		markdown += fmt.Sprintf("%s %s %s:\n\n", "###", galasaProject, "contains the following CVEs")
		for cveName, depChainsArray := range cveMap {
			markdown += fmt.Sprintf("%s %s\n\n", "####", cveName)
			comment, reviewDate := getAcceptanceData(cveName)
			if comment != "" {
				markdown += fmt.Sprintf("%s Manager's comment: %s\n\n", "####", comment)
			}
			if reviewDate != "" {
				markdown += fmt.Sprintf("%s Review date: %s\n\n", "####", reviewDate)
			}
			if len(depChainsArray) > 1 {
				markdown += fmt.Sprintf("%s\n\n", "##### Dependency chains:")
			} else {
				markdown += fmt.Sprintf("%s\n\n", "##### Dependency chain:")
			}
			for _, depChain := range depChainsArray {
				markdown += fmt.Sprintf("%s %s\n\n", "#####", depChain)
			}
			markdown += fmt.Sprintf("\n")
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
