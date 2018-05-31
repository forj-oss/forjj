package forjfile

// RepoModelStruct is the RepoStruct model
type RepoModelStruct struct {
	data *RepoStruct
}

// From build the repo model from a RepoStruct
func (r *RepoModelStruct) From(repo *RepoStruct) {
	r.data = repo
}

// Get return value for any recognized fields of a repository object.
func (r RepoModelStruct) Get(field string) string {
	return r.data.GetString(field)
}

// RemoteUrl return the remote URL field
func (r RepoModelStruct) RemoteUrl() string {
	return r.data.RemoteUrl()
}

// RemoteType return the remote type field
func (r RepoModelStruct) RemoteType() string {
	return r.data.RemoteType()
}

// UpstreamAPIUrl return the remote API url field
func (r RepoModelStruct) UpstreamAPIUrl() string {
	return r.data.UpstreamAPIUrl()
}

// Role return the repository role
func (r RepoModelStruct) Role() string {
	return r.data.GetString("role")
}

// Owner return the repository owner field
func (r RepoModelStruct) Owner() string {
	return r.data.Owner()
}
