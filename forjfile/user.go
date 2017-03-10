package forjfile

type UserStruct struct {
	forge *Forge
	Role string
	More map[string]string `yaml:",inline"`
}

// TODO: Add struct unit tests

func (u *UserStruct) Get(field string) (value string, found bool) {
	switch field {
	case "role":
		return u.Role, (u.Role != "")
	default:
		value, found = u.More[field]
	}
	return
}

func (u *UserStruct) Set(field, value string) {
	switch field {
	case "role":
		if u.Role != value {
			u.Role = value
			u.forge.dirty()
		}
	default:
		if u.More == nil {
			u.More = make(map[string]string)
		}
		if v, found := u.More[field] ; found && v != value {
			u.More[field] = value
			u.forge.dirty()
		}
	}
	return
}

func (r *UserStruct)set_forge(f *Forge) {
	r.forge = f
}
