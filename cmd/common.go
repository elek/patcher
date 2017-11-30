package cmd

import "github.com/andygrunwald/go-jira"
import (
	"gopkg.in/src-d/go-git.v4"
	"github.com/tcnksm/go-gitconfig"
	"strings"
	"github.com/getlantern/errors"
	"path/filepath"
	"os"
	"path"
)

type WorkEnv struct {
	Branch string
	Query  string
}

func createJiraClient() *jira.Client {
	url := "https://issues.apache.org/jira/"
	user, err := gitconfig.Global("patcher.jira.username")
	if err != nil {
		user = ""
	}
	password, err := gitconfig.Global("patcher.jira.password")
	if err != nil {
		password = ""
	}

	jiraClient, err := jira.NewClient(nil, url)
	if len(user) > 0 {
		jiraClient.Authentication.SetBasicAuth(user, password)
	}
	if err != nil {
		panic(err)
	}
	return jiraClient

}

func detectWorkEnv(dir string) (WorkEnv, error) {
	workEnv := WorkEnv{}
	gitDir, err := findGitDir(dir)
	if err != nil {
		return workEnv, errors.New("Can't find git repository", err)

	}
	repository, err := git.PlainOpen(gitDir)
	if err != nil {
		return workEnv, errors.New("Can't open git repository", err)
	}
	config, err := repository.Config()
	if err != nil {
		return workEnv, errors.New("Can't read git config of the git repository.", err)
	}
	for _, option := range (config.Raw.Section("patcher").Options) {

		if err != nil {
			panic(err)
		}

		log, err := repository.Log(&git.LogOptions{})
		if err != nil {
			panic(err)
		}

		i := 0

		for {
			commit, err := log.Next()
			if err != nil {
				break
			}

			references, err := repository.References()
			for {
				reference, err := references.Next()
				if err != nil {
					break
				}
				if reference.Hash().String() == commit.Hash.String() {
					if strings.HasSuffix(reference.Name().Short(), "/"+option.Key) {
						return WorkEnv{
							Branch: option.Key,
							Query:  option.Value,
						}, nil
					}
				}
			}
			i++
			if i > 20 {
				break
			}
		}
	}
	return workEnv, errors.New("No matching branch definition in .git/config file\nDefine branch with\n[patcher]\nbranch_name=jql_query")

}
func findGitDir(dir string) (string, error) {
	for cdir := dir; cdir != "/"; cdir = filepath.Dir(dir) {
		if _, err := os.Stat(path.Join(cdir, ".git")); !os.IsNotExist(err) {
			return cdir, nil
		}
	}
	return "", errors.New("Can't find .git in any parent directory.")

}
func findAttachments(jiraClient *jira.Client, issueKey string) ([]jira.Attachment, error) {
	options := jira.GetQueryOptions{}
	options.Expand = "changelog"

	issue, _, err := jiraClient.Issue.Get(issueKey, &options)
	if err != nil {
		return nil, err
	}
	var attachments []jira.Attachment
	for _, history := range (issue.Changelog.Histories) {
		for _, item := range (history.Items) {
			if item.Field == "Attachment" && item.FromString == "" {
				id := item.To.(string)
				attachments = append(attachments, jira.Attachment{ID: id, Filename: item.ToString})
			}
		}
	}
	return attachments, nil
}
