//
// Copyright contributors to the Galasa project
//

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
    secvulnMavenCmd = &cobra.Command{
		Use:   "maven",
		Short: "Generate psuedo maven project for security vulnerability scanning",
		Long:  "Generate psuedo maven project for security vulnerability scanning",
		Run:   secvulnMavenExecute,
	}

	secvulnMavenParentDir     string
	secvulnMavenPomUrls    *[]string
)

func init() {
	secvulnMavenCmd.PersistentFlags().StringVar(&secvulnMavenParentDir, "parent", "", "Parent project directory")
	secvulnMavenPomUrls = secvulnMavenCmd.PersistentFlags().StringArray("pom", nil, "Component Pom URLs")

	secvulnMavenCmd.MarkPersistentFlagRequired("parent")
	secvulnMavenCmd.MarkPersistentFlagRequired("pom")

	secvulnCmd.AddCommand(secvulnMavenCmd)
}

func secvulnMavenExecute(cmd *cobra.Command, args []string) {
	fmt.Printf("Galasa Build - Security Vulnerability Maven - version %v\n", rootCmd.Version)
}
