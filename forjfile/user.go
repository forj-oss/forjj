package forjfile

import "github.com/forj-oss/goforjj"

type UsersStruct map[string]*UserStruct

func (u UsersStruct) mergeFrom(from UsersStruct) {
	for k, userFrom := range from {
		if user, found := u[k]; found {
			user.mergeFrom(userFrom)
		} else {
			u[k] = userFrom
		}
	}
}

const (
	userRole = "role"
)

type UserStruct struct {
	forge *ForgeYaml
	Role  string
	More  map[string]string `yaml:",inline"`
}

// TODO: Add struct unit tests

// Flags returns the list of keys found in this object.
func (u *UserStruct) Flags() (flags []string){
	flags = make([]string, 1, 1 + len(u.More))
	flags[0] = userRole
	for k := range u.More {
		flags = append(flags, k)
	}
	return

}

func (u *UserStruct) mergeFrom(from *UserStruct) {
	for _, flag := range from.Flags() {
		if v, found := from.Get(flag); found {
			u.Set(flag, v.GetString())
		}
	}
}

func (u *UserStruct) Get(field string) (value *goforjj.ValueStruct, _ bool) {
	switch field {
	case userRole:
		return value.SetIfFound(u.Role, (u.Role != ""))
	default:
		v, f := u.More[field]
		return value.SetIfFound(v, f)
	}
	return
}

func (r *UserStruct) SetHandler(from func(field string) (string, bool), keys ...string) {
	for _, key := range keys {
		if v, found := from(key); found {
			r.Set(key, v)
		}
	}
}

func (u *UserStruct) Set(field, value string) {
	switch field {
	case userRole:
		if u.Role != value {
			u.Role = value
			u.forge.dirty()
		}
	default:
		if u.More == nil {
			u.More = make(map[string]string)
		}
		if v, found := u.More[field]; found && value == "" {
			u.More[field] = value
			u.forge.dirty()
		} else {
			if v != value {
				u.forge.dirty()
				u.More[field] = value
			}
		}
	}
	return
}

func (r *UserStruct) set_forge(f *ForgeYaml) {
	r.forge = f
}
