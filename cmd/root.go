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
	"github.com/spf13/cobra"
)

var JiraUrl string

var RootCmd = &cobra.Command{
	Use:   "patcher",
	Short: "Utility to help patch management for Jira",
	Long: `Patcher provides helper utility to download/apply/upload new patches to Jira according to the most
	common naming convention`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&JiraUrl, "jira", "https://issues.apache.org/jira/", "The JIRA instance to contact")
}
