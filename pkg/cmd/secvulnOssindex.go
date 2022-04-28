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
	"regexp"
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

	// Find all target directories in the structure of the provided parent directory
	findTargetDirectories(secvulnOssindexParentDir)

	// Create the yaml report of all vulnerabilities found
	createYamlReport()
	fmt.Printf("Exported Yaml report of all vulnerabilities to %s\n", secvulnOssindexOutput)

}

func findTargetDirectories(directory string) {

	subDirs, err := ioutil.ReadDir(directory)
	if err != nil {
		fmt.Printf("Unable to read the sub-directories of directory %s, %v\n", directory, err)
	}

	for _, f := range subDirs {

		if f.IsDir() && strings.HasPrefix(f.Name(), ".") == false {

			if f.Name() == "target" {

				auditReport, _ := os.ReadFile(fmt.Sprintf("%s/%s/%s", directory, f.Name(), "audit-report.json"))

				// If this target directory contains an OSS Index audit report, scan through it
				if auditReport != nil {
					// Pass through the directory to use later to find the pom and deps file
					scanAuditReportForVulnerabilities(auditReport, directory)
				}

			} else {
				// Repeat this function within this directory if a target directory hasn't been found
				findTargetDirectories(fmt.Sprintf("%s/%s", directory, f.Name()))
			}

		}
	}
}

func scanAuditReportForVulnerabilities(file []byte, directory string) {

	// Audit report is unstructered JSON so cannot use structs
	var auditReport map[string]interface{}
	err := json.Unmarshal([]byte(file), &auditReport)
	if err != nil {
		fmt.Printf("Unable to unmarshal the audit report in %s, %v\n", directory, err)
	}

	vulnerableArtifacts := auditReport["vulnerable"]

	// If this audit report has a vulnerable section, iterate through the vulnerable artifacts
	if vulnerableArtifacts != nil {

		for vulnerableArtifact, artifactDetails := range vulnerableArtifacts.(map[string]interface{}) {

			// Each vulnerable artifact might have more than one CVE so iterate through them
			vulnerabilities := artifactDetails.(map[string]interface{})["vulnerabilities"]

			for _, vulnerability := range vulnerabilities.([]interface{}) {

				/*
					Get the information needed for the Yaml report
					- the CVE
					- the name of the Galasa artifact this vulnerability came from
					- the dependency chain
					- the dependency type (direct or transient)
				*/

				cve := vulnerability.(map[string]interface{})["cve"].(string)

				// Get the Galasa artifact string (group:artifact:packaging:version) from the pom of this directory
				// to use to parse the dependency chain
				galasaArtifact, galasaArtifactString := getGalasaArtifactString(directory)

				// Get the digraph from the output of mvn dependency:tree and work out dependency chain
				digraph := getDigraph(directory)
				dependencyChain, dependencyType := getDependencyChain(vulnerableArtifact, digraph, galasaArtifactString)

				addToMapForYamlReport(cve, galasaArtifact, dependencyType, dependencyChain)
			}
		}
	}
}

func addToMapForYamlReport(cve, galasaArtifact, dependencyType, dependencyChain string) {
	// Form a Project struct
	project := &Project{}
	if dependencyType == "direct" {
		project.Project = galasaArtifact
		project.DependencyType = dependencyType
		project.DependencyChain = "n/a"
	} else {
		project.Project = galasaArtifact
		project.DependencyType = dependencyType
		project.DependencyChain = dependencyChain
	}

	// Add this Project to the CVE map to be put into the yaml report
	if cves[cve] != nil {
		// If this CVE has an entry in the map already
		cves[cve] = append(cves[cve], *project)
	} else {
		// If this CVE does not have an entry in the map then make one
		var projects []Project
		projects = append(projects, *project)
		cves[cve] = projects
	}
}

func getDependencyChain(vulnerability, digraph, galasaArtifactString string) (string, string) {

	// Regex for all lines in the digraph with two artifacts separated by ->
	// First capture group is the artifact before the arrow, second is the artifact after the arrow
	regex := "([a-zA-Z0-9.:-]+)\"\\s->\\s\"([a-zA-Z0-9.:-]+)"
	re := regexp.MustCompile(regex)

	submatches := re.FindAllStringSubmatch(digraph, -1)

	// Start forming the dependency chain
	var dependencyChain []string
	dependencyChain = append(dependencyChain, vulnerability)

	// Start looking for the vulnerability first then work backwards to the Galasa artifact
	targetString := vulnerability

	maxLoops := 100
	count := 0
	for targetString != galasaArtifactString {

		if count == maxLoops {
			fmt.Printf("Too many attempts to parse dependency chain for %s\n", vulnerability)
			panic(nil)
		}

		for _, submatch := range submatches {

			// If the second capture group is the current target string, change target string to the first capture group and repeat
			if submatch[2] == targetString {
				dependencyChain = append(dependencyChain, submatch[1])
				targetString = submatch[1]
				break
			}

		}
		count++
	}

	// Form the dependency chain string for the yaml report by reversing the array
	dependencyChainString := galasaArtifactString
	for i := len(dependencyChain) - 2; i > -1; i-- {
		dependencyChainString += ", " + dependencyChain[i]
	}

	// Determine dependency type based on how many artifacts in the chain
	var dependencyType string
	if len(dependencyChain) > 2 {
		dependencyType = "transient"
	} else {
		dependencyType = "direct"
	}

	return dependencyChainString, dependencyType

}

func createYamlReport() {

	var allVulnerabilities []Vulnerability

	// Iterate through all of the CVEs and list all Galasa projects this CVE can be found in
	for key, value := range cves {

		vulnerability := &Vulnerability{
			Cve:      key,
			Projects: value,
		}

		allVulnerabilities = append(allVulnerabilities, *vulnerability)
	}

	yamlReport := &YamlReport{
		Vulnerabilities: allVulnerabilities,
	}

	// Export the yaml report to the provided output directory
	filename := fmt.Sprintf("%s/%s", secvulnOssindexOutput, "report.yaml")
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Unable to create report.yaml, %v\n", err)
		panic(err)
	}

	xmlWriter := io.Writer(file)

	enc := yaml.NewEncoder(xmlWriter)
	err = enc.Encode(yamlReport)
	if err != nil {
		fmt.Printf("Unable to encode the pom.xml for security scanning project, %v\n", err)
		panic(err)
	}
}

func getPom(directory string) Pom {
	pomFile, err := os.ReadFile(fmt.Sprintf("%s/%s", directory, "pom.xml"))
	if err != nil {
		fmt.Printf("Unable to read the pom for directory %s, %v\n", directory, err)
	}

	var pom Pom
	err = xml.Unmarshal(pomFile, &pom)
	if err != nil {
		fmt.Printf("Unable to unmarshal the pom for directory %s, %v\n", directory, err)
	}

	return pom
}

func getGalasaArtifactString(directory string) (string, string) {
	pom := getPom(directory)

	group := pom.GroupId
	artifact := pom.ArtifactId
	packaging := pom.Packaging
	version := pom.Version

	galasaArtifactString := fmt.Sprintf("%s:%s:%s:%s", group, artifact, packaging, version)

	return artifact, galasaArtifactString
}

func getDigraph(directory string) string {
	digraphFile, err := os.ReadFile(fmt.Sprintf("%s/%s", directory, "deps.txt"))
	if err != nil {
		fmt.Printf("Unable to find the dependency chain digraph in %s, %v\n", directory, err)
	}

	digraph := string(digraphFile)

	return digraph
}
