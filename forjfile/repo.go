package forjfile

type RepoStruct struct {
	forge    *Forge
	Name     string
	Upstream string
	More     map[string]string `yaml:",inline"`
}

func (r *RepoStruct)Get(field string) (value string, found bool) {
	switch field {
	case "name":
		return r.Name, (r.Name != "")
	case "upstream":
		return r.Upstream, (r.Upstream != "")
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
