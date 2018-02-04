package forjfile

import (
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
	"strings"
)

// forj/settings: Collection of key/value pair
type ForjSettingsStruct struct {
	is_template            bool
	forge                  *ForgeYaml
	Organization           string
	ForjSettingsStructTmpl `yaml:",inline"`
}

type ForjSettingsStructTmpl struct {
	Default  DefaultSettingsStruct
	RepoApps DefaultRepoAppSettingsStruct `yaml:"default-repo-apps,omitempty"` // Default repo Application
	More     map[string]string            `yaml:",inline"`
}

type DefaultRepoAppSettingsStruct map[string]string

func (f *ForjSettingsStruct) MarshalYAML() (interface{}, error) {
	return f.ForjSettingsStructTmpl, nil
}

func (s *ForjSettingsStruct) Get(instance, key string) (value *goforjj.ValueStruct, _ bool) {
	switch instance {
	case "default":
		return s.Default.Get(key)
	case "default-repo-apps":
		if v, found := s.RepoApps[key]; found {
			return value.SetIfFound(v, (v != ""))
		}
		if key != "upstream" {
			return
		}
		// TODO: Remove obsolete reference to "upstream-instance"
		gotrace.Warning("Forjfile: `forj-settings/default/upstream-instance` is obsolete and will be ignored in the future." +
			" Please use `forj-settings/default-repo-apps/upstream` instead.")
		return s.Default.Get("upstream-instance")
	}
	switch key {
	case "organization":
		return value.SetIfFound(s.Organization, (s.Organization != ""))
	default:
		v, f := s.More[key]
		return value.SetIfFound(v, f)
	}
}

func (s *ForjSettingsStruct) GetInstance(instance string) interface{} {
	if instance == "default" {
		return s.Default
	}
	return s
}

func (r *ForjSettingsStruct) SetHandler(instance string, from func(field string) (string, bool), keys ...string) {
	for _, key := range keys {
		if v, found := from(key); found {
			r.Set(instance, key, v)
		}
	}
}

func (s *ForjSettingsStruct) Set(instance, key string, value string) {
	switch instance {
	case "default":
		s.Default.Set(key, value)
		s.forge.dirty()
		return
	case "default-repo-apps":
		if s.RepoApps == nil {
			s.RepoApps = make(map[string]string)
		}
		s.RepoApps[key] = value
		relApp := strings.Split(value, ":")
		if len(relApp) == 1 {
			s.forge.Repos.SetRelapps(key, value)
		} else {
			s.forge.Repos.SetRelapps(relApp[0], relApp[1])
		}
		s.forge.dirty()
		return
	}
	switch key {
	case "organization":
		s.Organization = value
		s.forge.dirty()
		return
	default:
		if v, found := s.More[key]; found && v != value {
			s.forge.dirty()
			s.More[key] = value
		}
	}
}

func (g *ForjSettingsStruct) set_forge(f *ForgeYaml) {
	g.forge = f
	g.Default.set_forge(f)
}
