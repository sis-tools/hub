package github

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/github/hub/utils"
)

const (
	pullRequestTemplate = "pull_request_template"
	issueTemplate       = "issue_template"
	githubTemplateDir   = ".github"
)

func GetPullRequestTemplate() string {
	return getGithubTemplate(pullRequestTemplate)
}

func GetIssueTemplate() string {
	return getGithubTemplate(issueTemplate)
}

func getGithubTemplate(pat string) (body string) {
	var path string

	if _, err := os.Stat(githubTemplateDir); err == nil {
		if p := getFilePath(githubTemplateDir, pat); p != "" {
			path = p
		}
	}

	if path == "" {
		if p := getFilePath(".", pat); p != "" {
			path = p
		}
	}

	if path == "" {
		return
	}

	body, err := readContentsFromFile(path)
	utils.Check(err)
	return
}

func getFilePath(dir, pattern string) string {
	files, err := ioutil.ReadDir(dir)
	utils.Check(err)

	for _, file := range files {
		fileName := file.Name()
		path := fileName

		if ext := filepath.Ext(fileName); ext == ".md" {
			path = strings.TrimRight(fileName, ".md")
		} else if ext == ".txt" {
			path = strings.TrimRight(fileName, ".txt")
		}

		path = strings.ToLower(path)

		if ok, _ := filepath.Match(pattern, path); ok {
			return filepath.Join(dir, fileName)
		}
	}
	return ""
}

func readContentsFromFile(filename string) (contents string, err error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	contents = strings.Replace(string(content), "\r\n", "\n", -1)
	contents = strings.TrimSpace(contents)
	return
}
