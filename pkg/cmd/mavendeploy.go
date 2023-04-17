/*
 * Copyright contributors to the Galasa project
 */

package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"galasa.dev/buildUtilities/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	mavenDeployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy local maven repository to remote one",
		Long:  "Deploy local maven artifacts at the set version to a remote maven repository",
		Run:   executeMavenDeploy,
	}

	mavenDeployDirectory string
	mavenDeployGroup     string
	mavenDeployVersion   string
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

func executeMavenDeploy(cmd *cobra.Command, args []string) {
	fmt.Printf("Galasa Build - Maven Deploy - version %v\n", rootCmd.Version)

	basicAuth, err := mavenGetBasicAuth()
	if err != nil {
		panic(err)
	}

	fileSystem := utils.NewOSFileSystem()

	mavenRepository = strings.TrimRight(mavenRepository, string(os.PathSeparator))

	err = mavenDeploy(fileSystem, mavenRepository, mavenDeployDirectory, mavenDeployGroup, mavenDeployVersion, basicAuth)
	if err != nil {
		panic(err)
	}
}

// Checks the given local repository for artifacts and deploys the identified artifacts to the remote Maven repository
func mavenDeploy(
	fileSystem utils.FileSystem,
	mavenRepository string,
	mavenDeployDirectory string,
	mavenDeployGroup string,
	mavenDeployVersion string,
	basicAuth string) error {

	groupDir := strings.ReplaceAll(mavenDeployGroup, ".", string(os.PathSeparator))
	mavenBaseDirectory := path.Join(mavenDeployDirectory, groupDir)

	artifactDirectories, err := fileSystem.ReadDir(mavenBaseDirectory)
	if err != nil {
		return err
	}

	// Create a map of artifacts. Keys correspond to artifact names and values correspond to the paths to
	// the artifacts' version directories
	artifacts := make(map[string]string)

	for _, potentialArtifact := range artifactDirectories {

		// Check if this is an artifact directory and not a subgroup
		artifactName := potentialArtifact.Name()
		artifactDirectory := path.Join(mavenBaseDirectory, artifactName)
		mavenMetadataFileName := "maven-metadata.xml"

		mavenMetadataExists, err := fileSystem.Exists(path.Join(artifactDirectory, mavenMetadataFileName))
		if !mavenMetadataExists {
			mavenMetadataPath := matchFileInDirectory(fileSystem, artifactDirectory, mavenMetadataFileName)

			// No maven-metadata.xml file found within artifact directory, move on to the next artifact
			if mavenMetadataPath == "" {
				continue
			}

		} else if err != nil {
			return err
		}

		// Check if this artifact is at the correct version
		artifactVersionPath := path.Join(artifactDirectory, mavenDeployVersion)
		versionDirectoryExists, err := fileSystem.DirExists(artifactVersionPath)
		if !versionDirectoryExists {
			artifactVersionPath = matchFileInDirectory(fileSystem, artifactDirectory, mavenDeployVersion)

			// No version directory found within the artifact directory, move on to the next artifact
			if artifactVersionPath == "" {
				continue
			}

		} else if err != nil {
			return err
		}

		artifacts[artifactName] = artifactVersionPath
	}

	if len(artifacts) < 1 {
		fmt.Println("No artifacts found to deploy")
		return err
	}

	// Now deploy the contents of the artifact version directories
	err = deployArtifacts(fileSystem, mavenRepository, mavenDeployGroup, mavenDeployVersion, artifacts, basicAuth)

	return err
}

// Deploys the given artifacts to a given Maven repository
func deployArtifacts(
	fileSystem utils.FileSystem,
	mavenRepository string,
	mavenDeployGroup string,
	mavenDeployVersion string,
	artifacts map[string]string,
	basicAuth string) error {

	var err error = nil
	client := &http.Client{}
	groupDir := strings.ReplaceAll(mavenDeployGroup, ".", string(os.PathSeparator))

	for artifactName, artifactVersionPath := range artifacts {
		fmt.Printf("Deploying %v/%v/%v\n", mavenDeployGroup, artifactName, mavenDeployVersion)

		versionArtifacts, err := fileSystem.ReadDir(artifactVersionPath)
		if err == nil {

			// Go through each file within the artifact's version directory and send a PUT request to deploy to the
			// remote Maven repository
			for _, artifactFile := range versionArtifacts {
				fmt.Printf("    %v\n", artifactFile.Name())

				artifactFilePath := path.Join(artifactVersionPath, artifactFile.Name())
				artifactPathFromGroup := artifactFilePath[strings.Index(artifactFilePath, groupDir):]

				url, err := url.JoinPath(mavenRepository, artifactPathFromGroup)
				if err == nil {
					file, err := fileSystem.Open(artifactFilePath)
					if err == nil {
						err = putMavenArtifact(url, file, client, basicAuth)
                        if err != nil {
                            return err
                        }
					}
				}
			}
		}
	}

	if len(artifacts) == 1 {
		fmt.Printf("Complete - 1 artifact deployed\n")
	} else {
		fmt.Printf("Complete - %v artifacts deployed\n", len(artifacts))
	}

	return err
}

// Sends a PUT request to a Maven repository to upload an artifact to it
func putMavenArtifact(
	mavenRepoUrl string,
	readCloser io.ReadCloser,
	client *http.Client,
	basicAuth string) error {

	defer readCloser.Close()

	// Create the PUT request
	req, err := http.NewRequest("PUT", mavenRepoUrl, readCloser)
	if err == nil {
		req.Header.Set("Authorization", basicAuth)

		// Send the PUT request
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				return fmt.Errorf("put for artifact for url %v - status line - %v", mavenRepoUrl, resp.Status)
			}
		}
	}

	return err
}

// Walks through a given directory, searching for a given file or directory name.
// Returns the path to the matching file or directory, or an empty string if no match was found.
func matchFileInDirectory(fileSystem utils.FileSystem, dirPath string, targetFileName string) string {
	var matchedPath string = ""
	_ = fileSystem.WalkDir(dirPath, func(path string, file os.DirEntry, err error) error {
		if file.Name() == targetFileName {
			matchedPath = path

			// Match found, no need to continue walking through the directory
			return fs.SkipAll
		}

		return nil
	})

	return matchedPath
}
