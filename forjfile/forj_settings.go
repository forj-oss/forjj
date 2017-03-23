package forjfile

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
	UpstreamInstance string `yaml:"upstream-instance"`
	Flow string
	More map[string]string `yaml:",inline"`
}

func (f *ForjSettingsStruct) MarshalYAML() (interface{}, error) {
	return f.ForjSettingsStructTmpl, nil
}

func (s *ForjSettingsStruct) Get(instance, key string) (value string, found bool) {
	if instance == "default" {
		return s.Default.Get(key)
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

func (r *ForjSettingsStruct)SetHandler(instance string, from func(field string)(string, bool), keys...string) {
	for _, key := range keys {
		if v, found := from(key) ; found {
			r.Set(instance, key, v)
		}
	}
}

func (s *ForjSettingsStruct) Set(instance, key string, value string) {
	if instance == "default" {
		s.Default.Set(key, value)
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
	g.Default.set_forge(f)
}

func (s *DefaultSettingsStruct) Get(key string) (value string, found bool) {
	switch key {
	case "upstream-instance":
		if value = s.UpstreamInstance ; value != "" {
			found = true
		}
		return
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
	case "upstream-instance":
		s.UpstreamInstance = value
		s.forge.dirty()
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
