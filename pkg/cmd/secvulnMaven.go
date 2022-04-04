//
// Copyright contributors to the Galasa project
//

package cmd

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
	secvulnMavenCmd = &cobra.Command{
		Use:   "maven",
		Short: "Generate psuedo maven project for security vulnerability scanning",
		Long:  "Generate psuedo maven project for security vulnerability scanning",
		Run:   secvulnMavenExecute,
	}

	secvulnMavenParentDir string
	secvulnMavenPomUrls   *[]string
	secvulnMavenPomRepos  *[]string

	completedProjects []string
	toDoProjects      []string
)

func init() {
	secvulnMavenCmd.PersistentFlags().StringVar(&secvulnMavenParentDir, "parent", "", "Parent project directory")
	secvulnMavenPomUrls = secvulnMavenCmd.PersistentFlags().StringArray("pom", nil, "Component Pom URLs")
	secvulnMavenPomRepos = secvulnMavenCmd.PersistentFlags().StringArray("repo", nil, "Repos to look for Poms")

	secvulnMavenCmd.MarkPersistentFlagRequired("parent")
	secvulnMavenCmd.MarkPersistentFlagRequired("pom")
	secvulnMavenCmd.MarkPersistentFlagRequired("repo")

	secvulnCmd.AddCommand(secvulnMavenCmd)
}

func secvulnMavenExecute(cmd *cobra.Command, args []string) {
	fmt.Printf("Galasa Build - Security Vulnerability Maven - version %v\n", rootCmd.Version)

	for _, pom := range *secvulnMavenPomUrls {
		startScanningPipeline(pom)
	}

	updateParent()
	fmt.Printf("Security scanning project pom.xml updated\n")

}

func startScanningPipeline(mainPomUrl string) {

	fmt.Printf("Starting the scanning pipeline at %v\n", mainPomUrl)

	mainPom, err := readPomFromUrl(mainPomUrl)
	if err != nil {
		fmt.Printf("Unable to find the main pom at %v\n", mainPomUrl)
		panic(err)
	}

	createPseudoMavenProject(mainPom)
	fmt.Printf("Pseudo maven project created for: %v\n", mainPom.ArtifactId)

	// Repeat the process for all projects in this dependency chain if the groupId is dev.galasa
	fmt.Printf("Pseudo maven projects created for dependency chain of %v:\n", mainPom.ArtifactId)

	for _, project := range toDoProjects {

		// If this project has already been processed then continue
		if checkIfCompleted(project) {
			continue
		}

		// Get the groupId of this project from the main pom to use in the url
		var groupId string
		if groupId = getGroupId(project, mainPom); groupId == "" {
			fmt.Printf("Unable to get groupId for artifact %s\n", project)
			panic(nil)
		}
		// Get the version of this project from the main pom to use in the url
		var version string
		if version = getVersion(project, mainPom); version == "" {
			fmt.Printf("Unable to get version for artifact %s\n", project)
			panic(nil)
		}

		// Get the current pom for this project to use to create the stripped down pom for the pseudo maven project
		var currentPom Pom
		if project == "com.jcraft.jsch" {
			// com.jcraft.jsch is an exception to the normal format
			currentPom, err = readPomFromRepos("jsch", "com.jcraft", version)
		} else {
			currentPom, err = readPomFromRepos(project, groupId, version)
		}
		if err != nil || (currentPom.ArtifactId != project && currentPom.ArtifactId != "jsch") {
			fmt.Printf("Could not find pom for artifact %s\n", project)
			panic(err)
		}

		createPseudoMavenProject(currentPom)

		fmt.Printf("- %s\n", project)

	}

}

func readPomFromUrl(url string) (Pom, error) {

	var pom Pom

	response, err := http.Get(url)
	if err != nil {
		return pom, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return pom, err
	}

	err = xml.Unmarshal(body, &pom)
	if err != nil {
		return pom, err
	}

	return pom, nil

}

func readPomFromRepos(artifactName, groupId, version string) (Pom, error) {

	var pom Pom
	var err error

	var url string
	// Iterate through the provided repos with the --repo tag
	for _, repo := range *secvulnMavenPomRepos {
		// Use the groupId, artifactName and version to build up the url of the pom
		s := strings.Split(groupId, ".")
		var group string
		for _, e := range s {
			group += e + "/"
		}

		url = fmt.Sprintf("%s/%s%s/%s/%s-%s.pom", repo, group, artifactName, version, artifactName, version)

		pom, err = readPomFromUrl(url)
		if pom.ArtifactId == artifactName {
			return pom, nil
		}
	}

	return pom, err

}

