package forjfile

import (
	"fmt"
	"forjj/git"
	"forjj/utils"
	"os"
	"path"
)

// DeploymentStruct represent the data structure of all deployment.
type DeploymentStruct struct {
	DeploymentCoreStruct `yaml:",inline"`
	Details              *DeployForgeYaml `yaml:"define,omitempty"`
}

// DeploymentCoreStruct contains only deployment information. anything others kind of information
type DeploymentCoreStruct struct {
	repoPath  string // Absolute path to the repository.
	savedPath string // Internal saved path, restored when git command is used.
	name      string // Name of the repository
	Desc      string `yaml:"description,omitempty"`
	Type      string
	Pars      map[string]string `yaml:"parameters,omitempty"`
}

// MarshalYAML provides the encoding part for DeploymentStruct
//
// In short we do not want to encode forjj deployment details) info except the core.
func (d DeploymentStruct) MarshalYAML() (interface{}, error) {
	return d.DeploymentCoreStruct, nil
}

func (d *DeploymentStruct)UpdateDeploymentCoreData(data DeploymentCoreStruct) {
	d.DeploymentCoreStruct = data
}

// SetRepo define where the Deployment repo is located.
func (d *DeploymentCoreStruct) SetRepo(aPath, origin string) (err error) {
	if v, err := utils.Abs(path.Join(aPath, d.name)); err != nil {
		return err
	} else {
		d.repoPath = v
	}

	if err = git.EnsureRepoExist(d.repoPath); err != nil {
		return fmt.Errorf("Unable to create Deployment git repository. %s", err)
	}

	if origin != "" {
		err = d.DefineRemote("origin", origin)
		d.ResetBranchFromRemote("master", "origin")
	}

	return
}

func (d *DeploymentCoreStruct) savePath() error {
	if v, err := os.Getwd(); err != nil {
		return err
	} else {
		d.savedPath = v
	}
	return nil
}

func (d *DeploymentCoreStruct) restorepath() (err error) {
	if d.savedPath != "" {
		return
	}
	err = os.Chdir(d.savedPath)
	if err != nil {
		return err
	}
	d.savedPath = ""
	return
}

// DefineRemote helps to configure a deployment repository with a remote
func (d *DeploymentCoreStruct) DefineRemote(name, uri string) (err error) {
	if err := d.savePath(); err != nil {
		return err
	}
	defer d.restorepath()

	if err = os.Chdir(d.repoPath); err != nil {
		return
	}

	if err = git.EnsureRemoteIs(name, uri); err != nil {
		return err
	}

	return
}

// ResetBranchFromRemote clean current branch, check out to the requested branch and reset against remote branch.
// The reset is not made if the fetch return an error.
func (d *DeploymentCoreStruct) ResetBranchFromRemote(branch, remote string) {
	git.Do("reset", "--hard", "HEAD")
	git.Do("checkout", branch)
	if git.Do("fetch", remote) == 0 {
		git.Do("reset", "--hard", remote+"/"+branch)
	}
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