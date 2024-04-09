/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package cmd

import (
	"github.com/dev-galasa/buildutils/openapi2beans/pkg/generator"
	"github.com/dev-galasa/buildutils/openapi2beans/pkg/utils"
	galasaUtils "github.com/galasa-dev/cli/pkg/utils"
	"github.com/spf13/cobra"
)

func NewGenerateCommand(factory utils.Factory, flags Openapi2beansFlagStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generates java from openapi yaml",
		Long:  "command used to generate java from an openapi yaml input.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeGenerateCmd(factory, &flags)
		},
	}

	addFlags(cmd, &flags)
	cmd.Flags().BoolP("help", "h", false, "Displays the options for the 'openapi2beans' command.")
	cmd.MarkPersistentFlagRequired("yaml")
	cmd.MarkPersistentFlagRequired("package")
	cmd.MarkPersistentFlagRequired("output")

	return cmd
}

func executeGenerateCmd(factory utils.Factory, flags *Openapi2beansFlagStore) error {
	var err error
	fs := factory.GetFileSystem()
	err = galasaUtils.CaptureLog(fs, flags.logFileName)
	if err == nil {
		err = generator.GenerateFiles(fs, flags.storeFilePath, flags.apiFilePath, flags.packageName)
	}
	return err
}