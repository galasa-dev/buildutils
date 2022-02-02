//
// Copyright contributors to the Galasa project 
//

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "galasabld",
	Short: "Build utilities for Galasa",
	Long:  "",
	Version: "0.0.7",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
