package main

import (
    "fmt"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "os"
    "path"
    "io/ioutil"
    "path/filepath"
    "regexp"
    "strings"
    "github.com/alecthomas/kingpin"
    "log"
)

func (a *Forj) RepoPath(repo_name string) string {
    return path.Clean(path.Join(a.Workspace_path, a.Workspace, repo_name))
}

// ensure local repo git exists and is initialized.
// - dir exist
// - repo initialized
func (a *Forj) ensure_local_repo_initialized(repo_name string) error {
    repo := a.RepoPath(repo_name)

    gotrace.Trace("Checking '%s' repository...", repo)
    dir, err := os.Stat(repo)
    if os.IsNotExist(err) {
        if git("init", repo) > 0 {
            return fmt.Errorf("Unable to initialize %s\n", repo)
        }
        gotrace.Trace("Created '%s' repository...", repo)
    }

    dir, err = os.Stat(path.Join(repo, ".git"))
    if os.IsNotExist(err) {
        gotrace.Trace("Existing directory '%s' will became a git repo", repo)
        if git("init", repo) > 0 {
            return fmt.Errorf("Unable to initialize %s\n", repo)
        }
        gotrace.Trace("Initialized '%s' directory as git repository...", repo)
    }

    if os.IsExist(err) && !dir.IsDir() {
        return fmt.Errorf("'%s' is not a valid GIT repository. Please fix it first. '%s' is not a directory.\n", repo, path.Join(repo, ".git"))
    }

    if err := os.Chdir(repo) ; err != nil {
        return fmt.Errorf("Unable to move to '%s' : %s\n", err)
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
// - both locally and remotely repo exist.     : Just test remote connection.
func (a *Forj) ensure_local_repo_synced(repo_name, branch, remote, upstream, README_content string) error {
    repo := path.Clean(path.Join(a.Workspace_path, a.Workspace, repo_name))

    if repo_name == "" {
        return fmt.Errorf("Invalid Repo name. Repository name is empty.")
    }

    if branch == "" {
        branch = "master" // Default branch name.
    }

    if found, err := git_get("config", "--local", "-l"); err != nil {
        return fmt.Errorf("'%s' is not a valid GIT repository. Please fix it first. %s\n", repo, err)
    } else {
        gotrace.Trace("Valid local git config found: \n%s", found)
    }

    // Ensure we are on local <branch>
    if str, _ := git_get("symbolic-ref", "--short", "HEAD"); strings.Trim(str, "\n") != branch {
        // TODO: Replace this by a test on branch existence to avoid confusable red messages.
        if git("checkout", "-b", branch) >0 {
            git("checkout", branch)
        }
    }

    // Upstream to configure ???
    if upstream == "" || remote == "" { // No upstream to configure
        // Create initial commit
        if _, err := git_get("log", "-1", "--oneline"); err != nil {
            git_1st_commit(repo, README_content)
        } else {
            gotrace.Trace("nothing done on non empty '%s' git repository...", repo)
        }
        return nil
    }

    // Configuring upstream and use it to pull or push.

    if err := ensure_git_remote(upstream, remote) ; err != nil {
        return err
    }

    remote_exist := git_1st_commit_exist(remote + "/" + branch)
    local_exist := git_1st_commit_exist(branch)

    // ensure local <branch> branch is connected to <remote>/<branch>
    switch {
    case local_exist && !remote_exist:
        if git("push", "-u", remote, branch) != 0 {
            return fmt.Errorf("Unable to push to '%s'.", upstream)
        }
    case local_exist && remote_exist:
        // Nothing to do. We do not push pending code. We let end user to do it, himself.
        log.Printf("%s is properly configured. And, no push has been done. You may need to do it yourself to approve commits to be delivered to your DevOps team.", repo)
    case remote_exist && !local_exist :
        if git("pull", remote, branch) != 0 {
                return fmt.Errorf("Unable to pull from '%s'. Please fix the issue and retry.", upstream)
            }
        if git("branch", branch, "--set-upstream-to", remote + "/" + branch)>0 {
            return fmt.Errorf("Unable to set git branch upstream to '%s'", upstream)
        }
    case !remote_exist && !local_exist :
        git_1st_commit(repo, README_content)
        if git("push", "-u", remote, branch) != 0 {
            return fmt.Errorf("Unable to push to '%s'.", upstream)
        }
    }
    return nil
}

// Ensure remote is properly configured
// If remote is missing. It will be created then fetched.
// if exists, check remote. If different, old is renamed to original, then created and fetched
func ensure_git_remote(upstream, upstream_name string) error {
    origin_ok_regex, _ := regexp.Compile(upstream_name + "\t*" + upstream)
    origin_exist_regex, _ := regexp.Compile(upstream_name)

    ret, err := git_get("remote", "-v")
    if err != nil {
        return fmt.Errorf("Issue to get git remote list. %s", err)
    }

    if origin_exist_regex.Match([]byte(ret)) {
        if ! origin_ok_regex.Match([]byte(ret)) {
            if git("remote", "rename", upstream_name, "original_" + upstream_name) != 0 {
                return fmt.Errorf("Unable to rename the '%s' remote to 'original_%s'.", upstream_name, upstream_name)
            }
            if git("remote", "add", upstream_name, upstream) != 0 {
                return fmt.Errorf("Unable to create '%s' remote with '%s'", upstream_name, upstream)
            }
        }
    } else {
        if git("remote", "add", upstream_name, upstream) != 0 {
            return fmt.Errorf("Unable to create '%s' remote with '%s'", upstream_name, upstream)
        }
    }
    if git("fetch", upstream_name) != 0 {
        return fmt.Errorf("Unable to fetch '%s'.", upstream_name)
    }
    return nil
}

// test a repository, master connected to an upstream repo master branch.
func git_remote_exist(branch, remote, upstream string) (exist, found bool, err error) {
    var out string

    out, err = git_get("branch", "-vv")
    if err != nil {
        return false, false, fmt.Errorf("Issue to get git branch list. %s", err)
    }

    exist, err = regexp.MatchString(`[* ] ` + branch + `.* \[` + remote + "/" + branch + `( .*)?\]`, out)
    if !exist {
        return
    }

    out, err = git_get("remote", "-v")
    if err != nil {
        return false, false, fmt.Errorf("Issue to get git branch list. %s", err)
    }

    found, err = regexp.MatchString(`origin\s*` + upstream, out) // remote + " *" + upstream, out)
    return
}

// return true is at least one commit exists.
func git_1st_commit_exist(branch string) bool {
    if _, err := git_get("log", branch, "-1", "--oneline"); err == nil {
        return true
    }
    return false
}

// Commit All Plugins generated files and plugin options updates.
func (a *Forj) gitCommitAll(commit_msg string) error {

    for _, driver := range a.drivers {
        err := driver.gitAddPluginFiles()
        if err != nil {
            return fmt.Errorf("Issue to add driver '%s' generated files. %s.", driver.Name, err)
        }
        commit_msg += "\n - " + driver.plugin.Result.Data.CommitMessage
    }

    if err := a.SaveForjjPluginsOptions() ; err != nil {
        return err
    }

    git("commit", "-m", commit_msg)
    return nil
}

// Commit a Plugin generated files.
func (d *Driver) gitCommit() error {
    if git("commit", "-m", d.plugin.Result.Data.CommitMessage) > 0 {
        return fmt.Errorf("Unable to commit.")
    }
    return nil
}

// Add Plugins generated files to ready to be commit git space.
func (d *Driver)gitAddPluginFiles() (error) {
    if d.plugin.Result == nil {
        return fmt.Errorf("Strange... The plugin as no result (plugin.Result is nil). Did the plugin '%s' executed?", d.Name)
    }

    gotrace.Trace("Adding %d files related to '%s'", len(d.plugin.Result.Data.Files), d.plugin.Result.Data.CommitMessage)
    if len(d.plugin.Result.Data.Files) == 0 {
        return fmt.Errorf("Nothing to commit")
    }

    if d.plugin.Result.Data.CommitMessage == "" {
        return fmt.Errorf("Unable to commit without a commit message.")
    }

    for _, file := range d.plugin.Result.Data.Files {
        if i := git("add", path.Join("apps", d.DriverType, file)); i >0 {
            return fmt.Errorf("Issue while adding code to git. RC=%d", i)
        }
    }
    return nil
}

// Create initial commit
func git_1st_commit(repo, README_content string) {
    readme_path := path.Join(repo, "README.md")

    // check if an existing README exist to keep
    _, err := os.Stat(readme_path)
    if ! os.IsNotExist(err) {
        gotrace.Trace("Renaming '%s' to '%s'", readme_path, readme_path + ".forjj_tmp")
        if err := os.Rename(readme_path, readme_path + ".forjj_tmp") ; err!= nil {
            kingpin.Fatalf("Unable to rename '%s' to '%s'. %s\n", readme_path, readme_path + ".forjj_tmp", err)
            os.Exit(1)
        }
    }

    // Generate README.md
    // TODO: Support for a template data instead.
    gotrace.Trace("Generate %s", readme_path)
    data := []byte(fmt.Sprintf("FYI: This Repository has been created by forjj\n\n%s %s\n", filepath.Base(repo), README_content))
    if err := ioutil.WriteFile(readme_path, data, 0644) ; err!= nil {
        kingpin.Fatalf("Unable to create '%s'. %s\n", readme_path, err)
        os.Exit(1)
    }

    git("add", "README.md")
    git("commit", "-m", "Initial commit")

    // check if an original README.md was there to restore his content.
    _, err = os.Stat(readme_path + ".forjj_tmp")
    if ! os.IsNotExist(err) {
        gotrace.Trace("Renaming '%s' to '%s'", readme_path + ".forjj_tmp", readme_path)
        if err := os.Rename(readme_path + ".forjj_tmp", readme_path) ; err!= nil {
            kingpin.Fatalf("Unable to rename '%s' to '%s'. %s\n", readme_path + ".forjj_tmp", readme_path, err)
            os.Exit(1)
        }
    }
    gotrace.Trace("Initial commit created.")
}

