/*
 * Copyright contributors to the Galasa project
 */

package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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

	cveInfoMap = make(map[string]map[string]interface{})

	projectHierarchyMap = make(map[string]map[string][]string)

	depChainMap = make(map[string]string)

	processedProjectCount int
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
	findAuditReports(secvulnOssindexParentDir)

	// Create the yaml report of all vulnerabilities found
	secVulnYamlReport := createReport()

	// Export to provided location
	exportReport(*secVulnYamlReport)
}

func findAuditReports(directory string) {
	err := filepath.Walk(directory, walkFunc)
	if err != nil {
		fmt.Printf("Error walking the path %s, %v\n", directory, err)
		panic(err)
	}
}

func walkFunc(path string, info fs.FileInfo, err error) error {
	if err != nil {
		fmt.Printf("Could not access path %s, %v\n", path, err)
		return err
	}

	if info.Name() == "audit-report.json" {
		// OSS Index report found
		auditReport, _ := os.ReadFile(path)
		newPath := strings.Replace(path, "/target/audit-report.json", "", -1)
		// Pass the audit report for scanning
		// and the path two dirs up to find the pom and digraph later
		scanAuditReportForVulnerabilities(auditReport, newPath)
		return nil
	}
	return nil
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
					- the CVSS Score
					- the name of the vulnerable external artifact
					- the name of the Galasa artifact first in the dependency chain (fix point)
					- the dependency chain
					- other Galasa artifacts indirectly affected by this
				*/

				cve := vulnerability.(map[string]interface{})["cve"].(string)
				cvssScore := vulnerability.(map[string]interface{})["cvssScore"].(float64)

				reference := vulnerability.(map[string]interface{})["reference"].(string)

				// This map will be iterated through later to make the Yaml report
				if cveInfoMap[cve] == nil {
					cveInfoMap[cve] = make(map[string]interface{})
					cveInfoMap[cve]["cvssScore"] = cvssScore
					cveInfoMap[cve]["reference"] = reference
					cveInfoMap[cve]["vulnerableArtifact"] = vulnerableArtifact
				}

				// Get the Galasa artifact string (group:artifact:packaging:version) from the pom of the provided
				// directory to use to parse the dependency chain
				galasaArtifactString := getGalasaArtifactString(directory)
				// Get the digraph from the output of mvn dependency:tree to work out dependency chain
				digraph := getDigraph(directory)

				dependencyChain := getDependencyChain(cve, vulnerableArtifact, digraph, galasaArtifactString)
				// Add to map to pull from later
				depChainMap[getArtifact(galasaArtifactString)] = dependencyChain
			}
		}
	}

	processedProjectCount++
}

func getDependencyChain(cve, vulnerability, digraph, galasaArtifactString string) string {

	// Regex for all lines in the digraph with two artifacts separated by ->
	// First capture group is the artifact before the arrow, second is the artifact after the arrow
	regex := "([a-zA-Z0-9.:_-]+)\"\\s->\\s\"([a-zA-Z0-9.:_-]+)"
	re := regexp.MustCompile(regex)

	submatches := re.FindAllStringSubmatch(digraph, -1)

	// Start forming the dependency chain
	var dependencyChain []string
	dependencyChain = append(dependencyChain, vulnerability)

	// Start looking for the vulnerability first then work backwards to the Galasa artifact
	targetString := vulnerability

	maxLoops := 100
	count := 0
	// For finding the first dev.galasa artifact in the chain
	firstArtifactFound := false
	firstArtifact := ""
	for targetString != galasaArtifactString {

		if count == maxLoops {
			fmt.Printf("Too many attempts to parse dependency chain from %s to %s\n", galasaArtifactString, vulnerability)
			panic(nil)
		}

		for _, submatch := range submatches {

			// If the second capture group is the current target string, change target string to the first capture group and repeat
			// Compare the artifact names only as versions might be different
			if getArtifact(submatch[2]) == getArtifact(targetString) {

				dependencyChain = append(dependencyChain, submatch[1])
				targetString = submatch[1]

				/* Find first dev.galasa artifact that this vulnerability is found in and
				determine which dev.galasa artifacts are indirectly affected by it
				*/
				if strings.HasPrefix(submatch[1], "dev.galasa") {
					if firstArtifactFound == false {
						firstArtifact = submatch[1]
						firstArtifactFound = true
						if projectHierarchyMap[cve] != nil {
							if projectHierarchyMap[cve][firstArtifact] == nil {
								var artifactArray []string
								projectHierarchyMap[cve][firstArtifact] = artifactArray
							}
						} else {
							projectHierarchyMap[cve] = make(map[string][]string)
							var artifactArray []string
							projectHierarchyMap[cve][firstArtifact] = artifactArray
						}
					} else if firstArtifactFound == true {
						if contains(submatch[1], projectHierarchyMap[cve][firstArtifact]) == false {
							projectHierarchyMap[cve][firstArtifact] = append(projectHierarchyMap[cve][firstArtifact], submatch[1])
						}
					}
				}
				break
			}

		}
		count++
	}

	// Form the dependency chain string for the yaml report by reversing the array
	var dependencyChainString string
	for i := len(dependencyChain) - 2; i > -1; i-- {
		if i == 0 {
			dependencyChainString += dependencyChain[i]
		} else {
			dependencyChainString += dependencyChain[i] + " -> "
		}
	}

	return dependencyChainString

}

