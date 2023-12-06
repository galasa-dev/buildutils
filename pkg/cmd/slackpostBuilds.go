/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	slackpostBuildsCmd = &cobra.Command{
		Use:   "tests",
		Short: "Make a Slack post of the failing tests from a galasactl runs submit report",
		Long:  "",
		Run:   slackpostTestsExecute,
	}
	repo            string
	pipelineRunName string
)

func init() {
	slackpostBuildsCmd.PersistentFlags().StringVar(&repo, "repo", "", "The repo the build failed for")
	slackpostBuildsCmd.PersistentFlags().StringVar(&pipelineRunName, "prun", "", "The name of the PipelineRun that failed")

	slackpostBuildsCmd.MarkPersistentFlagRequired("repo")
	slackpostBuildsCmd.MarkPersistentFlagRequired("prun")

	slackpostCmd.AddCommand(slackpostBuildsCmd)
}

func slackpostBuildsExecute(cmd *cobra.Command, args []string) {
	fmt.Printf("Galasa Build - Slack Failed Build Report - version %v\n", rootCmd.Version)

	linkToPipelineRun := fmt.Sprintf("http://localhost:8001/api/v1/namespaces/tekton-pipelines/services/tekton-dashboard:http/proxy/#/namespaces/galasa-build/pipelineruns/%s", pipelineRunName)
	content := fmt.Sprintf("Galasa Build Failed for the %s repository.\n\nPlease see %s for details.", repo, linkToPipelineRun)

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
