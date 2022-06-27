/*
 * Copyright contributors to the Galasa project
 */

package cmd

import (
	"fmt"
	"regexp"
)

func arrayContainsString(targetString string, array []string) bool {
	for _, element := range array {
		if element == targetString {
			return true
		}
	}
	return false
}

func getRegexSubmatches(fullString string) []string {
	regex := "[a-zA-Z0-9._-]+"
	re := regexp.MustCompile(regex)
	submatches := re.FindAllString(fullString, -1)

	return submatches
}

func getGroupAndArtifact(fullString string) string {
	submatches := getRegexSubmatches(fullString)
	return fmt.Sprintf("%s:%s", submatches[0], submatches[1])
}
