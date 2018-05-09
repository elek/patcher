package cmd

import "github.com/andygrunwald/go-jira"
import (
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/getlantern/errors"
	"github.com/tcnksm/go-gitconfig"
	"gopkg.in/src-d/go-git.v4"
)

var gitRepo *git.Repository

type Branch string
type WorkEnv struct {
	Branch string
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
	return currentissue()
}

func currentissue() string {

	head, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		panic(err)
	}

	regexp, err := regexp.Compile("^[A-Z]+\\-[0-9]+")
	if err != nil {
		panic(err)
	}
	match := regexp.FindStringSubmatch(string(head))
	if len(match) > 0 {
		log.Printf("Detected issues: %s (from branch name convention)", match[0])
		return match[0]
	}
	panic("Issue is not defined and can't be determined.")
}

func basecommit() (string, error) {
	return "HEAD^", nil
	//head, err := exec.Command("git", "log", "-n", "20", "--pretty=tformat:'%h %d %s'").Output()
	//if err != nil {
	//	return string(head), err
	//}

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

func detectWorkEnv(dir string) (Branch, error) {
	repository, err := openGit(dir)
	if err != nil {
		return "", errors.New("Can't open git repository", err)
	}

	gitlog, err := repository.Log(&git.LogOptions{})
	if err != nil {
		return "", errors.New("Can't read git log", err)
	}
	var i = 0
	for {
		commit, err := gitlog.Next()
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
				if strings.HasPrefix(reference.Name().Short(), "apache") {
					branch := reference.Name().Short()[len("apache")+1:]
					log.Printf("Detected branch: %s (it has apache prefix in the history)", branch)
					return Branch(branch), nil
				}
			}
		}
		i++
		if i > 40 {
			break
		}

	}
	return "", errors.New("Can't find the base branch in your current history. You may need a rebase.")

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
	for _, history := range issue.Changelog.Histories {
		for _, item := range history.Items {
			if item.Field == "Attachment" && item.FromString == "" {
				id := item.To.(string)
				attachments = append(attachments, jira.Attachment{ID: id, Filename: item.ToString})
			}
		}
	}
	return attachments, nil
}
