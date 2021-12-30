package cmd

import "github.com/andygrunwald/go-jira"
import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/getlantern/errors"
	"github.com/mitchellh/go-homedir"
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
		log.Printf("Detected issue: %s (from branch name convention)", match[0])
		return match[0]
	}
	panic("Issue is not defined and can't be determined.")
}

func createJiraClient() *jira.Client {
	var jiraClient *jira.Client
	var err error

	if TokenAuth {
		_, password := findJiraAuth()
		pat := &jira.PATAuthTransport{
			Token: password,
		}
		jiraClient, err = jira.NewClient(pat.Client(), JiraUrl)
	} else {
		user, password := findJiraAuth()
		jiraClient, err = jira.NewClient(nil, JiraUrl)
		if len(user) > 0 {
			jiraClient.Authentication.SetBasicAuth(user, password)
		}
	}

	if err != nil {
		panic(err)
	}
	return jiraClient
}

func findJiraAuth() (string, string) {
	user := ""
	password := ""
	var err error
	user, err = gitconfig.Global("patcher.jira.username")
	if err == nil {
		password, err = gitconfig.Global("patcher.jira.password")
		if err == nil {
			return user, password
		}
	}

	user, password = "", ""
	path := os.Getenv("NETRC")
	if path == "" {
		path, err = homedir.Expand("~/.netrc")
		if err == nil {
			n, err := netrc.ParseFile(path)
			if err == nil {
				u, err := url.Parse(JiraUrl)
				if err == nil {
					machine := n.FindMachine(u.Host)
					if machine != nil {
						return machine.Login, machine.Password
					}
				}
			}
		}
	}

	return "", ""
}

func openGit(dir string) (*git.Repository, error) {
	gitDir, err := findGitDir(dir)
	if err != nil {
		return nil, errors.New("Can't find git repository", err)

	}
	return git.PlainOpen(gitDir)
}

func executeAndReturn(command string) string {
	out, err := exec.Command("bash", "-c", command).CombinedOutput()
	if err != nil {
		fmt.Printf("'%s' is failed\n", command)
		fmt.Printf("%s\n", out)
		panic(err)
	}
	return string(out)
}

func executeAndPrint(command string) {
	log.Printf("%s", executeAndReturn(command))
}

func findBaseBranch(dir string, remotePrefix string) (Branch, error) {
	decorate := executeAndReturn("git log HEAD~40..HEAD --pretty=\"%D\"")
	scanner := bufio.NewScanner(strings.NewReader(decorate))
	for scanner.Scan() {
		line := scanner.Text()
		for _, element := range (strings.Split(line, ",")) {
			decoration := strings.TrimSpace(element)
			if strings.HasPrefix(decoration, remotePrefix) {
				branch := decoration[len(remotePrefix)+1:]
				log.Printf("Detected branch: %s (it has apache prefix in the history)", branch)
				return Branch(branch), nil
			}
		}

	}

	return "", fmt.Errorf("Can't find the base (a remote branch with prefix %s) "+
		"branch in your current history. You may need a rebase.", remotePrefix)

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
