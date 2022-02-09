//
// Copyright contributors to the Galasa project
//

package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
    mavenDeployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy local maven repository to remote one",
		Long:  "Deploy local maven artifacts at the set version to a remote maven repository",
		Run:   mavenDeployExecute,
	}

	mavenDeployDirectory   string
	mavenDeployGroup       string
	mavenDeployVersion     string
)

func init() {
	mavenDeployCmd.PersistentFlags().StringVarP(&mavenDeployDirectory, "local", "", "", "local repository")
	mavenDeployCmd.PersistentFlags().StringVarP(&mavenDeployGroup, "group", "", "", "groupId to deploy")
	mavenDeployCmd.PersistentFlags().StringVarP(&mavenDeployVersion, "version", "", "", "version to deploy")

	mavenDeployCmd.MarkPersistentFlagRequired("local")
	mavenDeployCmd.MarkPersistentFlagRequired("group")
	mavenDeployCmd.MarkPersistentFlagRequired("version")

	mavenCmd.AddCommand(mavenDeployCmd)
}

func mavenDeployExecute(cmd *cobra.Command, args []string) {
	fmt.Printf("Galasa Build - Maven Deploy - version %v\n", rootCmd.Version)

	basicAuth, err := mavenGetBasicAuth()
    if err != nil {
        panic(err)
    }

	groupDir := strings.ReplaceAll(mavenDeployGroup, ".", "/")


	// Search local directory for dev.galasa artifacts at the correct version

	mavenBaseDirectory := mavenDeployDirectory + "/" + groupDir

	artifactDirectories, err := ioutil.ReadDir(mavenBaseDirectory)
	if (err != nil) {
		panic(err)
	}

	artifacts := []string{}
	client := &http.Client{}

	for _, potentialArtifact := range artifactDirectories {
		//Check this is a artifact directory and not a subgroup
		artifactDirectory := mavenBaseDirectory + "/" + potentialArtifact.Name()
		if _, err := os.Stat(artifactDirectory + "/maven-metadata.xml"); err == nil {
		} else if errors.Is(err, os.ErrNotExist) {
			continue
		} else {
			panic(err)
		}

		//Check this is a artifact is at the correct version
		if _, err := os.Stat(artifactDirectory + "/" + mavenDeployVersion); err == nil {
		} else if errors.Is(err, os.ErrNotExist) {
			continue
		} else {
			panic(err)
		}

		artifacts = append(artifacts, potentialArtifact.Name())
	}

	if len(artifacts) < 1 {
		fmt.Println("No artifacts found to be deployed")
		os.Exit(0)
	}

	sort.Strings(artifacts)

	// Now deploy the contents of the artifact version directory

	for _, artifact := range artifacts {
		fmt.Printf("Deploying %v/%v/%v\n", mavenDeployGroup, artifact, mavenDeployVersion)

		versionDirectory := mavenBaseDirectory + "/" + artifact + "/" + mavenDeployVersion
		versionArtifacts, err := ioutil.ReadDir(versionDirectory)
		if (err != nil) {
			panic(err)
		}

		for _, artifactFile := range versionArtifacts {
			fmt.Printf("    %v\n", artifactFile.Name())

			filename := versionDirectory + "/" + artifactFile.Name()

			url := mavenRepository + "/" + groupDir + "/" + artifact + "/" + mavenDeployVersion + "/" + artifactFile.Name()

			file, err := os.Open(filename)
			defer file.Close()

			req, err := http.NewRequest("PUT", url, file)
			if err != nil {
				panic(nil)
			}
			
			req.Header.Set("Authorization", basicAuth)

			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
		
			if resp.StatusCode != http.StatusCreated {
				fmt.Printf("Put for artifact for url %v - status line - %v\n", url, resp.Status);
				os.Exit(1)
			}
		}

		if len(artifacts) == 1 {
			fmt.Printf("Complete - 1 artifact deployed\n")
		} else {
			fmt.Printf("Complete - %v artifacts deployed\n", len(artifacts))
		}
	
	}

}