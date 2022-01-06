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
}

var version = "0.0.3"

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
