package forjfile

import (
	"github.com/forj-oss/goforjj"
	"forjj/drivers"
)

type ReposStruct map[string]*RepoStruct

func (r ReposStruct) MarshalYAML() (interface{}, error) {
	to_marshal := make(map[string]*RepoStruct)
	for name, repo := range r {
		if ! repo.is_infra {
			to_marshal[name] = repo
		}
	}
	return to_marshal, nil
}

type RepoStruct struct {
	name         string
	is_infra     bool
	forge        *ForgeYaml
	owner        string
	driverOwner  *drivers.Driver
	Upstream     string `yaml:"upstream-app,omitempty"` // Name of the application upstream hosting this repository.
	GitRemote    string `yaml:"git-remote,omitempty"`
	remote       goforjj.PluginRepoRemoteUrl // Git remote string to use/set
	Title        string `yaml:",omitempty"`
	Flow         string `yaml:",omitempty"`
	RepoTemplate string `yaml:"repo-template,omitempty"`
	More         map[string]string `yaml:",inline"`
}

func (r *RepoStruct)Owner() string {
	return r.owner
}

func (r *RepoStruct)setFromInfra(infra *RepoStruct) {
	if r == nil {
		return
	}
	*r = *infra
	delete(r.More, "name")
	r.is_infra = true
}

func (r *RepoStruct)setToInfra(infra *RepoStruct) {
	*infra = *r
	infra.is_infra = false // Unset it to ensure data is saved in yaml
}

func (r *RepoStruct)GetString(field string) (string) {
	if v, found := r.Get(field) ; found {
		return v.GetString()
	}
	return ""
}

func (r *RepoStruct)RemoteUrl() string {
	return r.remote.Url
}

func (r *RepoStruct)RemoteGit() string {
	return r.remote.Ssh
}

func (r *RepoStruct)Get(field string) (value *goforjj.ValueStruct, _ bool) {
	switch field {
	case "name":
		return value.SetIfFound(r.name, (r.name != ""))
	case "upstream":
		return value.SetIfFound(r.Upstream, (r.Upstream != ""))
	case "git-remote":
		return value.SetIfFound(r.GitRemote, (r.GitRemote != ""))
	case "remote":
		return value.SetIfFound(r.remote.Ssh, (r.remote.Ssh != ""))
	case "remote-url":
		return value.SetIfFound(r.remote.Url, (r.remote.Url != ""))
	case "title":
		return value.SetIfFound(r.Title, (r.Title != ""))
	case "flow":
		return value.SetIfFound(r.Flow, (r.Flow != ""))
	case "repo-template":
		return value.SetIfFound(r.RepoTemplate, (r.RepoTemplate != ""))
	default:
		v, f := r.More[field]
		return value.SetIfFound(v, f)
	}
	return
}

func (r *RepoStruct)SetHandler(from func(field string)(string, bool), keys...string) {
	for _, key := range keys {
		if v, found := from(key) ; found {
			r.Set(key, v)
		}
	}
}

func (r *RepoStruct)Set(field, value string) {
	if r == nil {
		return
	}
	switch field {
	case "name":
		if r.name != value {
			r.name = value
		}
	case "upstream":
		if r.Upstream != value {
			r.Upstream = value
			r.forge.dirty()
		}
	case "git-remote":
		if r.GitRemote != value {
			r.GitRemote = value
			r.forge.dirty()
		}
	case "remote":
		if r.remote.Ssh != value {
			r.remote.Ssh = value
		}
	case "remote-url":
		if r.remote.Url != value {
			r.remote.Url = value
		}
	case "repo-template":
		if r.RepoTemplate != value {
			r.RepoTemplate = value
			r.forge.dirty()
		}
	case "title":
		if r.Title != value {
			r.Title = value
			r.forge.dirty()
		}
	case "flow":
		if r.Flow != value {
			r.Flow = value
			r.forge.dirty()
		}
	default:
		if r.More == nil {
			r.More = make(map[string]string)
		}
		if v, found := r.More[field] ; found && v != value {
			r.More[field] = value
			r.forge.dirty()
		}
	}
}

func (r *RepoStruct)SetInstanceOwner(owner string) {
	r.owner = owner
}

func (r *RepoStruct)SetPluginOwner(d *drivers.Driver) {
	r.driverOwner = d
}

func (r *RepoStruct)RemoteType() string {
	if r.driverOwner == nil {
		return "git"
	}
	return r.driverOwner.Name
}

func (r *RepoStruct)UpstreamAPIUrl() string {
	if r.driverOwner == nil {
		return ""
	}
	return r.driverOwner.DriverAPIUrl
}

func (r *RepoStruct)set_forge(f *ForgeYaml) {
	r.forge = f
}
