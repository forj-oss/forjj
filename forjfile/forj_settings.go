package forjfile

// forj/settings: Collection of key/value pair
type ForjSettingsStruct struct {
	is_template bool
	forge *Forge
	Organization string
	ForjSettingsStructTmpl `yaml:",inline"`
}

type ForjSettingsStructTmpl struct {
	More map[string]string `yaml:",inline"`
}

func (f *ForjSettingsStruct) MarshalYAML() (interface{}, error) {
	return f.ForjSettingsStructTmpl, nil
}

func (s *ForjSettingsStruct) Get(key string) (value string, found bool) {
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

func (s *ForjSettingsStruct) Set(key string, value string) {
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

func (g *ForjSettingsStruct) set_forge(f *Forge) {
	g.forge = f
}