func getArtifact(fullString string) string {
	regex := "[a-zA-Z0-9._]+"
	re := regexp.MustCompile(regex)
	submatches := re.FindAllString(fullString, -1)
	return submatches[1]
}

func createReport() *SecVulnYamlReport {

	var vulns []Vulnerability

	// Unsure if necessary as need to resort by CVSS Score in the Markdown command
	// as pulling in multiple Yamls might ruin the order
	sortedCvssScores := sortCvssScores()

	index := 0
	for index < len(sortedCvssScores) {

		highestCvss := sortedCvssScores[index]

		for cve, cveInfo := range cveInfoMap {

			if cveInfo["cvssScore"].(float64) == highestCvss {

				var directProjects []DirectProject

				cvesProjects := projectHierarchyMap[cve]

				for directProject, innerProjects := range cvesProjects {

					var transientProjs []TransientProject

					for _, innerProject := range innerProjects {

						depChain := depChainMap[getArtifact(innerProject)]
						if depChain == "" {
							fmt.Printf("Unable to find dependency chain for affected project %s\n", innerProject)
							panic(nil)
						}

						transientProj := &TransientProject{
							ProjectName:     innerProject,
							DependencyChain: depChain,
						}
						transientProjs = append(transientProjs, *transientProj)

					}

					directDepChain := depChainMap[getArtifact(directProject)]
					if directDepChain == "" {
						fmt.Printf("Unable to find dependency chain for directly affected project %s\n", directProject)
						panic(nil)
					}

					directProject := &DirectProject{
						ProjectName:       directProject,
						DependencyChain:   directDepChain,
						TransientProjects: transientProjs,
					}
					directProjects = append(directProjects, *directProject)

				}

				vuln := &Vulnerability{
					Cve:                cve,
					CvssScore:          cveInfo["cvssScore"].(float64),
					Reference:          cveInfo["reference"].(string),
					VulnerableArtifact: cveInfo["vulnerableArtifact"].(string),
					DirectProjects:     directProjects,
				}

				vulns = append(vulns, *vuln)

				index++
			}

		}

	}

	yamlReport := &SecVulnYamlReport{
		Vulnerabilities: vulns,
	}

	return yamlReport

}

func exportReport(yamlReport SecVulnYamlReport) {
	// Export the yaml report to the provided output directory
	file, err := os.Create(secvulnOssindexOutput)
	if err != nil {
		fmt.Printf("Unable to create the security vulnerability report, %v\n", err)
		panic(err)
	}

	yamlWriter := io.Writer(file)

	enc := yaml.NewEncoder(yamlWriter)
	err = enc.Encode(yamlReport)
	if err != nil {
		fmt.Printf("Unable to encode the security vulnerability report, %v\n", err)
		panic(err)
	}

	fmt.Printf("%v vulnerabilities found in %v projects\nExported security vulnerability report to %s\n", len(cveInfoMap), processedProjectCount, secvulnOssindexOutput)
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

func getGalasaArtifactString(directory string) string {
	pom := getPom(directory)

	group := pom.GroupId
	artifact := pom.ArtifactId
	packaging := pom.Packaging
	version := pom.Version

	galasaArtifactString := fmt.Sprintf("%s:%s:%s:%s", group, artifact, packaging, version)

	return galasaArtifactString
}

func getDigraph(directory string) string {
	digraphFile, err := os.ReadFile(fmt.Sprintf("%s/%s/%s", directory, "target", "deps.txt"))
	if err != nil {
		fmt.Printf("Unable to find the dependency chain digraph in %s, %v\n", directory, err)
	}

	digraph := string(digraphFile)

	return digraph
}

func sortCvssScores() []float64 {
	// Sort list of CvssScores from highest to lowest
	var cvssScores []float64
	for _, innerMap := range cveInfoMap {
		cvssScores = append(cvssScores, innerMap["cvssScore"].(float64))
	}
	sort.Float64s(cvssScores)
	sort.Sort(sort.Reverse(sort.Float64Slice(cvssScores)))

	return cvssScores
}

func contains(artifact string, array []string) bool {
	for _, element := range array {
		if artifact == element {
			return true
		}
	}
	return false
}
