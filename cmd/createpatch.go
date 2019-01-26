// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"github.com/andygrunwald/go-jira"
	"github.com/spf13/cobra"
	"log"
)

func upload(issueKey string, upload bool, branch Branch, baseref string) {
	jiraClient := createJiraClient()
	attachments, err := findAttachments(jiraClient, issueKey)
	if err != nil {
		panic(err)
	}

	maxId := findMaxId(attachments)
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if branch == "auto" {
		branch, err = findBaseBranch(dir, "apache")
		if err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
		if baseref == "HEAD^" {
			baseref = "apache/" + string(branch)
		}

	}
	patchName := createPatchFileName(maxId, issueKey, branch)
	fileName := "/tmp/" + patchName
	createPatch(fileName, baseref)
	println("Patch is created " + fileName + " and will be uploaded to " + issueKey)
	if upload {
		uploadPatch(jiraClient, issueKey, patchName, fileName)
	}

}
func uploadPatch(jiraClient *jira.Client, issueKey string, patchName string, fileName string) {
	options := jira.GetQueryOptions{}
	issue, _, err := jiraClient.Issue.Get(issueKey, &options)
	if err != nil {
		panic(err)
	}
	println("Uploading patch " + fileName + " to the issue " + issue.Key)
	file, error := os.Open(fileName)
	if error != nil {
		panic(error)
	}
	defer file.Close()
	_, response, error := jiraClient.Issue.PostAttachment(issueKey, file, patchName)
	if error != nil {
		panic(error)
	}
	if response.StatusCode > 400 {
		print("HTTP response status is " + response.Status)
	} else {
		println("Patch is uploaded successfully. See: https://issues.apache.org/jira/browse/" + issueKey)
	}
}
func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
func createPatch(fileName string, baseref string) {
	basecommit := baseref
	diffCommand := fmt.Sprintf("git diff --binary %s..HEAD > %s", basecommit, fileName)
	logCommandIncluded := fmt.Sprintf("git log --pretty=oneline %s..HEAD", basecommit)
	logCommandPre := fmt.Sprintf("git log --pretty=oneline %s~3..%s", basecommit, basecommit)

	log.Println("Included commits: ")
	executeAndPrint(logCommandIncluded)
	log.Println("Excluded commits: ")
	executeAndPrint(logCommandPre)
	log.Print(diffCommand)
	out, err := exec.Command("bash", "-c", diffCommand).CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
		panic(err)
	}
	fmt.Printf("%s", out)
}


func createPatchFileName(maxId int, issueKey string, branch Branch) string {
	if len(branch) > 0 && branch != "master" && branch != "trunk" {
		return fmt.Sprintf("%s-%s.%03d.patch", issueKey, branch, maxId+1)
	} else {
		return fmt.Sprintf("%s.%03d.patch", issueKey, maxId+1)
	}
}
func findMaxId(attachments []jira.Attachment) int {
	maxId := 0
	re := regexp.MustCompile(".+\\.0*(\\d+)\\.patch")
	for _, attachment := range attachments {
		submatch := re.FindStringSubmatch(attachment.Filename)
		if submatch != nil {
			index, err := strconv.Atoi(submatch[1])
			if err == nil {
				maxId = max(index, maxId)
			}
		}
	}
	return maxId
}

func init() {
	var doUpload bool = false
	var branch string
	var base string
	// applyCmd represents the apply command
	createPatchCmd := &cobra.Command{
		Use:   "create",
		Short: "Create patch and upload it to the jira.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			upload(issueFromArgsOrDetect(args), doUpload, Branch(branch), base)
		},
	}
	createPatchCmd.Flags().BoolVarP(&doUpload, "upload", "u", false, "Upload to the jira after issue creation")
	createPatchCmd.Flags().StringVar(&branch, "branch", "auto", "Define "+"the working branch (auto means autodetect)")
	createPatchCmd.Flags().StringVar(&base, "base", "HEAD^", "Define the "+
		"base commit which should be used to create the diff (git diff base."+
		".HEAD")
	RootCmd.AddCommand(createPatchCmd)

}
