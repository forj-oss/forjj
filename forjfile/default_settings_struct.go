package forjfile

import "github.com/forj-oss/goforjj"

type DefaultSettingsStruct struct {
	forge            *ForgeYaml
	UpstreamInstance string                `yaml:"upstream-instance,omitempty"` // TODO: to remove - obsolete
	Flow             string                `yaml:",omitempty"`
	More             map[string]string     `yaml:",inline"`
}


func (s *DefaultSettingsStruct) Get(key string) (value *goforjj.ValueStruct, _ bool) {
	switch key {
	// TODO: Remove obsolete reference to "upstream-instance"
	case "upstream-instance":
		return value.SetIfFound(s.UpstreamInstance, (s.UpstreamInstance != ""))
	case "flow":
		return value.SetIfFound(s.Flow, (s.Flow != ""))
	default:
		v, f := s.More[key]
		return value.SetIfFound(v, f)
	}
}

func (s *DefaultSettingsStruct) Set(key string, value string) {
	switch key {
	case "upstream-instance":
		if s.UpstreamInstance != value {
			s.UpstreamInstance = value
			s.forge.dirty()
		}
	case "flow":
		if s.Flow != value {
			s.Flow = value
			s.forge.dirty()
		}
		return
	default:
		if v, found := s.More[key]; found && value == "" {
			delete(s.More, key)
			s.forge.dirty()
		} else {
			if v != value {
				s.forge.dirty()
				s.More[key] = value
			}
		}
	}
}

func (d *DefaultSettingsStruct) set_forge(f *ForgeYaml) {
	d.forge = f
}

func (d *DefaultSettingsStruct) getFlow() string {
	return d.Flow
}
