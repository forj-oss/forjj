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
)

// Ensure local repo exist with at least 1 commit.
// If non existent, or no commit, it will create it all.
func (a *Forj) ensure_local_repo(repo_name, upstream, README_content string) error {
    repo := path.Clean(path.Join(a.Workspace_path, a.Workspace, repo_name))

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

    // if _, err := git_get("config", "--get", "remote.origin.url"); err != nil {
    if err := os.Chdir(repo) ; err != nil {
        fmt.Printf("Unable to move to '%s' : %s\n", err)
        os.Exit(1)
    }
    if found, err := git_get("config", "--local", "-l"); err != nil {
        return fmt.Errorf("'%s' is not a valid GIT repository. Please fix it first. %s\n", repo, err)
    } else {
        gotrace.Trace("Valid local git config found: \n%s", found)
    }

    if upstream == "" {
        if _, err := git_get("log", "-1", "--oneline"); err != nil {
            git_1st_commit(repo, README_content)
            gotrace.Trace("Initial commit created.")
        } else {
            gotrace.Trace("nothing done on existing '%s' git repository...", repo)
        }
        return nil
    } else {
        if err := git_ensure_remote(upstream) ; err != nil {
            return err
        }

        // Ensure we are on local master branch
        if str, _ := git_get("symbolic-ref", "--short", "HEAD"); strings.Trim(str, "\n") != "master" {
            if git("checkout", "-b", "master") >0 {
                git("checkout", "master")
            }
        }

        var remote_exist, local_exist bool

        if _, err := git_get("log", "origin/master", "-1", "--oneline") ; err == nil {
            remote_exist = true
        }
        if _, err := git_get("log", "-1", "--oneline"); err == nil {
            local_exist = true
        }

        // ensure local master branch is connected to origin/master
        switch {
        case local_exist:
            // TODO: Replace following git sequences to avoid unwanted errors that end user will likely need to ignore.
            if git("branch", "master", "--set-upstream-to", "origin/master")>0 {
                if git("push", "-u", "origin", "master") != 0 {
                    return fmt.Errorf("Unable to push to '%s'.", upstream)
                }
            } else {
                if git("push") != 0 {
                    return fmt.Errorf("Unable to push to '%s'.", upstream)
                }
            }
        case remote_exist && !local_exist:
            if git("pull") != 0 {
                    return fmt.Errorf("Unable to pull from '%s'. Please fix the issue and retry.", upstream)
                }
        case !remote_exist && !local_exist :
            git_1st_commit(repo, README_content)
            if git("push", "-u", "origin", "master") != 0 {
                return fmt.Errorf("Unable to push to '%s'.", upstream)
            }
        }
        return nil
    }
}

// Ensure remote is properly configured
// If remote is missing. It will be created then fetched.
// if exists, check remote. If different, old is renamed to original, then created and fetched
func git_ensure_remote(upstream string) error {
    origin_ok_regex, _ := regexp.Compile("^origin *" + upstream)
    origin_exist_regex, _ := regexp.Compile("^origin")

    ret, err := git_get("remote", "-v")
    if err != nil {
        return fmt.Errorf("Issue to get git remote list. %s", err)
    }

    if origin_exist_regex.Match([]byte(ret)) {
        if ! origin_ok_regex.Match([]byte(ret)) {
            if git("remote", "rename", "origin", "original") != 0 {
                return fmt.Errorf("Unable to rename the 'origin' remote.")
            }
            if git("remote", "add", "origin", upstream) != 0 {
                return fmt.Errorf("Unable to create 'origin' remote with '%s'", upstream)
            }
        }
    } else {
        if git("remote", "add", "origin", upstream) != 0 {
            return fmt.Errorf("Unable to create 'origin' remote with '%s'", upstream)
        }
    }
    if git("fetch", "origin") != 0 {
        return fmt.Errorf("Unable to fetch origin.")
    }
    return nil
}

// Commit Plugins generated files and plugin options updates.
func (a *Forj) DoCommitAll(commit_msg string) error {

    for _, driver := range a.drivers {
        plugin_msg, err := driver.gitAddPluginFiles()
        if err != nil {
            return fmt.Errorf("Issue to add driver '%s' generated files. %s.", driver.name, err)
        }
        commit_msg += "\n - " + plugin_msg
    }

    if err := a.SaveForjjPluginsOptions() ; err != nil {
        return err
    }

    git("commit", "-m", commit_msg)
    return nil
}

// Add Plugins generated files to ready to be commit git space.
func (d *Driver)gitAddPluginFiles() (string, error) {
    if d.plugin.Result == nil {
        return "", fmt.Errorf("Strange... The plugin as no result (plugin.Result is nil). Did the plugin '%s' executed?", d.name)
    }

    gotrace.Trace("Adding %d files related to '%s'", len(d.plugin.Result.Data.Files), d.plugin.Result.Data.CommitMessage)
    if len(d.plugin.Result.Data.Files) == 0 {
        return "", fmt.Errorf("Nothing to commit")
    }

    if d.plugin.Result.Data.CommitMessage == "" {
        return "", fmt.Errorf("Unable to commit without a commit message.")
    }

    for _, file := range d.plugin.Result.Data.Files {
        if i := git("add", path.Join("apps", d.driver_type, file)); i >0 {
            return "", fmt.Errorf("Issue while adding code to git. RC=%d", i)
        }
    }
    return d.plugin.Result.Data.CommitMessage, nil
}


// Create initial commit
func git_1st_commit(repo, README_content string) {
    readme_path := path.Join(repo, "README.md")

    // check if an existing README exist to keep
    _, err := os.Stat(readme_path)
    if ! os.IsNotExist(err) {
        gotrace.Trace("Renaming '%s' to '%s'", readme_path, readme_path + ".forjj_tmp")
        if err := os.Rename(readme_path, readme_path + ".forjj_tmp") ; err!= nil {
            fmt.Printf("Unable to rename '%s' to '%s'. %s\n", readme_path, readme_path + ".forjj_tmp", err)
            os.Exit(1)
        }
    }

    // Generate README.md
    // TODO: Support for a template data instead.
    gotrace.Trace("Generate %s", readme_path)
    data := []byte(fmt.Sprintf("FYI: This Repository has been created by forjj\n\n%s %s\n", filepath.Base(repo), README_content))
    if err := ioutil.WriteFile(readme_path, data, 0644) ; err!= nil {
        fmt.Printf("Unable to create '%s'. %s\n", readme_path, err)
        os.Exit(1)
    }

    git("add", "README.md")
    git("commit", "-m", "Initial commit")

    // check if an original README.md was there to restore his content.
    _, err = os.Stat(readme_path + ".forjj_tmp")
    if ! os.IsNotExist(err) {
        gotrace.Trace("Renaming '%s' to '%s'", readme_path + ".forjj_tmp", readme_path)
        if err := os.Rename(readme_path + ".forjj_tmp", readme_path) ; err!= nil {
            fmt.Printf("Unable to rename '%s' to '%s'. %s\n", readme_path + ".forjj_tmp", readme_path, err)
            os.Exit(1)
        }
    }
}

