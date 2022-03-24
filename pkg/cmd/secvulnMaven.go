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
	secvulnMavenPomRepos  string // TO DO - change to array

	completedProjects []string
	toDoProjects      []string
)

func init() {
	secvulnMavenCmd.PersistentFlags().StringVar(&secvulnMavenParentDir, "parent", "", "Parent project directory")
	secvulnMavenPomUrls = secvulnMavenCmd.PersistentFlags().StringArray("pom", nil, "Component Pom URLs")
	secvulnMavenCmd.PersistentFlags().StringVar(&secvulnMavenPomRepos, "repo", "", "Repo")

	secvulnMavenCmd.MarkPersistentFlagRequired("parent")
	secvulnMavenCmd.MarkPersistentFlagRequired("pom")
	secvulnMavenCmd.MarkPersistentFlagRequired("repo")

	secvulnCmd.AddCommand(secvulnMavenCmd)
}

func secvulnMavenExecute(cmd *cobra.Command, args []string) {
	fmt.Printf("Galasa Build - Security Vulnerability Maven - version %v\n", rootCmd.Version)

	// TO DO - Create security scanning project called dev.galasa:security-scanning:x.xx.x (current version from obr release.yaml)

	// FUTURE PROOF - It won't just be dev.galasa.uber.obr this command starts with, could be isolated, mvp, simplatform etc
	// TO DO - How do we tell it we want to use dev.galasa.uber.obr for the mainPomName for this run of the galasabld command?
	// Is it part of the command?
	mainPomName := "/dev.galasa.uber.obr-0.22.0.pom"

	fmt.Println("Starting the scanning pipeline from " + secvulnMavenParentDir + mainPomName)
	mainPom, err := readPomFromFile(secvulnMavenParentDir + mainPomName)
	if err != nil {
		fmt.Println("Unable to read the main pom from " + secvulnMavenParentDir + mainPomName)
		panic(err)
	}

	createPseudoMavenProject(mainPom.ArtifactId, mainPom)
	fmt.Println("Pseudo maven project created for: " + mainPom.ArtifactId)

	// Repeat the process for all dependencies of groupId dev.galasa
	fmt.Println("Pseudo maven projects created for: ")
	for len(toDoProjects) > 0 {

		var index = len(toDoProjects) - 1
		var artifactName = toDoProjects[index]

		// TO DO - Remove
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
		currentPom, err := readPomFromRepo(artifactName, version)
		if err != nil {
			fmt.Printf("Could not find pom for artifact %s \n", artifactName)
			panic(err)
		}
		createPseudoMavenProject(artifactName, currentPom)

		toDoProjects = removeItemFromArrayByIndex(toDoProjects, index)

		fmt.Println("- " + artifactName)
	}

	updateParent(mainPom)
	fmt.Println("Security scanning project pom.xml updated")

}

func readPomFromRepo(artifactName, version string) (Pom, error) {
	var pom Pom

	// TO DO - Iterate through repos passed through the CLI
	var url string
	if secvulnMavenPomRepos == "https://galasadev-cicsk8s.hursley.ibm.com/main/maven/obr" {
		url = fmt.Sprintf("%s/dev/galasa/%s/%s/%s-%s.pom", secvulnMavenPomRepos, artifactName, version, artifactName, version)
	} else if secvulnMavenPomRepos == "otherRepo" {
		url = "otherRepo"
	} else {
		url = ""
	}

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

func readPomFromFile(repo string) (Pom, error) {
	var pom Pom

	xmlFile, err := os.Open(repo)
	if err != nil {
		return pom, err
	}

	defer xmlFile.Close()

	byteValue, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		return pom, err
	}

	err = xml.Unmarshal(byteValue, &pom)
	if err != nil {
		return pom, err
	}

	return pom, nil
}

func createPseudoMavenProject(artifactName string, pom Pom) {

	// artifactName := pom.ArtifactId

	createDirectory(artifactName)

	createPom(artifactName, pom)

	completedProjects = append(completedProjects, artifactName)
}

func createDirectory(artifactName string) {

	if err := os.Mkdir(fmt.Sprintf("%s/%s", secvulnMavenParentDir, artifactName), os.ModePerm); err != nil {
		// if err := os.Mkdir(artifactName, os.ModePerm); err != nil {
		fmt.Println(err)
		return
	}

	// Change into the sub-directory
	os.Chdir(artifactName)
}

func createPom(artifactName string, pom Pom) {
	// TO DO - get xmlns, xmlns:xsi and xsi:schemaLocation
	newPom := &Pom{}

	newPom.Xmlns = pom.Xmlns
	newPom.XmlnsXsi = pom.XmlnsXsi
	newPom.XsiSchemaLocation = pom.XsiSchemaLocation

	newPom.GroupId = pom.GroupId
	newPom.ArtifactId = artifactName
	newPom.Version = pom.Version
	newPom.Packaging = "jar"

	newPom.Parent.GroupId = "dev.galasa"
	newPom.Parent.ArtifactId = "security-scanning"
	newPom.Parent.Version = "0.21.0"

	for i := 0; i < len(pom.Dependencies.Dependencies); i++ {
		groupId := pom.Dependencies.Dependencies[i].GroupId
		if groupId == "dev.galasa" {
			artifactId := pom.Dependencies.Dependencies[i].ArtifactId
			version := pom.Dependencies.Dependencies[i].Version
			newPom.Dependencies.addDependency(groupId, artifactId, version)

			// If a pseudo maven project hasn't been made for this dependency, add to the to do list
			bool := checkIfCompleted(artifactId)
			if bool == false {
				toDoProjects = append(toDoProjects, artifactId)
			}
		}
	}

	filename := "pom.xml"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Unable to create pom.xml for artifact %s", artifactName)
		panic(err)
	}

	xmlWriter := io.Writer(file)

	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("  ", "    ")
	err = enc.Encode(newPom)
	if err != nil {
		fmt.Printf("Unable to encode the pom.xml for artifact %s", artifactName)
		panic(err)
	}

	// Change back to the parent directory
	os.Chdir(secvulnMavenParentDir)
}

func updateParent(pom Pom) {
	securityScanningPom := &Pom{}

	securityScanningPom.Xmlns = pom.Xmlns
	securityScanningPom.XmlnsXsi = pom.XmlnsXsi
	securityScanningPom.XsiSchemaLocation = pom.XsiSchemaLocation

	securityScanningPom.GroupId = "dev.galasa"
	securityScanningPom.ArtifactId = "security-scanning"
	securityScanningPom.Version = "0.21.0"
	securityScanningPom.Packaging = "pom"

	sort.Strings(completedProjects)
	for i := 0; i < len(completedProjects); i++ {
		securityScanningPom.Modules.Module = append(securityScanningPom.Modules.Module, completedProjects[i])
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

func (s *Dependencies) addDependency(groupId, artifactId, version string) {
	dependency := Dependency{GroupId: groupId, ArtifactId: artifactId, Version: version}
	s.Dependencies = append(s.Dependencies, dependency)
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
