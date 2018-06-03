package forjfile

import (
	"fmt"
	"forjj/git"
	"forjj/utils"
	"os"
	"path"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"
)

// GitSetRepo define where the Deployment repo is located. It creates the repo even just empty and sync if possible and origin given.
// It switch to master branch
func (d *DeploymentCoreStruct) GitSetRepo(aPath, origin string) (err error) {
	d.repoPath, err = utils.Abs(path.Join(aPath, d.name))
	if err != nil {
		return err
	}

	if err = git.EnsureRepoExist(d.repoPath); err != nil {
		return fmt.Errorf("Unable to create Deployment git repository. %s", err)
	}

	if origin != "" {
		err = d.GitDefineRemote("origin", origin)
		d.SwitchTo("master")
		d.GitSyncFrom("origin", "master")
	}

	return
}

// GitDefineRemote helps to configure a deployment repository with a remote
func (d *DeploymentCoreStruct) GitDefineRemote(name, uri string) (err error) {
	return d.runInContext(func() (err error) {
		if err = git.EnsureRemoteIs(name, uri); err != nil {
			return
		}
		return
	})
}

// GitSyncFrom refresh the remote, and synchronize.
func (d *DeploymentCoreStruct) GitSyncFrom(remote, branch string) error {
	return d.runInContext(func() (_ error) {
		if !git.RemoteExist(remote) {
			return
		}
		d.syncRemoteBranch = remote + "/" + branch
		d.syncRemote = remote
		d.syncStatus = 2 // Doing the sync up
		return d.GitSyncUp()
	})
}

// GitSyncUp set and report sync status
func (d *DeploymentCoreStruct) GitSyncUp() error {
	return d.runInContext(func() (_ error) {
		if d.syncStatus == 0 {
			return fmt.Errorf("Internal error! Unable to sync up. The synchronization was not initiliazed. You must call GitSyncFrom, Once")
		}
		if git.Do("fetch", d.syncRemote) == 0 {
			if found, _ := git.RemoteBranchExist(d.syncRemoteBranch); found {
				git.Do("reset", "--soft", d.syncRemoteBranch)
				git.Do("branch", "--set-upstream-to="+d.syncRemoteBranch)
				d.syncStatus = 1
			} else {
				d.syncStatus = -1
			}
		} else {
			d.syncStatus = -2
		}
		return
	})
}

// SwitchTo move to the requested branch
// if files were updated, they are stashed and restored in the other branch
// !!! Conflict can happen !!!
//
func (d *DeploymentCoreStruct) SwitchTo(branch string) error {
	return d.runInContext(func() (err error) {
		if git.GetCurrentBranch() != branch {

			trackedFiles := git.GetStatus().CountTracked()
			if trackedFiles > 0 {
				git.Do("stash")
			}
			git.Do("reset", "--hard", "HEAD")
			if found, err := git.BranchExist(branch); err != nil {
				return err
			} else if found {
				git.Do("checkout", branch)
			} else {
				git.Do("checkout", "-b", branch)
			}
			if trackedFiles > 0 {
				git.Do("stash", "pop")
			}
		}
		return
	})
}

// GitCommit do the commit in the Deployment repository.
func (d *DeploymentCoreStruct) GitCommit(message string) (_ error) {
	return d.runInContext(func() (err error) {
		status := git.GetStatus()
		if status.Ready.CountFiles() > 0 {
			git.Commit(message, true)
		}
		return
	})
}

// GitPush do a git push
// depending on the previous Git SyncFrom, a push can take place
func (d *DeploymentCoreStruct) GitPush(force bool) (_ error) {
	return d.runInContext(func() (err error) {
		if d.syncStatus == -2 {
			return fmt.Errorf("Unable to push to an inexistent remote")
		}
		if d.syncStatus == 0 {
			return fmt.Errorf("Unable to push. You need to sync up before")
		}
		push := make([]string, 1, 4)
		push[0] = "push"
		if force {
			push = append(push, "-f")
		}
		if d.syncStatus == -1 {
			push = append(push, "-u")
			push = append(push, strings.Split(d.syncRemoteBranch, "/")...)
		}
		if git.Do(push...) != 0 {
			err = fmt.Errorf("Unable to push")
		} else {
			d.syncStatus = 1
		}
		return
	})
}

// GitResetBranchFromRemote clean current branch, check out to the requested branch and reset against remote branch.
// The reset is not made if the fetch return an error.
func (d *DeploymentCoreStruct) GitResetBranchFromRemote(branch, remote string) {
	d.runInContext(func() (_ error) {
		git.Do("reset", "--hard", "HEAD")
		git.Do("checkout", branch)
		if git.Do("fetch", remote) == 0 {
			git.Do("reset", "--hard", remote+"/"+branch)
		}
		return
	})
}

// ------------------ Internal functions

// runInContext ensure GIT commands are executed in the right GIT repo context.
func (d *DeploymentCoreStruct) runInContext(doRun func() error) (err error) {
	if d.savedPath != "" {
		return doRun()
	}

	if err = d.savePath(); err != nil {
		return err
	}
	defer d.restorePath()
	if err = os.Chdir(d.repoPath); err != nil {
		return
	}
	return doRun()
}

// savePath Save and move to the repo path of this Deployment repo.
func (d *DeploymentCoreStruct) savePath() error {
	if d.repoPath == "" {
		return fmt.Errorf("repoPath is empty. Unable to move to the repository")
	}
	v, err := os.Getwd()
	if err != nil {
		return err
	}
	if v == d.repoPath {
		return nil
	}
	err = os.Chdir(d.repoPath)
	if err != nil {
		return err
	}
	d.savedPath = v
	gotrace.Trace("GIT: Moved to '%s'", d.repoPath)
	git.Indent("---- GIT"+git.ShowGitPath(), " - ", "--------")
	return err
}

// restorePath restore the saved path set by savePath()
func (d *DeploymentCoreStruct) restorePath() (err error) {
	if d.savedPath == "" {
		return
	}
	err = os.Chdir(d.savedPath)
	if err != nil {
		return err
	}
	gotrace.Trace("GIT: Restored to '%s'", d.savedPath)
	d.savedPath = ""
	git.UnIndent()
	return
}
