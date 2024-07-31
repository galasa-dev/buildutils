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

var versionSuffix string

var (
	versioningSuffixSetCmd = &cobra.Command{
		Use:   "set",
		Short: "Setting the build version suffix of source code.",
		Long:  "Sets the source module versions suffix recursively.",
		Run:   versioningSuffixSetExecute,
	}
)

func init() {
	// --suffix flag. Optional. If missing, assumed that no suffix is wanted.
	versioningSuffixSetCmd.PersistentFlags().StringVarP(&versionSuffix, "suffix", "s", "-SNAPSHOT",
		"The version suffix to set all modules to use. For example -SNAPSHOT"+
			" Suffixes must start with '_' or '-' ")
	versioningSuffixSetCmd.MarkPersistentFlagRequired("suffix")

	versioningSuffixCmd.AddCommand(versioningSuffixSetCmd)
}

func versioningSuffixSetExecute(cmd *cobra.Command, args []string) {

	fs := utils.NewOSFileSystem()
	err := versioning.SuffixSetExecute(fs, sourceCodeFolderPath, versionSuffix)

	if err != nil {
		panic(err)
	}

}
