package repository

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/trace"
	//"log"
	"os"
	"path"
	"regexp"
	//"strings"
	"forjj/utils"
	"forjj/git"
)

func (i *GitRepoStruct) Path() string {
	return path.Clean(i.path)
}

// ensure local repo git exists and is initialized.
// - dir exist
// - repo initialized
// At the end Current dir is in the Repo.
func (i *GitRepoStruct) EnsureInitialized() error {
	if creatable := i.is_creatable(); !creatable {
		return i.err
	}

	if git.Do("init", i.path) > 0 {
		return fmt.Errorf("Unable to initialize %s", i.path)
	}

	if err := os.Chdir(i.path); err != nil {
		return fmt.Errorf("Unable to move repository at %s. %s", i.path, err)
	}

	return nil
}

func (i *GitRepoStruct) EnsureBranchConnected(branch, remote string) error {
	if err := i.use() ; err != nil {
		return fmt.Errorf("Unable to connect branches. %s", err)
	}
	if git.Do("branch", "--set-upstream=" + remote, branch) > 0 {
		return fmt.Errorf("Unable to set url '%s' to branch '%s'", remote, branch)
	}
	return nil
}

func (i *GitRepoStruct) CheckOut(branch string) error {
	if err := i.use() ; err != nil {
		return fmt.Errorf("Unable to connect branches. %s", err)
	}
	if git.Do("checkout", branch) > 0 {
		return fmt.Errorf("Unable to checkout to branch '%s'", branch)
	}
	return nil

}

// Ensure local repo exist with at least 1 commit.
// If non existent, or no commit, it will create it all.
// If upstream or remote is empty, upstream won't be tested and configured.
//
// It will take care of several scenarios related to commits on local and remote git cache.
// - no remote defined                         : Create 1st commit if needed.
// - nothing exists locally and remotely       : Create 1st commit and push
// - nothing exists locally but remotely       : clone/pull
// - something exist locally, but not remotely : push
// - both locally and remotely repo exist.     : Nothing done. No push.
//func (i *GitRepoStruct) ensure_local_repo_synced(repo_name, branch, remote, upstream, README_content string) error {
//	log.Printf("Updating your workspace with '%s(%s)'.", repo_name, upstream)
//	repo := i.path
//
//	if repo_name == "" {
//		return fmt.Errorf("Invalid Repo name. Repository name is empty.")
//	}
//
//	if branch == "" {
//		branch = "master" // Default branch name.
//	}
//
//	if found, err := git.Get("config", "--local", "-l"); err != nil {
//		return fmt.Errorf("'%s' is not a valid GIT repository. Please fix it first. %s\n", repo, err)
//	} else {
//		gotrace.Trace("Valid local git config found: \n%s", found)
//	}
//
//	// Ensure we are on local <branch>
//	if str, _ := git.Get("symbolic-ref", "--short", "HEAD"); strings.Trim(str, "\n") != branch {
//		// TODO: Replace this by a test on branch existence to avoid confusable red messages.
//		if git.Do("checkout", "-b", branch) > 0 {
//			git.Do("checkout", branch)
//		}
//	}
//
//	// Upstream to configure ???
//	if upstream == "" || remote == "" { // No upstream to configure
//		// Create initial commit
//		if _, err := git.Get("log", "-1", "--oneline"); err != nil {
//			git_1st_commit(repo, README_content)
//		} else {
//			gotrace.Trace("nothing done on non empty '%s' git repository...", repo)
//		}
//		return nil
//	}
//
//	// Configuring upstream and use it to pull or push.
//
//	if err := ensure_git_remote(upstream, remote); err != nil {
//		return err
//	}
//
//	remote_exist := git_1st_commit_exist(remote + "/" + branch)
//	local_exist := git_1st_commit_exist(branch)
//
//	// ensure local <branch> branch is connected to <remote>/<branch>
//	switch {
//	case local_exist && !remote_exist:
//		if git.Do("push", "-u", remote, branch) != 0 {
//			return fmt.Errorf("Unable to push to '%s'.", upstream)
//		}
//	case local_exist && remote_exist:
//		// Nothing to do. We do not push pending code. We let end user to do it, himself.
//		log.Printf("%s is properly configured. And, no push has been done. You may need to do it yourself to approve commits to be delivered to your DevOps team.", repo)
//	case remote_exist && !local_exist:
//		if git.Do("pull", remote, branch) != 0 {
//			return fmt.Errorf("Unable to pull from '%s'. Please fix the issue and retry.", upstream)
//		}
//		if git.Do("branch", branch, "--set-upstream-to", remote+"/"+branch) > 0 {
//			return fmt.Errorf("Unable to set git branch upstream to '%s'", upstream)
//		}
//	case !remote_exist && !local_exist:
//		git_1st_commit(repo, README_content)
//		if git.Do("push", "-u", remote, branch) != 0 {
//			return fmt.Errorf("Unable to push to '%s'.", upstream)
//		}
//	}
//	return nil
//}

