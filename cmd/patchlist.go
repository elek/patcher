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
	"github.com/olekukonko/tablewriter"
	"os"
)

func patchlist() {
	//branch := "HDFS-7240"
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	workEnv, err := detectWorkEnv(dir)
	if err != nil {
		panic(err)
	}
	jiraClient := createJiraClient()
	options := jira.SearchOptions{}
	options.MaxResults = 50
	query := "status = 'Patch Available' ORDER BY UPDATED DESC"
	if len(workEnv.Query) > 0 {
		query = workEnv.Query + " AND " + query
	}
	issues, _, err := jiraClient.Issue.Search(query, &options)
	if err != nil {
		panic(err)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	for _, issue := range (issues) {
		table.Append([]string{issue.Key, issue.Fields.Summary, issue.Fields.Assignee.Name})
	}
	table.Render()

}

// applyCmd represents the apply command
var patchlistCmd = &cobra.Command{
	Use:   "list",
	Short: "List issues with patch available status",
	Run: func(cmd *cobra.Command, args []string) {
		patchlist()
	},
}

func init() {
	RootCmd.AddCommand(patchlistCmd)

}
