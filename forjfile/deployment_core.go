package forjfile

import (
	"path"
)

// DeploymentCoreStruct contains only deployment information. anything others kind of information
type DeploymentCoreStruct struct {
	repoPath         string      // Absolute path to the repository.
	savedPath        string      // Internal saved path, restored when git command is used.
	name             string      // Name of the repository
	syncStatus       int         // 0 if sync has not been checked, 1 if succeed, -1 if remote exist but empty, -2 remote doesn't exist
	syncRemote       string      // string representing the remote to sync up. Usually origin/master.
	syncRemoteBranch string      // string representing the remote branch to sync up. Usually origin/master.
	syncUpstream     string      // string representing the upstream remote branch to pull from
	Desc             string      `yaml:"description,omitempty"`
	repo             *RepoStruct // Source code Repository for the deployment.
	Type             string
	Pars             map[string]string `yaml:"parameters,omitempty"`
}

// GetRepoPath returns the absolute path of the current deployment repository.
func (d *DeploymentCoreStruct) GetRepoPath() string {
	return d.repoPath
}

// GetReposPath returns the absolute path to all deployment repositories.
func (d *DeploymentCoreStruct) GetReposPath() string {
	return path.Dir(d.repoPath)
}

func (d *DeploymentCoreStruct) Name() string {
	return d.name
}

// InSync return true is deployment local repo is already connected and were synced at east once.
// If not sync, you may need to call d.SyncUp()
func (d *DeploymentCoreStruct) InSync() bool {
	return d.syncStatus == 1
}

// InSync return true is deployment local repo is already connected and were synced at east once.
// If not sync, you may need to call d.SyncUp()
func (d *DeploymentCoreStruct) GitRemoteReady() bool {
	return (d.syncStatus != 0 && d.syncStatus != -2)
}

func (d *DeploymentCoreStruct) AttachRepo(repo *RepoStruct, organization string) {
	if d == nil {
		return
	}
	d.repo = repo
	repo.isDeploy = true
	repo.deployment = d.name
	repoName := organization + "-" + d.name
	if repoName == repo.name {
		return
	}
}

// AttachedRepo return the Deployment attached repository.
// If found, it returns the repo attached. nil otherwise.
func (d *DeploymentCoreStruct) AttachedRepo() (repo *RepoStruct) {
	if d == nil {
		return
	}
	repo = d.repo
	return
}
