package cmd

import "github.com/andygrunwald/go-jira"
import (
	"gopkg.in/src-d/go-git.v4"
	"github.com/tcnksm/go-gitconfig"
	"strings"
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
func detectWorkEnv(dir string) WorkEnv {
	repository, err := git.PlainOpen(dir)
	if err != nil {
		panic(err)
	}
	config, err := repository.Config()
	if err != nil {
		panic(err)
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
						}
					}
				}
			}
			i++
			if i > 20 {
				return WorkEnv{}
			}
		}
	}
	return WorkEnv{}

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
