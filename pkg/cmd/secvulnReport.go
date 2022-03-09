//
// Copyright contributors to the Galasa project
//

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
    secvulnReportCmd = &cobra.Command{
		Use:   "report",
		Short: "Generate report",
		Long:  "Generate report",
		Run:   secvulnReportExecute,
	}

	secvulnReportExtracts        *[]string
	secvulnReportAcceptance      string
	secvulnReportOutput          string
)

func init() {
	secvulnReportExtracts = secvulnReportCmd.PersistentFlags().StringArray("extract", nil, "Extract yaml files")
	secvulnReportCmd.PersistentFlags().StringVar(&secvulnReportAcceptance, "acceptance", "", "Acceptance yaml URL")
	secvulnReportCmd.PersistentFlags().StringVar(&secvulnOssindexOutput, "output", "", "Output markdown file")

	secvulnReportCmd.MarkPersistentFlagRequired("extract")
	secvulnReportCmd.MarkPersistentFlagRequired("acceptance")
	secvulnReportCmd.MarkPersistentFlagRequired("output")

	secvulnCmd.AddCommand(secvulnReportCmd)
}

func secvulnReportExecute(cmd *cobra.Command, args []string) {
	fmt.Printf("Galasa Build - Security Vulnerability Report - version %v\n", rootCmd.Version)
}
