package forjfile

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
	*r = *infra
	delete(r.More, "name")
	r.is_infra = true
}

func (r *RepoStruct)setToInfra(infra *RepoStruct) {
	*infra = *r
	infra.is_infra = false // Unset it to ensure data is saved in yaml
}

func (r *RepoStruct) MarshalYAML() (interface{}, error) {
	if r.is_infra {
		// If a Repo is identified infra. do not save it.
		// This is used to save infra under `infra` section and NOT under `repositories/{repository}`
		return nil, nil

	}
	return r, nil
}

func (r *RepoStruct)Get(field string) (value string, found bool) {
	switch field {
	case "name":
		return r.name, (r.name != "")
	case "upstream":
		return r.Upstream, (r.Upstream != "")
	case "git-remote":
		return r.GitRemote, (r.GitRemote != "")
	case "remote":
		return r.remote, (r.remote != "")
	case "title":
		return r.Title, (r.Title != "")
	case "flow":
		return r.Flow, (r.Flow != "")
	case "repo-template":
		return r.RepoTemplate, (r.RepoTemplate != "")
	default:
		value, found = r.More[field]
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