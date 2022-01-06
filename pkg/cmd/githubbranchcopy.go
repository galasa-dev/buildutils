//
// Copyright contributors to the Galasa project
//

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"galasa.dev/buildUtilities/pkg/githubjson"
)

var (
    githubBranchCopyCmd = &cobra.Command{
		Use:   "copy",
		Short: "Copy a branch to a new one",
		Long:  "Copy an existing branch to a new one,  without having to clone and push locally",
		Run:   githubBranchCopyExecute,
	}

	branchCopyFromBranch   string
	branchCopyFromTag      string
	branchCopyTo           string
)

func init() {
	githubBranchCopyCmd.PersistentFlags().StringVarP(&branchCopyFromBranch, "branch", "", "", "from branch")
	githubBranchCopyCmd.PersistentFlags().StringVarP(&branchCopyFromTag, "tag", "", "", "from branch")
	githubBranchCopyCmd.PersistentFlags().StringVarP(&branchCopyTo, "to", "", "", "to branch")

	githubBranchCopyCmd.MarkPersistentFlagRequired("to")

	githubBranchCmd.AddCommand(githubBranchCopyCmd)
}

func githubBranchCopyExecute(cmd *cobra.Command, args []string) {

	if branchCopyFromBranch == "" && branchCopyFromTag == "" {
		branchCopyFromBranch = "main"
	}

	basicAuth, err := githubGetBasicAuth()
    if err != nil {
        panic(err)
    }

	// First get the sha of the from branch

	var url = ""
	if branchCopyFromBranch != "" {
		url = fmt.Sprintf("https://api.github.com/repos/galasa-dev/%v/git/ref/heads/%v", githubRepository, branchCopyFromBranch)
	} else {
		url = fmt.Sprintf("https://api.github.com/repos/galasa-dev/%v/git/ref/tags/%v", githubRepository, branchCopyFromTag)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(nil)
	}
    
    req.Header.Set("Authorization", basicAuth)

    client := &http.Client{}
    resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Get from sha failed for url %v - status line - %v\n", url, resp.Status);
		os.Exit(1)
	}

	defer resp.Body.Close()
	var reference githubjson.Reference

	err = json.NewDecoder(resp.Body).Decode(&reference)

	if err != nil {
		panic(err)
	}

	if branchCopyFromBranch != "" {
		fmt.Printf("SHA for branch %v is %v\n", branchCopyFromBranch, reference.Object.Sha)
	} else {
		url = fmt.Sprintf("https://api.github.com/repos/galasa-dev/%v/git/tags/%v", githubRepository, reference.Object.Sha)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(nil)
		}
		
		req.Header.Set("Authorization", basicAuth)
	
		client := &http.Client{}
		respTag, err := client.Do(req)
		if err != nil {
			panic(err)
		}
	
		if respTag.StatusCode != http.StatusOK {
			fmt.Printf("Get from sha failed for url %v - status line - %v\n", url, respTag.Status);
			os.Exit(1)
		}
	
		defer respTag.Body.Close()

		err = json.NewDecoder(respTag.Body).Decode(&reference)

		if err != nil {
			panic(err)
		}

		fmt.Printf("SHA for tag %v is %v\n", branchCopyFromTag, reference.Object.Sha)
	}

	// Now create the new branch based on that sha

	url = fmt.Sprintf("https://api.github.com/repos/galasa-dev/%v/git/refs", githubRepository)

	var newReference githubjson.NewReference
	newReference.Ref = fmt.Sprintf("refs/heads/%v", branchCopyTo)
	newReference.Sha = reference.Object.Sha

	newReferenceBuffer := new(bytes.Buffer)
	err = json.NewEncoder(newReferenceBuffer).Encode(newReference)
	if err != nil {
		panic(err)
	}

	req, err = http.NewRequest("POST", url, newReferenceBuffer)
	if err != nil {
		panic(nil)
	}
    
    req.Header.Set("Authorization", basicAuth)
	req.Header.Set("Content-Type", "application/json")

    respNew, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer respNew.Body.Close()

    if respNew.StatusCode != http.StatusCreated {
    	fmt.Printf("Get from sha failed for url %v - status line - %v\n", url, respNew.Status);
        os.Exit(1)
	}

    fmt.Printf("Branch %v created on repository %v at sha %v\n", branchCopyTo, githubRepository,reference.Object.Sha)

}