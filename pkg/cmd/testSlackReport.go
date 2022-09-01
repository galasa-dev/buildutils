/*
 * Copyright contributors to the Galasa project
 */
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"galasa.dev/buildUtilities/pkg/galasajson"
	"github.com/spf13/cobra"
)

var (
	testSlackReportCmd = &cobra.Command{
		Use:   "slackPost",
		Short: "Slack post of the failing tests from a galasactl runs submit report",
		Long:  "",
		Run:   testSlackReportExecute,
	}
	testReportPath string
	slackWebhook   string
)

func init() {
	testSlackReportCmd.PersistentFlags().StringVar(&testReportPath, "path", "", "Path to the galasactl report")
	testSlackReportCmd.PersistentFlags().StringVar(&slackWebhook, "hook", "", "Webhook to post to slack")

	testSlackReportCmd.MarkPersistentFlagRequired("path")
	testSlackReportCmd.MarkPersistentFlagRequired("hook")

	rootCmd.AddCommand(testSlackReportCmd)
}

func testSlackReportExecute(cmd *cobra.Command, args []string) {
	fmt.Printf("Galasa Build - Slack Test Report - version %v\n", rootCmd.Version)
	failures := 0
	var failingTests []string

	report, err := unmarshalReport()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, test := range report.Tests {
		if test.Result != "Passed" {
			failures++
			classNameFull := strings.Split(test.Class, ".")
			failure := fmt.Sprintf("\t%s %s: %s\n", test.Name, classNameFull[len(classNameFull)-1], test.Result)
			failingTests = append(failingTests, failure)
		}
	}

	content := fmt.Sprintf("Galasa Full Regression Testing - Failure Report\nFailures: %v\n", failures)
	for _, f := range failingTests {
		content += f
	}

	client := http.Client{
		Timeout: time.Second * 30,
	}

	body := fmt.Sprintf("{\"text\":\"%s\"}", content)

	resp, err := client.Post(slackWebhook, "application/json", strings.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stdout, resp.Status)
}

func unmarshalReport() (galasajson.Results, error) {

	jsonReport, _ := os.ReadFile(testReportPath)

	var report galasajson.Results
	err := json.Unmarshal([]byte(jsonReport), &report)
	if err != nil {
		return report, err
	}

	if len(report.Tests) <= 0 {
		return report, fmt.Errorf("No results found in file %s", testReportPath)
	}
	return report, nil
}
