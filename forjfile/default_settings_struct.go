package forjfile

import (
	"forjj/sources_info"
	"github.com/forj-oss/goforjj"
)

type DefaultSettingsStruct struct {
	forge            *ForgeYaml
	UpstreamInstance string            `yaml:"upstream-instance,omitempty"` // TODO: to remove - obsolete
	Flow             string            `yaml:",omitempty"`
	DevDeploy        string            `yaml:",omitempty"`
	More             map[string]string `yaml:",inline"`
	sources          *sourcesinfo.Sources
}

// Get return the value of the default setting.
func (s *DefaultSettingsStruct) Get(key string) (value *goforjj.ValueStruct, _ bool) {
	switch key {
	// TODO: Remove obsolete reference to "upstream-instance"
	case "upstream-instance":
		return value.SetIfFound(s.UpstreamInstance, (s.UpstreamInstance != ""))
	case "flow":
		return value.SetIfFound(s.Flow, (s.Flow != ""))
	case "dev-deploy":
		return value.SetIfFound(s.DevDeploy, (s.DevDeploy != ""))
	default:
		v, f := s.More[key]
		return value.SetIfFound(v, f)
	}
}

// Set udpate the value of the default setting key.
func (s *DefaultSettingsStruct) Set(source, key, value string) {
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
	case "dev-deploy":
		if s.DevDeploy != value {
			s.DevDeploy = value
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
	s.sources = s.sources.Set(source, key, value)
}

// set_forge set the forge reference.
func (s *DefaultSettingsStruct) set_forge(f *ForgeYaml) {
	s.forge = f
}

// getFlow return the default flow used.
func (s *DefaultSettingsStruct) getFlow() string {
	return s.Flow
}

// getDevDeploy return the default Development environment.
func (s *DefaultSettingsStruct) getDevDeploy() string {
	return s.DevDeploy
}
