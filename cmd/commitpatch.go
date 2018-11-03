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
	"fmt"
)

var OmitContributor bool

func commitPatch(issueKey string) {
	jiraClient := createJiraClient()
	options := jira.GetQueryOptions{}
	issue, _, err := jiraClient.Issue.Get(issueKey, &options)
	if err != nil {
		panic(err)
	}
	summary := issue.Fields.Summary
	if summary[len(summary)-1] != '.' {
		summary = summary + "."
	}

	contributedBy := ""
	if (!OmitContributor) {
		name := issue.Fields.Assignee.DisplayName
		if len(name) == 0 {
			name = issue.Fields.Assignee.Name
		}
		contributedBy = fmt.Sprintf(" Contributed by %s.", name)
	}

	commitMessage := fmt.Sprintf("%s. %s%s", issueKey, summary, contributedBy)

	executeAndPrint(fmt.Sprintf("git commit -m \"%s\"",commitMessage))

}

// applyCmd represents the apply command
var commitPatchCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit patch with the right commit message",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		commitPatch(issueFromArgsOrDetect(args))
	},
}

func init() {
	commitPatchCmd.Flags().BoolVarP(&OmitContributor, "omit-contributor", "C", false, "Omit 'Contributed by ...' from commit message")
	RootCmd.AddCommand(commitPatchCmd)
}
