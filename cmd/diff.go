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
	"github.com/spf13/cobra"
	"os/exec"
	"fmt"
	"errors"
	"strings"
	"gopkg.in/src-d/go-git.v4"
)

func execute(command string, args ...string) {
	out, err := exec.Command(command, args...).CombinedOutput()
	println("Executed command: " + command + " " + strings.Join(args, " "))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s", out)
}
func diff(issueKey string) {

	repository, err := GetGitRepo()
	if err != nil {
		panic(err)
	}
	ref, err := repository.Head()
	if err != nil {
		panic(errors.New("Can't identify the current HEAD"))
	}
	if branchExists(repository, "patcher-test") {
		execute("git", "branch", "-D", "patcher-test")
	}
	execute("git", "checkout", "-b", "patcher-test", "HEAD^")
	apply(issueKey)
	execute("git", "add", ".")
	commitPatch(issueKey)
	execute("git", "checkout", ref.Name().Short())
	execute("git", "diff", "HEAD", "patcher-test")
}
func branchExists(repository *git.Repository, branchName string) bool{
	iter,err := repository.Branches()
	if err!=nil{
		panic(err)
	}
	for {
		branch,err := iter.Next()
		if err!=nil{
			return false
		}
		print(branch.Name().Short())
		if branch.Name().Short() == branchName {
			return true
		}
	}
	return false
}

// applyCmd represents the apply command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare the patch with the last commit of the current branch.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		diff(issueFromArgsOrDetect(args))
	},
}



func init() {
	RootCmd.AddCommand(diffCmd)
}
