/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */

package cmd

import (
	"galasa.dev/buildUtilities/pkg/utils"
	"galasa.dev/buildUtilities/pkg/versioning"
	"github.com/spf13/cobra"
)

var (
	versioningListCmd = &cobra.Command{
		Use:   "list",
		Short: "Clears the build version suffix of source code.",
		Long:  "Removes the source module versions recursively.",
		Run:   versioningListExecute,
	}
)

func init() {
	versioningCmd.AddCommand(versioningListCmd)
}

func versioningListExecute(cmd *cobra.Command, args []string) {

	fs := utils.NewOSFileSystem()
	err := versioning.ListExecute(fs, sourceCodeFolderPath)

	if err != nil {
		panic(err)
	}

}
