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
	"regexp"
)

var gitRepo *git.Repository

type WorkEnv struct {
	Branch string
	Query  string
}

func GetGitRepo() (*git.Repository, error) {
	if gitRepo != nil {
		return gitRepo, nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return gitRepo, errors.New("Can't find current directory")
	}
	gitRepo, err := openGit(dir)
	return gitRepo, err
}

func issueFromArgsOrDetect(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	gitRepo, err := GetGitRepo();
	if err != nil {
		panic(err)
	}
	head, err := gitRepo.Head()
	if err != nil {
		panic(err)
	}
	head.Name().Short()
	regexp, err := regexp.Compile("^[A-Z]+\\-[0-9]+")
	if err != nil {
		panic(err)
	}
	match := regexp.FindStringSubmatch(head.Name().Short())
	if len(match) > 0 {
		println("Detected issues is " + match[0])
		return match[0]
	}
	panic("Issue is not defined and can't be determined")
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
func openGit(dir string) (*git.Repository, error) {
	gitDir, err := findGitDir(dir)
	if err != nil {
		return nil, errors.New("Can't find git repository", err)

	}
	return git.PlainOpen(gitDir)
}
func detectWorkEnv(dir string) (WorkEnv, error) {
	workEnv := WorkEnv{}
	repository, err := openGit(dir)
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
	for cdir := dir; cdir != "/"; cdir = filepath.Dir(cdir) {
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
