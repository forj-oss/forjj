package forjfile

import (
	"fmt"
	"forjj/git"
	"forjj/utils"
	"os"
	"path"

	"github.com/forj-oss/forjj-modules/trace"
)

// GitSetRepo define where the Deployment repo is located.
func (d *DeploymentCoreStruct) GitSetRepo(aPath, origin string) (err error) {
	if v, err := utils.Abs(path.Join(aPath, d.name)); err != nil {
		return err
	} else {
		d.repoPath = v
	}

	if err = git.EnsureRepoExist(d.repoPath); err != nil {
		return fmt.Errorf("Unable to create Deployment git repository. %s", err)
	}

	if origin != "" {
		err = d.GitDefineRemote("origin", origin)
		d.GitSyncFrom("origin", "master")
	}

	return
}

// GitDefineRemote helps to configure a deployment repository with a remote
func (d *DeploymentCoreStruct) GitDefineRemote(name, uri string) (err error) {
	return d.runInContext(func() (err error) {
		if err = os.Chdir(d.repoPath); err != nil {
			return
		}

		if err = git.EnsureRemoteIs(name, uri); err != nil {
			return
		}
		return
	})
}

// GitSyncFrom refresh the remote, and synchronize.
func (d *DeploymentCoreStruct) GitSyncFrom(remote, branch string) {
	d.runInContext(func() (_ error) {
		remoteBranch := remote + "/" + branch
		if git.Do("fetch", remote) == 0 {
			git.Do("reset", remoteBranch)
			git.Do("branch", "--set-upstream-to="+remoteBranch)
		} else {
			gotrace.Info("Your remote '%s' is currently not available. Synchonization delayed.", remoteBranch)
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
	if err = d.savePath(); err != nil {
		return err
	}
	defer d.restorePath()
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
	return
}
