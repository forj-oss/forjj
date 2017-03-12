package forjfile

type RepoStruct struct {
	forge       *Forge
	Name        string
	Upstream    string `yaml:"upstream-app"`
	GitRemote   string `yaml:"git-remote"`
	remote      string // Git remote string to use/set
	More        map[string]string `yaml:",inline"`
}

func (r *RepoStruct)Get(field string) (value string, found bool) {
	switch field {
	case "name":
		return r.Name, (r.Name != "")
	case "upstream":
		return r.Upstream, (r.Upstream != "")
	case "git-remote":
		return r.GitRemote, (r.GitRemote != "")
	case "remote":
		return r.remote, (r.remote != "")
	default:
		value, found = r.More[field]
	}
	return
}

func (r *RepoStruct)Set(field, value string) {
	switch field {
	case "name":
		if r.Name != value {
			r.Name = value
			r.forge.dirty()
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

func (r *RepoStruct)set_forge(f *Forge) {
	r.forge = f
}
