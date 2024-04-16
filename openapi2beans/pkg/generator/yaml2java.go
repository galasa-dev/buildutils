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

func GenerateFiles(fs files.FileSystem, projectFilePath string, apiFilePath string, packageName string) error {
	var fatalErr error
	var apiyaml string
	var errList map[string]error

	storeFilePath := generateStoreFilePath(projectFilePath, packageName)
	fatalErr = generateDirectories(fs, storeFilePath)
	if fatalErr == nil {
		apiyaml, fatalErr = fs.ReadTextFile(apiFilePath)
		if fatalErr == nil {
			var schemaTypes map[string]*SchemaType
			schemaTypes, errList, fatalErr = getSchemaTypesFromYaml([]byte(apiyaml))
			if fatalErr == nil || len(errList) > 0 {
				javaPackage := translateSchemaTypesToJavaPackage(schemaTypes, packageName)
				convertJavaPackageToJavaFiles(javaPackage, fs, storeFilePath)
			}
		}
	}

	handleErrList(errList)
	return fatalErr
}

// Cleans and/or creates the store file
func generateDirectories(fs files.FileSystem, storeFilePath string) error {
	log.Println("Cleaning generated beans directory: " + storeFilePath)
	exists, err := fs.DirExists(storeFilePath)
	if err == nil {
		if exists {
			fs.DeleteDir(storeFilePath)
		}
		log.Printf("Creating output directory: %s\n", storeFilePath)
		err = fs.MkdirAll(storeFilePath)
	}
	return err
}

func handleErrList(errList map[string]error) error {
	log.Println("Failing on non-fatal errors:")
	var err error
	errorString := ""
	for _, individualError := range errList {
		errorString += "Error: " + individualError.Error()
	}
	err = openapi2beans_errors.NewError(errorString)
	return err
}

func generateStoreFilePath(projectFilePath string, packageName string) string {
	packageFilePath := strings.ReplaceAll(packageName, ".", "/")
	if projectFilePath[len(projectFilePath)-1:] != "/" {
		projectFilePath += "/"
	}
	return projectFilePath + packageFilePath
}
