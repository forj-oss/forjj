package forjfile

import "github.com/forj-oss/goforjj"

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
	Upstream     string `yaml:"upstream-app"`
	GitRemote    string `yaml:"git-remote"`
	remote       string // Git remote string to use/set
	Title        string
	Flow         string
	RepoTemplate string `yaml:"repo-template"`
	More         map[string]string `yaml:",inline"`
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

func (r *RepoStruct)Get(field string) (value *goforjj.ValueStruct, _ bool) {
	switch field {
	case "name":
		return value.Set(r.name, (r.name != ""))
	case "upstream":
		return value.Set(r.Upstream, (r.Upstream != ""))
	case "git-remote":
		return value.Set(r.GitRemote, (r.GitRemote != ""))
	case "remote":
		return value.Set(r.remote, (r.remote != ""))
	case "title":
		return value.Set(r.Title, (r.Title != ""))
	case "flow":
		return value.Set(r.Flow, (r.Flow != ""))
	case "repo-template":
		return value.Set(r.RepoTemplate, (r.RepoTemplate != ""))
	default:
		v, f := r.More[field]
		return value.Set(v, f)
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
		if r.remote != value {
			r.remote = value
			r.forge.dirty()
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

func (r *RepoStruct)set_forge(f *ForgeYaml) {
	r.forge = f
}
