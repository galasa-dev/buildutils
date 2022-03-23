//
// Copyright contributors to the Galasa project
//

package cmd

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"net/http"

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
	secvulnMavenPomRepos string // TO DO - change to array

	modules []string

	completedProjects []string
	toDoProjects []string
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
	mainPom := readPomFromRepo(secvulnMavenParentDir + mainPomName)

	createPseudoMavenProject(mainPom.ArtifactId, mainPom)
	fmt.Println("Pseudo maven project created for: " + mainPom.ArtifactId)

	// Repeat the process for all dependencies of groupId dev.galasa
	fmt.Println("Pseudo maven projects created for: ")
	for len(toDoProjects) > 0 {

		var index = len(toDoProjects) - 1
		var artifactName = toDoProjects[index]

		bool := checkIfCompleted(artifactName)
		if bool == true {
			toDoProjects = removeItemFromArrayByIndex(toDoProjects, index)
			continue
		}

		// TEMPORARY - hard code artifact managers pom to show stripped down pom
		var currentPom Pom
		if artifactName == "dev.galasa.artifact.manager" {
			currentPom = readPomFromUrl(artifactName)
		}
		// currentPom = readPomFromUrl(artifactName)

		createPseudoMavenProject(artifactName, currentPom)
		fmt.Println("- " + artifactName)

		toDoProjects = removeItemFromArrayByIndex(toDoProjects, index)

	}

	fmt.Println("Updating the pom.xml for the security scanning project")
	updateParent(mainPom)

}

func readPomFromUrl(artifactName string) Pom {
	// TO DO - Iterate through repos passed through the CLI
	// TO DO - Get url from the sample repo name and artifact name
	url := secvulnMavenPomRepos + "/dev/galasa/" + artifactName +"/" + "0.21.0" + "/" + artifactName + "-" + "0.21.0" + ".pom"
	
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	var pom Pom
	err1 := xml.Unmarshal(body, &pom)
	if err1 != nil {
		fmt.Println(err1)
	}

	return pom
}

func readPomFromRepo(repo string) Pom {
	xmlFile, err := os.Open(repo)
	if err != nil {
		fmt.Println(err)
	}

	defer xmlFile.Close()

	byteValue, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		fmt.Println(err)
	}

	var pom Pom
	err = xml.Unmarshal(byteValue, &pom)
	if err != nil {
		fmt.Println(err)
	}

	return pom
}

func createPseudoMavenProject(artifactName string, pom Pom){

	// artifactName := pom.ArtifactId

	createDirectory(artifactName)

	createPom(artifactName, pom)

	modules = append(modules, artifactName)

	completedProjects = append(completedProjects, artifactName)
}


func createDirectory(artifactName string) {
	// TO DO - Make sure it goes into the secVulnMavenParentDir
	if err := os.Mkdir(artifactName, os.ModePerm); err != nil {
		fmt.Println(err)
	}

	// Change into the sub-directory
	os.Chdir(artifactName)
}

func createPom(artifactName string, pom Pom) {
	// TO DO - get xmlns, xmlns:xsi and xsi:schemaLocation
	newPom := &NewPom{}

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
	file, _ := os.Create(filename)

	xmlWriter := io.Writer(file)

	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("  ", "    ")
	enc.Encode(newPom)

	// Change back to the parent directory
	os.Chdir("..")
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

func updateParent(pom Pom){
	securityScanningPom := &SecurityScanningPom{}

	securityScanningPom.Xmlns = pom.Xmlns
	securityScanningPom.XmlnsXsi = pom.XmlnsXsi
	securityScanningPom.XsiSchemaLocation = pom.XsiSchemaLocation

	securityScanningPom.GroupId = "dev.galasa"
	securityScanningPom.ArtifactId = "security-scanning"
	securityScanningPom.Version = "0.21.0"
	securityScanningPom.Packaging = "pom"

	sort.Strings(modules)
	for i := 0; i < len(modules); i++ {
		securityScanningPom.Modules.Module = append(securityScanningPom.Modules.Module, modules[i])
	}

	filename := "pom.xml"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	}

	xmlWriter := io.Writer(file)

	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("  ", "    ")
	err = enc.Encode(securityScanningPom)
	if err != nil {
		fmt.Println(err)
	}

}