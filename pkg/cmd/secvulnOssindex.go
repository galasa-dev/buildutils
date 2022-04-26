/*
 * Copyright contributors to the Galasa project
 */

package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

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

	// Get all sub-directories of the provided parent directory
	getDirectories()

	// Create the yaml report of all vulnerabilities found
	createYamlReport()
	fmt.Printf("Exported Yaml report of all vulnerabilities to %s\n", secvulnOssindexOutput)

}

func getDirectories() {

	files, err := ioutil.ReadDir(secvulnOssindexParentDir)
	if err != nil {
		fmt.Printf("Unable to read the sub-directories of the provided parent directory %v\n", err)
	}

	for _, f := range files {

		if f.IsDir() && strings.HasPrefix(f.Name(), ".") == false {

			auditReport, err := os.ReadFile(fmt.Sprintf("%s/%s/%s/%s", secvulnOssindexParentDir, f.Name(), "target", "audit-report.json"))
			if err != nil {
				fmt.Printf("Unable to find audit report in %s directory, %v\n", f.Name(), err)
			}

			// Scan the OSS Index audit report for the directory
			if auditReport != nil {
				scanAuditReport(auditReport, f.Name())
			}

		}
	}

}

func scanAuditReport(file []byte, devGalasaArtifact string) {

	// Audit report is unstructered JSON so cannot use structs
	var auditReport map[string]interface{}
	err := json.Unmarshal([]byte(file), &auditReport)
	if err != nil {
		fmt.Printf("Unable to unmarshal the audit report for module %s %v\n", devGalasaArtifact, err)
	}

	auditReportArtifacts := auditReport["reports"].(map[string]interface{})

	for auditReportArtifact, artifactDetails := range auditReportArtifacts {

		vulnerabilities := artifactDetails.(map[string]interface{})["vulnerabilities"]

		// If this artifact has vulnerabilities, add these to the yaml report
		if vulnerabilities != nil {

			for _, vulnerability := range vulnerabilities.([]interface{}) {

				// Get the CVE of the vulnerability from the audit report
				cve := vulnerability.(map[string]interface{})["cve"]

				// Get dependency chain and dependency type from the OSS Index dependency tree report
				dependencyTree, dependencyType := getDependencyTree(auditReportArtifact, devGalasaArtifact)

				project := &Project{}
				if dependencyType == "direct" {
					project.Project = devGalasaArtifact
					project.DependencyType = dependencyType
					project.DependencyChain = "n/a"
					// project.DependencyChain = dependencyTree
				} else {
					project.Project = devGalasaArtifact
					project.DependencyType = dependencyType
					project.DependencyChain = dependencyTree
				}

				if cves[cve.(string)] != nil {
					// If this CVE has an entry in the map for the yaml report, add this project to this CVE's map
					cves[cve.(string)] = append(cves[cve.(string)], *project)
				} else {
					// If this CVE does not have an entry in the map for the yaml report, then make one
					var projects []Project
					projects = append(projects, *project)
					cves[cve.(string)] = projects
				}
			}
		}
	}
}

func getDependencyTree(vulnerability, devGalasaArtifact string) (string, string) {

	devGalasaArtifactString := getFullString(devGalasaArtifact)

	digraphFile, err := os.ReadFile(fmt.Sprintf("%s/%s/%s", secvulnOssindexParentDir, devGalasaArtifact, "deps.txt"))
	if err != nil {
		fmt.Printf("Unable to get the dependency tree digraph for %s %v", devGalasaArtifact, err)
	}

	digraph := getReportExtract(string(digraphFile))

	// Split the dependency tree report into individual lines
	lines := strings.Split(digraph, "\n")

	// Add all lines to a two dimensional array
	var array [][]string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		twoParts := strings.Split(line, "->")
		array = append(array, twoParts)
	}

	// Start forming the dependency tree string for the yaml report
	targetString := vulnerability

	// Add each artifact in the dependency tree from the vulnerability to the galasa artifact to array
	var dependencyTree []string
	dependencyTree = append(dependencyTree, vulnerability)

	maxLoops := 100
	count := 0
	for targetString != devGalasaArtifactString {

		if count == maxLoops {
			fmt.Printf("Too many attempts to parse dependency tree for %s\n", vulnerability)
			panic(err)
		}

		for _, element := range array {

			// "dev.galasa:dev.galasa.artifact.manager:jar:0.21.0" -> "dev.galasa:dev.galasa:jar:0.21.0:compile"
			// If the artifact on the right is the target string, see what artifact this comes from and repeat

			if element[1] == targetString {

				dependencyTree = append(dependencyTree, element[0])

				targetString = element[0]

				break
			}

		}

		count++
	}

	// Form the dependency tree string by reversing the array
	dependencyTreeString := devGalasaArtifactString
	for i := len(dependencyTree) - 2; i > -1; i-- {
		dependencyTreeString += ", " + dependencyTree[i]
	}

	// Determine dependency type (direct or transient) based on how many artifacts in the tree
	var dependencyType string
	if len(dependencyTree) > 2 {
		dependencyType = "transient"
	} else {
		dependencyType = "direct"
	}

	return dependencyTreeString, dependencyType

}

func createYamlReport() {

	var allVulnerabilities []Vulnerability

	// Iterate through all of the CVEs
	for key, value := range cves {

		value = removeDuplicateValues(value)

		vulnerability := &Vulnerability{
			Cve:      key,
			Projects: value,
		}

		allVulnerabilities = append(allVulnerabilities, *vulnerability)
	}

	yamlReport := &YamlReport{
		Title:           "Galasa security vulnerability report",
		Description:     "All security vulnerabilities found in dev.galasa artifacts",
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

func getReportExtract(str string) string {

	str = str[strings.Index(str, "{")+1 : strings.Index(str, "}")]
	str = strings.Replace(str, "[INFO]", "", -1)
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "\"", "", -1)
	str = strings.Replace(str, ";", "", -1)
	str = strings.TrimSpace(str)

	return str
}

func getFullString(module string) string {

	pomFile, err := os.ReadFile(fmt.Sprintf("%s/%s/%s", secvulnOssindexParentDir, module, "pom.xml"))
	if err != nil {
		fmt.Printf("Error\n")
	}

	var pom Pom
	err = xml.Unmarshal(pomFile, &pom)
	if err != nil {
		fmt.Printf("Error\n")
	}

	group := pom.GroupId
	artifact := pom.ArtifactId
	packaging := pom.Packaging
	version := pom.Version

	return fmt.Sprintf("%s:%s:%s:%s", group, artifact, packaging, version)
}
