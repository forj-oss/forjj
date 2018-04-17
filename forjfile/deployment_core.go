package forjfile

import (
	"path"
)

// DeploymentCoreStruct contains only deployment information. anything others kind of information
type DeploymentCoreStruct struct {
	repoPath  string // Absolute path to the repository.
	savedPath string // Internal saved path, restored when git command is used.
	name      string // Name of the repository
	Desc      string `yaml:"description,omitempty"`
	Type      string
	Pars      map[string]string `yaml:"parameters,omitempty"`
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