// Ensure remote is properly configured
// If remote is missing. It will be created then fetched.
// if exists, check remote. If different, old is renamed to original, then created and fetched
func (i *GitRepoStruct) EnsureGitRemote(upstream, upstream_name string) error {
	if err := i.use() ; err != nil {
		return fmt.Errorf("Unable to ensure Git remote properly configured. %s", err)
	}

	origin_ok_regex, _ := regexp.Compile(upstream_name + "\t*" + upstream)
	origin_exist_regex, _ := regexp.Compile(upstream_name)

	ret, err := git.Get("remote", "-v")
	if err != nil {
		return fmt.Errorf("Issue to get git remote list. %s", err)
	}

	if origin_exist_regex.Match([]byte(ret)) {
		if !origin_ok_regex.Match([]byte(ret)) {
			if git.Do("remote", "rename", upstream_name, "original_"+upstream_name) != 0 {
				return fmt.Errorf("Unable to rename the '%s' remote to 'original_%s'.", upstream_name, upstream_name)
			}
			if git.Do("remote", "add", upstream_name, upstream) != 0 {
				return fmt.Errorf("Unable to create '%s' remote with '%s'", upstream_name, upstream)
			}
		}
	} else {
		if git.Do("remote", "add", upstream_name, upstream) != 0 {
			return fmt.Errorf("Unable to create '%s' remote with '%s'", upstream_name, upstream)
		}
	}
	if git.Do("fetch", upstream_name) != 0 {
		return fmt.Errorf("Unable to fetch '%s'.", upstream_name)
	}
	return nil
}

// GitRemoteExist test a repository, master connected to an upstream repo master branch.
func GitRemoteExist(branch, remote, upstream string) (exist, found bool, err error) {
	var out string

	out, err = git.Get("branch", "-vv")
	if err != nil {
		return false, false, fmt.Errorf("Issue to get git branch list. %s", err)
	}

	exist, err = regexp.MatchString(`[* ] `+branch+`.* \[`+remote+"/"+branch+`( .*)?\]`, out)
	gotrace.Trace("Branches: \n%sFound? %t", out, exist)
	if !exist {
		return
	}

	out, err = git.Get("remote", "-v")
	if err != nil {
		return false, false, fmt.Errorf("Issue to get git branch list. %s", err)
	}

	found, err = regexp.MatchString(`origin\s*`+upstream, out) // remote + " *" + upstream, out)
	gotrace.Trace("Remotes: \n%sFound? %t", out, found)
	return
}

// return true is at least one commit exists.
func (i *GitRepoStruct)git_1st_commit_exist(branch string) bool {
	if _, err := git.Get("log", branch, "-1", "--oneline"); err == nil {
		return true
	}
	return false
}

// Create initial commit
func (i *GitRepoStruct)git_1st_commit(initial_commit func()([]string, error)) (err error) {
	var files []string

	if files, err = initial_commit() ; err != nil {
		return
	} else {
		git.Add(files)
	}
	git.Do("commit", "-m", "Initial commit")

	gotrace.Trace("Initial commit created.")
	return nil
}

// This function check if the repo exist and state if it is create-able.
// If the path is a valid repo, create-able will be false and err = nil
// else err is won't be nil. But create-able is true only if the repo
// can be initialized.
func (i *GitRepoStruct) is_creatable() (creatable bool) {

	gotrace.Trace("Checking '%s' repository...", i.path)

	if dir, err := os.Stat(i.path); err != nil && os.IsNotExist(err) {
		creatable = true
		return
	} else {
		if !dir.IsDir() {
			i.err = fmt.Errorf("%s must be a directory.", i.path)
			return
		}
	}

	if dir, err := os.Stat(path.Join(i.path, ".git")); err == nil {
		if !dir.IsDir() {
			i.err = fmt.Errorf("%s is not a valid GIT repository. You may need to remove %s/.git", i.path, i.path)
			return
		}
	} else {
		creatable = true
		i.err = fmt.Errorf("%s is not a GIT reporitory.", i.path)
		return
	}

	if err := os.Chdir(i.path); err != nil {
		i.err = fmt.Errorf("Unable to move to '%s' : %s\n", i.path, err)
		return
	}

	if utils.RunCmd("git", "rev-parse", "--show-toplevel") > 0 {
		i.err = fmt.Errorf("%s is not a valid git repository work tree.", i.path)
		return
	}
	i.err = fmt.Errorf("'%s' already exist and can't be created a second time.\n", i.path)
	return
}

func (i *GitRepoStruct) is_valid() (valid bool) {
	gotrace.Trace("Checking '%s' repository...", i.path)

	if dir, err := os.Stat(i.path); err != nil && os.IsNotExist(err) {
		return
	} else {
		if !dir.IsDir() {
			i.err = fmt.Errorf("%s must be a directory.", i.path)
			return
		}
	}

	if dir, err := os.Stat(path.Join(i.path, ".git")); err == nil {
		if !dir.IsDir() {
			i.err = fmt.Errorf("%s is not a valid GIT repository. You may need to remove %s/.git", i.path, i.path)
			return
		}
	} else {
		i.err = fmt.Errorf("%s is not a GIT reporitory.", i.path)
		return
	}

	if err := os.Chdir(i.path); err != nil {
		i.err = fmt.Errorf("Unable to move to '%s' : %s\n", i.path, err)
		return
	}

	if utils.RunCmd("git", "rev-parse", "--show-toplevel") > 0 {
		i.err = fmt.Errorf("%s is not a valid git repository work tree.", i.path)
		return
	}
	valid = true
	return
}