func createPseudoMavenProject(pom Pom) {

	var artifactName string
	// com.jcraft.jsch is an exception so need to make sure it is in the correct format
	if artifactName = pom.ArtifactId; artifactName == "jsch" {
		artifactName = "com.jcraft.jsch"
	}
	createDirectory(artifactName)

	// Using the current pom from the repo, create a stripped down pom for the pseudo maven project
	createPom(pom, artifactName)

	// Add this project to the list of completed projects so we don't reprocess and duplicate
	completedProjects = append(completedProjects, artifactName)
}

func createDirectory(artifactName string) {
	if err := os.Mkdir(fmt.Sprintf("%s/%s", secvulnMavenParentDir, artifactName), os.ModePerm); err != nil {
		fmt.Printf("Unable to create directory for artifact %s - %v\n", artifactName, err)
	}
}

func createPom(pom Pom, artifactName string) {
	newPom := &Pom{}

	newPom.GroupId = "dev.galasa"
	newPom.ArtifactId = artifactName
	newPom.Version = pom.Version
	newPom.Packaging = "jar"

	newPom.Parent = &Parent{
		GroupId:    "dev.galasa",
		ArtifactId: "security-scanning",
		Version:    "0.0.1",
	}

	var dependencies []Dependency

	for _, dep := range pom.Dependencies.Dependencies {
		groupId := dep.GroupId
		if groupId == "dev.galasa" {
			artifactId := dep.ArtifactId
			version := dep.Version
			dependency := &Dependency{
				GroupId:    groupId,
				ArtifactId: artifactId,
				Version:    version,
			}
			dependencies = append(dependencies, *dependency)

			// If a pseudo maven project hasn't been made for this dependency, add to the to do list
			if checkIfCompleted(artifactId) == false {
				toDoProjects = append(toDoProjects, artifactId)
			}
		}
	}

	if len(dependencies) > 0 {
		newPom.Dependencies = &Dependencies{
			Dependencies: dependencies,
		}
	}

	filename := fmt.Sprintf("%s/%s/%s", secvulnMavenParentDir, artifactName, "pom.xml")
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Unable to create pom.xml for artifact %s\n", artifactName)
		panic(err)
	}

	xmlWriter := io.Writer(file)

	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("  ", "    ")
	err = enc.Encode(newPom)
	if err != nil {
		fmt.Printf("Unable to encode the pom.xml for artifact %s\n", artifactName)
		panic(err)
	}
}

func updateParent() {
	securityScanningPom := &Pom{}

	securityScanningPom.GroupId = "dev.galasa"
	securityScanningPom.ArtifactId = "security-scanning"
	securityScanningPom.Version = "0.0.1"
	securityScanningPom.Packaging = "pom"

	var array []string
	sort.Strings(completedProjects)
	for _, project := range completedProjects {
		array = append(array, project)
	}

	securityScanningPom.Modules = &Modules{
		Module: array,
	}

	// Add OSS Index Maven Plugin
	securityScanningPom.Build = &Build{}
	securityScanningPom.Build.Plugins.Plugin.GroupId = "org.sonatype.ossindex.maven"
	securityScanningPom.Build.Plugins.Plugin.ArtifactId = "ossindex-maven-plugin"
	securityScanningPom.Build.Plugins.Plugin.Version = "3.1.0"
	securityScanningPom.Build.Plugins.Plugin.Executions.Execution.Id = "audit-dependencies"
	securityScanningPom.Build.Plugins.Plugin.Executions.Execution.Phase = "validate"
	securityScanningPom.Build.Plugins.Plugin.Executions.Execution.Goals.Goal = "audit"
	securityScanningPom.Build.Plugins.Plugin.Configuration.ReportFile = "${project.build.directory}/audit-report.json"
	securityScanningPom.Build.Plugins.Plugin.Configuration.Fail = "false"

	filename := fmt.Sprintf("%s/%s", secvulnMavenParentDir, "pom.xml")
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Unable to create pom.xml for security scanning project\n")
		panic(err)
	}

	xmlWriter := io.Writer(file)

	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("  ", "    ")
	err = enc.Encode(securityScanningPom)
	if err != nil {
		fmt.Printf("Unable to encode the pom.xml for security scanning project\n")
		panic(err)
	}

}

func checkIfCompleted(a string) bool {
	for _, b := range completedProjects {
		if b == a {
			return true
		}
	}
	return false
}

func getVersion(artifactName string, pom Pom) string {
	for _, dep := range pom.Dependencies.Dependencies {
		if dep.ArtifactId == artifactName {
			return dep.Version
		}
	}
	return ""
}

func getGroupId(artifactName string, pom Pom) string {
	for _, dep := range pom.Dependencies.Dependencies {
		if dep.ArtifactId == artifactName {
			return dep.GroupId
		}
	}
	return ""
}
