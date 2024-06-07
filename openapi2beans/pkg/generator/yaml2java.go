/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package generator

import (
	"log"
	"strings"

	openapi2beans_errors "github.com/dev-galasa/buildutils/openapi2beans/pkg/errors"
	"github.com/galasa-dev/cli/pkg/files"
)

var filepathSeparator = "/"

func GenerateFiles(fs files.FileSystem, projectFilePath string, apiFilePath string, packageName string, force bool) error {
	var fatalErr error
	var apiyaml string
	var errList map[string]error
	filepathSeparator = fs.GetFilePathSeparator()

	apiyaml, fatalErr = fs.ReadTextFile(apiFilePath)
	if fatalErr == nil {
		var schemaTypes map[string]*SchemaType
		schemaTypes, errList, fatalErr = getSchemaTypesFromYaml([]byte(apiyaml))
		if fatalErr == nil {
			if len(errList) > 0 {
				fatalErr = handleErrList(errList)
			}
			if fatalErr == nil {
				storeFilepath := generateStoreFilepath(projectFilePath, packageName)
				fatalErr = generateDirectories(fs, storeFilepath, force)
				if fatalErr == nil {
					javaPackage := translateSchemaTypesToJavaPackage(schemaTypes, packageName)
					convertJavaPackageToJavaFiles(javaPackage, fs, storeFilepath)
				}
			}
		}
	}

	return fatalErr
}

// Cleans or creates the store folder at the storeFilepath
func generateDirectories(fs files.FileSystem, storeFilepath string, force bool) error {
	log.Println("Cleaning generated beans directory: " + storeFilepath)
	exists, err := fs.DirExists(storeFilepath)
	if err == nil {
		if exists {
			if !force {
				err = openapi2beans_errors.NewError("generateDirectories: files located in directory requested to to produce beans in: %s", storeFilepath)
			}
			if err == nil {
				err = deleteAllJavaFiles(fs, storeFilepath)
			}
		} else {
			log.Printf("Creating output directory: %s\n", storeFilepath)
			err = fs.MkdirAll(storeFilepath)
		}
	}
	return err
}

func handleErrList(errList map[string]error) error {
	log.Println("Failing on non-fatal errors:")
	var err error
	errorString := ""
	for _, individualError := range errList {
		errorString += individualError.Error()
	}
	err = openapi2beans_errors.NewError(errorString)
	return err
}

// Creates the store filepath from the output filepath + the package name seperated out into folders
func generateStoreFilepath(outputFilepath string, packageName string) string {
	packageFilepath := strings.ReplaceAll(packageName, ".", filepathSeparator)
	if outputFilepath[len(outputFilepath)-1:] != filepathSeparator {
		outputFilepath += filepathSeparator
	}
	return outputFilepath + packageFilepath
}

func deleteAllJavaFiles(fs files.FileSystem, storeFilepath string) error {
	filepaths, err := fs.GetAllFilePaths(storeFilepath)
	for _, filepath := range filepaths {
		relativePath := filepath[len(storeFilepath)+1:]
		if len(relativePath) - 5 > 0 {
			if relativePath[len(relativePath) - 5 :] == ".java" && !strings.Contains(relativePath, filepathSeparator){
				fs.DeleteFile(filepath)
			}
		}
	}
	return err
}
