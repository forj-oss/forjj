package forjfile

import (
	"forjj/sources_info"
	"github.com/forj-oss/goforjj"
)

const (
	userRole = "role"
)

type UserStruct struct {
	forge   *ForgeYaml
	Role    string
	More    map[string]string `yaml:",inline"`
	sources *sourcesinfo.Sources
}

// TODO: Add struct unit tests

// Flags returns the list of keys found in this object.
func (u *UserStruct) Flags() (flags []string) {
	flags = make([]string, 1, 1+len(u.More))
	flags[0] = userRole
	for k := range u.More {
		flags = append(flags, k)
	}
	return

}

func (u *UserStruct) mergeFrom(from *UserStruct) {
	for _, flag := range from.Flags() {
		if v, found, source := from.Get(flag); found {
			u.Set(source, flag, v.GetString())
		}
	}
}

func (u *UserStruct) Get(field string) (value *goforjj.ValueStruct, found bool, source string) {
	source = u.sources.Get(field)
	switch field {
	case userRole:
		value, found = value.SetIfFound(u.Role, (u.Role != ""))
	default:
		v, f := u.More[field]
		value, found = value.SetIfFound(v, f)
	}
	return
}

func (r *UserStruct) SetHandler(source string, from func(field string) (string, bool), keys ...string) {
	for _, key := range keys {
		if v, found := from(key); found {
			r.Set(source, key, v)
		}
	}
}

func (u *UserStruct) Set(source, field, value string) {
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
		if v, found := u.More[field]; found {
			if v != value {
				u.forge.dirty()
				u.More[field] = value
			}
		} else {
			u.forge.dirty()
			u.More[field] = value
		}
	}
	u.sources = u.sources.Set(source, field, value)
	return
}

func (r *UserStruct) set_forge(f *ForgeYaml) {
	r.forge = f
}
