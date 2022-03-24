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
	fmt.Println("Security scanning project pom.xml updated")

}

func startScanningPipeline(mainPomUrl string) {

	fmt.Println("Starting the scanning pipeline at " + mainPomUrl)

	mainPom, err := readPomFromUrl(mainPomUrl)
	if err != nil {
		fmt.Println("Unable to find the main pom at " + mainPomUrl)
		panic(err)
	}

	createPseudoMavenProject(mainPom)
	fmt.Println("Pseudo maven project created for: " + mainPom.ArtifactId)

	// Repeat the process for all dependencies with groupId dev.galasa
	fmt.Println("Pseudo maven projects created for dependency chain of " + mainPom.ArtifactId + ":")
	for len(toDoProjects) > 0 {

		var index = len(toDoProjects) - 1
		var artifactName = toDoProjects[index]

		if artifactName == "com.jcraft.jsch" {
			toDoProjects = removeItemFromArrayByIndex(toDoProjects, index)
			continue
		}

		if bool := checkIfCompleted(artifactName); bool == true {
			toDoProjects = removeItemFromArrayByIndex(toDoProjects, index)
			continue
		}

		var version string
		if version = getVersion(artifactName, mainPom); version == "" {
			fmt.Printf("Unable to get version for artifact %s \n", artifactName)
			panic(nil)
		}
		var currentPom Pom
		var err error
		for _, repo := range *secvulnMavenPomRepos {
			currentPom, err = readPomFromRepo(repo, artifactName, version)
			if currentPom.ArtifactId == artifactName {
				break
			}
		}
		if err != nil {
			fmt.Printf("Could not find pom for artifact %s \n", artifactName)
			panic(err)
		}

		createPseudoMavenProject(currentPom)

		toDoProjects = removeItemFromArrayByIndex(toDoProjects, index)

		fmt.Println("- " + artifactName)
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

func readPomFromRepo(repo, artifactName, version string) (Pom, error) {

	var url string
	if repo == "https://galasadev-cicsk8s.hursley.ibm.com/main/maven/obr" {
		url = fmt.Sprintf("%s/dev/galasa/%s/%s/%s-%s.pom", repo, artifactName, version, artifactName, version)

	} else if repo == "https://repo.maven.apache.org/maven2" {
		// Future proofing for artifacts like com.jcraft.jsch - can be improved
		s := strings.Split(artifactName, ".")
		var artifact string
		for i := 0; i < len(s); i++ {
			artifact += s[i] + "/"
		}
		url = fmt.Sprintf("%s/%s%s/%s-%s.pom", repo, artifact, version, s[len(s)-1], version)
	} else {
		fmt.Println("Invalid repositories provided to galasabld")
		panic(nil)
	}

	return readPomFromUrl(url)

}

func createPseudoMavenProject(pom Pom) {

	createDirectory(pom.ArtifactId)

	createPom(pom)

	completedProjects = append(completedProjects, pom.ArtifactId)
}

func createDirectory(artifactName string) {

	if err := os.Mkdir(fmt.Sprintf("%s/%s", secvulnMavenParentDir, artifactName), os.ModePerm); err != nil {
		fmt.Printf("Unable to create directory for artifact %s - %v", artifactName, err)
	}

	// Change into the sub-directory
	os.Chdir(fmt.Sprintf("%s/%s", secvulnMavenParentDir, artifactName))
}

func createPom(pom Pom) {
	newPom := &Pom{}

	newPom.GroupId = pom.GroupId
	newPom.ArtifactId = pom.ArtifactId
	newPom.Version = pom.Version
	newPom.Packaging = "jar"

	newPom.Parent = &Parent{
		GroupId:    "dev.galasa",
		ArtifactId: "security-scanning",
		Version:    "0.21.0",
	}

	var dependencies []Dependency
	for i := 0; i < len(pom.Dependencies.Dependencies); i++ {
		groupId := pom.Dependencies.Dependencies[i].GroupId
		if groupId == "dev.galasa" {
			artifactId := pom.Dependencies.Dependencies[i].ArtifactId
			version := pom.Dependencies.Dependencies[i].Version
			dependency := &Dependency{
				GroupId:    groupId,
				ArtifactId: artifactId,
				Version:    version,
			}
			dependencies = append(dependencies, *dependency)

			// If a pseudo maven project hasn't been made for this dependency, add to the to do list
			bool := checkIfCompleted(artifactId)
			if bool == false {
				toDoProjects = append(toDoProjects, artifactId)
			}
		}
	}

	if len(dependencies) > 0 {
		newPom.Dependencies = &Dependencies{
			Dependencies: dependencies,
		}
	}

	filename := "pom.xml"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Unable to create pom.xml for artifact %s", pom.ArtifactId)
		panic(err)
	}

	xmlWriter := io.Writer(file)

	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("  ", "    ")
	err = enc.Encode(newPom)
	if err != nil {
		fmt.Printf("Unable to encode the pom.xml for artifact %s", pom.ArtifactId)
		panic(err)
	}

	// Change back to the parent directory
	os.Chdir(secvulnMavenParentDir)
}

func updateParent() {
	securityScanningPom := &Pom{}

	securityScanningPom.GroupId = "dev.galasa"
	securityScanningPom.ArtifactId = "security-scanning"
	securityScanningPom.Version = "0.21.0"
	securityScanningPom.Packaging = "pom"

	var array []string
	sort.Strings(completedProjects)
	for i := 0; i < len(completedProjects); i++ {
		array = append(array, completedProjects[i])
	}

	securityScanningPom.Modules = &Modules{
		Module: array,
	}

	filename := "pom.xml"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Unable to create pom.xml for security scanning project")
		panic(err)
	}

	xmlWriter := io.Writer(file)

	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("  ", "    ")
	err = enc.Encode(securityScanningPom)
	if err != nil {
		fmt.Println("Unable to encode the pom.xml for security scanning project")
		panic(err)
	}

}

func removeItemFromArrayByIndex(array []string, index int) []string {
	return append(array[:index], array[index+1:]...)
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
	for i := 0; i < len(pom.Dependencies.Dependencies); i++ {
		if pom.Dependencies.Dependencies[i].ArtifactId == artifactName {
			return pom.Dependencies.Dependencies[i].Version
		}
	}
	return ""
}
