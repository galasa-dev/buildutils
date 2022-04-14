//
// Copyright contributors to the Galasa project
//

package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	secvulnOssindexCmd = &cobra.Command{
		Use:   "ossindex",
		Short: "Extract OSS Index reports",
		Long:  "Extract OSS Index reports",
		Run:   secvulnOssindexExecute,
	}

	secvulnOssindexParentDir string
	secvulnOssindexOutput    string

	modules []string

	allVulnerabilities []Vulnerability

	cves = make(map[string][]Project)
)

func init() {
	secvulnOssindexCmd.PersistentFlags().StringVar(&secvulnOssindexParentDir, "parent", "", "Parent project directory")
	secvulnOssindexCmd.PersistentFlags().StringVar(&secvulnOssindexOutput, "output", "", "Output yaml extract")

	secvulnOssindexCmd.MarkPersistentFlagRequired("parent")
	secvulnOssindexCmd.MarkPersistentFlagRequired("output")

	secvulnCmd.AddCommand(secvulnOssindexCmd)
}

func secvulnOssindexExecute(cmd *cobra.Command, args []string) {
	fmt.Printf("Galasa Build - Security Vulnerability OSS Index - version %v\n", rootCmd.Version)

	// Get all of the modules that were processed for security scanning
	getModules()

	for _, module := range modules {

		// Scan the OSS Index audit report for each module
		scanAuditReport(module)

	}

	// Create the yaml report of all vulnerabilities found
	createYamlReport()
	fmt.Printf("Exported Yaml report of all vulnerabilities to %s\n", secvulnOssindexOutput)

}

func getModules() {

	file, err := os.ReadFile(fmt.Sprintf("%s/%s", secvulnOssindexParentDir, "pom.xml"))
	if err != nil {
		fmt.Printf("Could not read pom %v", err)
	}

	var pom Pom
	err = xml.Unmarshal(file, &pom)
	if err != nil {
		fmt.Printf("Could not unmarshal pom %v", err)
	}

	modules = pom.Modules.Module

}

func scanAuditReport(module string) {

	file, err := os.ReadFile(fmt.Sprintf("%s/%s/%s/%s", secvulnOssindexParentDir, module, "target", "audit-report.json"))
	if err != nil {
		fmt.Printf("Unable to get the audit report for module %s %v", module, err)
	}

	// Audit report is unstructered JSON so cannot use structs
	var auditReport map[string]interface{}
	err = json.Unmarshal([]byte(file), &auditReport)
	if err != nil {
		fmt.Printf("Unable to unmarshal the audit report for module %s %v", module, err)
	}

	reports := auditReport["reports"].(map[string]interface{})

	for _, value := range reports {

		vulnerabilities := value.(map[string]interface{})["vulnerabilities"]

		// If this artifact has vulnerabilities, add this to the yaml report
		if vulnerabilities != nil {

			for _, vulnerability := range vulnerabilities.([]interface{}) {

				// Get the CVE from the audit report
				cve := vulnerability.(map[string]interface{})["cve"]

				project := &Project{
					Project: module,
				}

				if cves[cve.(string)] != nil {
					// If this CVE has an entry in the map, add this project to this CVE's map

					cves[cve.(string)] = append(cves[cve.(string)], *project)
				} else {
					// If this CVE does not have an entry in the map then make one

					var projects []Project
					projects = append(projects, *project)
					cves[cve.(string)] = projects
				}
			}
		}
	}
}

func createYamlReport() {

	for key, value := range cves {

		value = removeDuplicateValues(value)

		vulnerability := &Vulnerability{
			Cve:      key,
			Projects: value,
		}
		allVulnerabilities = append(allVulnerabilities, *vulnerability)
	}

	yamlReport := &YamlReport{
		Vulnerabilities: allVulnerabilities,
	}

	filename := fmt.Sprintf("%s/%s", secvulnOssindexOutput, "report.yaml")
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Unable to create report.yaml %v\n", err)
		panic(err)
	}

	xmlWriter := io.Writer(file)

	enc := yaml.NewEncoder(xmlWriter)
	err = enc.Encode(yamlReport)
	if err != nil {
		fmt.Printf("Unable to encode the pom.xml for security scanning project %v\n", err)
		panic(err)
	}
}

func removeDuplicateValues(allProjects []Project) []Project {
	var projects []Project
	for _, project := range allProjects {
		if checkIfInArray(project, projects) == false {
			projects = append(projects, project)
		}
	}
	return projects
}

func checkIfInArray(a Project, projects []Project) bool {
	for _, b := range projects {
		if b == a {
			return true
		}
	}
	return false
}
