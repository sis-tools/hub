package commands

import (
	"fmt"
	"regexp"

	"github.com/github/hub/github"
	"github.com/github/hub/utils"
)

var cmdCheckout = &Command{
	Run:          checkout,
	GitExtension: true,
	Usage:        "checkout <PULLREQ-URL> [<BRANCH>]",
	Long: `Check out the head of a pull request as a local branch.

## Examples:
		$ hub checkout https://github.com/jingweno/gh/pull/73
		> git remote add -f --no-tags -t feature git://github:com/jingweno/gh.git
		> git checkout --track -B jingweno-feature jingweno/feature

## See also:

hub-merge(1), hub-am(1), hub(1), git-checkout(1)
`,
}

func init() {
	CmdRunner.Use(cmdCheckout)
}

func checkout(command *Command, args *Args) {
	if !args.IsParamsEmpty() {
		err := transformCheckoutArgs(args)
		utils.Check(err)
	}
}

func transformCheckoutArgs(args *Args) error {
	words := args.Words()

	if len(words) == 0 {
		return nil
	}

	checkoutURL := words[0]
	var newBranchName string
	if len(words) > 1 {
		newBranchName = words[1]
	}

	url, err := github.ParseURL(checkoutURL)
	if err != nil {
		// not a valid GitHub URL
		return nil
	}

	pullURLRegex := regexp.MustCompile("^pull/(\\d+)")
	projectPath := url.ProjectPath()
	if !pullURLRegex.MatchString(projectPath) {
		// not a valid PR URL
		return nil
	}

	err = sanitizeCheckoutFlags(args)
	if err != nil {
		return err
	}

	id := pullURLRegex.FindStringSubmatch(projectPath)[1]
	gh := github.NewClient(url.Project.Host)
	pullRequest, err := gh.PullRequest(url.Project, id)
	if err != nil {
		return err
	}

	if idx := args.IndexOfParam(newBranchName); idx >= 0 {
		args.RemoveParam(idx)
	}

	branch := pullRequest.Head.Ref
	headRepo := pullRequest.Head.Repo
	if headRepo == nil {
		return fmt.Errorf("Error: that fork is not available anymore")
	}
	user := headRepo.Owner.Login

	if newBranchName == "" {
		newBranchName = fmt.Sprintf("%s-%s", user, branch)
	}

	repo, err := github.LocalRepo()
	utils.Check(err)

	_, err = repo.RemoteByName(user)
	if err == nil {
		args.Before("git", "remote", "set-branches", "--add", user, branch)
		remoteURL := fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", branch, user, branch)
		args.Before("git", "fetch", user, remoteURL)
	} else {
		u := url.Project.GitURL(headRepo.Name, user, headRepo.Private)
		args.Before("git", "remote", "add", "-f", "--no-tags", "-t", branch, user, u)
	}

	remoteName := fmt.Sprintf("%s/%s", user, branch)
	replaceCheckoutParam(args, checkoutURL, newBranchName, remoteName)

	return nil
}

func sanitizeCheckoutFlags(args *Args) error {
	if i := args.IndexOfParam("-b"); i != -1 {
		return fmt.Errorf("Unsupported flag -b when checking out pull request")
	}

	if i := args.IndexOfParam("--orphan"); i != -1 {
		return fmt.Errorf("Unsupported flag --orphan when checking out pull request")
	}

	return nil
}

func replaceCheckoutParam(args *Args, checkoutURL, branchName, remoteName string) {
	idx := args.IndexOfParam(checkoutURL)
	args.RemoveParam(idx)
	args.InsertParam(idx, "--track", "-B", branchName, remoteName)
}
