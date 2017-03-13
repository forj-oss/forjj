package forjfile

import "strings"

// forj/settings: Collection of key/value pair
type ForjSettingsStruct struct {
	is_template bool
	forge *ForgeYaml
	Organization string
	ForjSettingsStructTmpl `yaml:",inline"`
}

type ForjSettingsStructTmpl struct {
	Default DefaultSettingsStruct
	More map[string]string `yaml:",inline"`
}

type DefaultSettingsStruct struct {
	forge *ForgeYaml
	Flow string
	More map[string]string `yaml:",inline"`
}

func (f *ForjSettingsStruct) MarshalYAML() (interface{}, error) {
	return f.ForjSettingsStructTmpl, nil
}

func (s *ForjSettingsStruct) Get(key string) (value string, found bool) {
	defaults := strings.Split(key, ":")
	if defaults[0] == "default" {
		return s.Default.Get(defaults[1])
	}
	switch key {
	case "organization":
		if value = s.Organization ; value != "" {
			found = true
		}
		return
	default:
		value, found = s.More[key]
		return
	}
}

func (r *ForjSettingsStruct)SetHandler(from func(field string)(string), keys...string) {
	for _, key := range keys {
		r.Set(key, from(key))
	}
}

func (s *ForjSettingsStruct) Set(key string, value string) {
	defaults := strings.Split(key, ":")
	if defaults[0] == "default" {
		s.Default.Set(defaults[1], value)
		return
	}
	switch key {
	case "organization":
		s.Organization = value
		s.forge.dirty()
		return
	default:
		if v, found := s.More[key] ; found && v != value {
			s.forge.dirty()
			s.More[key] = value
		}
	}
}

func (g *ForjSettingsStruct) set_forge(f *ForgeYaml) {
	g.forge = f
}

func (s *DefaultSettingsStruct) Get(key string) (value string, found bool) {
	switch key {
	case "flow":
		if value = s.Flow ; value != "" {
			found = true
		}
		return
	default:
		value, found = s.More[key]
		return
	}
}



func (s *DefaultSettingsStruct) Set(key string, value string) {
	switch key {
	case "flow":
		s.Flow = value
		s.forge.dirty()
		return
	default:
		if v, found := s.More[key] ; found && v != value {
			s.forge.dirty()
			s.More[key] = value
		}
	}
}

func (d *DefaultSettingsStruct) set_forge(f *ForgeYaml) {
	d.forge = f
}
