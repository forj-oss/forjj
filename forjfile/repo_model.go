package forjfile

import (
	"github.com/forj-oss/forjj-modules/trace"
)

// RepoModelStruct is the RepoStruct model
type RepoModel struct {
	repo *RepoStruct
	Apps map[string]RepoAppModel
}

// From build the repo model from a RepoStruct
func (r *RepoModel) From(repo *RepoStruct) {
	r.repo = repo
}

// Get return value for any recognized fields of a repository object.
func (r RepoModel) Get(field string) string {
	return r.repo.GetString(field)
}

// RemoteUrl return the remote URL field
func (r RepoModel) RemoteUrl() string {
	return r.repo.RemoteUrl()
}

// RemoteType return the remote type field
func (r RepoModel) RemoteType() string {
	return r.repo.RemoteType()
}

// UpstreamAPIUrl return the remote API url field
func (r RepoModel) UpstreamAPIUrl() string {
	return r.repo.UpstreamAPIUrl()
}

// Role return the repository role
func (r RepoModel) Role() string {
	return r.repo.GetString("role")
}

// Owner return the repository owner field
func (r RepoModel) Owner() string {
	return r.repo.Owner()
}

// HasApps check is repo applications rules return true or not.
// a rule is true if the key:value is found in the application object attached.
//
// See details in RepoStruct.HasApps()
func (r *RepoModel) HasApps(rules ...string) (_ bool) {
	if r.repo == nil {
		return
	}
	if v, err := r.repo.HasApps(rules...); err != nil {
		gotrace.Error("%s", err)
	} else {
		return v
	}
	return
}

// IsCurrentDeploy returns true if the current repo is the current deployment repository.
func (r RepoModel) IsCurrentDeploy() bool {
	return r.repo.IsCurrentDeploy()
}

// IsDeployable return true if the repository identified is deployable in the current deployment context
func (r RepoModel) IsDeployable() bool {
	if r.repo.forge == nil {
		return false
	}
	if r.repo.forge.ForjCore.deployTo == "" {
		return false // We are not in a deployable context (no merge done between master and deployment Forjfiles)
	}
	if r.repo.Role() == "code" {
		return true
	}
	if r.repo.Role() == "infra" {
		return (r.repo.forge.ForjCore.deployTo == r.repo.deployment)
	}
	return r.repo.IsCurrentDeploy()
}