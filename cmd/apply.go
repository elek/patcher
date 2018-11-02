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
	"github.com/andygrunwald/go-jira"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
	"fmt"
)

func downloadFile(client *jira.Client, attachment jira.Attachment, filepath string) (err error) {

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := client.Issue.DownloadAttachment(attachment.ID)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func apply(issueKey string) {
	jiraClient := createJiraClient()
	attachments, err := findAttachments(jiraClient, issueKey)
	if err != nil {
		panic(err)
	}
	if (len(attachments) == 0) {
		println(fmt.Sprintf("No attachments for %s", issueKey))
		os.Exit(1)
	}
	lastPatch := attachments[len(attachments)-1]
	println("Downloading " + lastPatch.Filename)

	patchFile := "/tmp/" + lastPatch.Filename
	downloadFile(jiraClient, lastPatch, patchFile)
	applyFile(patchFile)

}
func applyFile(fileName string) {
	out, err := exec.Command("git", "apply", fileName).CombinedOutput()
	if err != nil {
		println(err.Error())
	}
	fmt.Printf("%s", out)
}

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Download patch from jira and apply",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apply(issueFromArgsOrDetect(args))
	},
}

func init() {
	RootCmd.AddCommand(applyCmd)
}
