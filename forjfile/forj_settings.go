package forjfile

import (
	"forjj/sources_info"
	"strings"

	"github.com/forj-oss/goforjj"
)

// forj/settings: Collection of key/value pair
type ForjSettingsStruct struct {
	is_template            bool
	forge                  *ForgeYaml
	sources                *sourcesinfo.Sources
	Organization           string
	ForjSettingsStructTmpl `yaml:",inline"`
}

const (
	settingsDeployTo     = "deploy-to"
	settingsOrganization = "organization"
	settingsDefault      = "default"
)

func (s *ForjSettingsStruct) Flags() (flags []string) {
	flags = make([]string, 3, 3+len(s.More))
	flags[0] = settingsDefault
	flags[1] = settingsDeployTo
	flags[2] = settingsOrganization
	for k := range s.More {
		flags = append(flags, k)
	}
	return
}

func (s *ForjSettingsStruct) mergeFrom(from *ForjSettingsStruct) {
	for _, instance := range []string{"default", "default-repo-apps", "noinstance"} {
		for _, flag := range from.Flags() {
			if v, found, source := from.Get(instance, flag); found {
				s.Set(source, instance, flag, v.GetString())
			}
		}
	}

}

type ForjSettingsStructTmpl struct {
	Default  DefaultSettingsStruct
	RepoApps DefaultRepoAppSettingsStruct `yaml:"default-repo-apps,omitempty"` // Default repo Application
	More     map[string]string            `yaml:",inline"`
}

func (f *ForjSettingsStruct) MarshalYAML() (interface{}, error) {
	return f.ForjSettingsStructTmpl, nil
}

func (s *ForjSettingsStruct) Get(instance, key string) (value *goforjj.ValueStruct, found bool, source string) {
	source = s.sources.Get(key)
	switch instance {
	case "default":
		return s.Default.Get(key)
	case "default-repo-apps":
		if v, found2 := s.RepoApps.Get(key); found2 {
			value, found = value.SetIfFound(v, (v != ""))
			source = "forjj"
			return
		}
		if key != "upstream" {
			return
		}
		return s.Default.Get("upstream-instance")
	}
	switch key {
	case "organization":
		value, found = value.SetIfFound(s.Organization, (s.Organization != ""))
	default:
		v, f := s.More[key]
		value, found = value.SetIfFound(v, f)
	}
	return
}

func (s *ForjSettingsStruct) GetInstance(instance string) interface{} {
	if instance == "default" {
		return s.Default
	}
	return s
}

func (r *ForjSettingsStruct) SetHandler(source, instance string, from func(field string) (string, bool), keys ...string) {
	for _, key := range keys {
		if v, found := from(key); found {
			r.Set(source, instance, key, v)
		}
	}
}

func (s *ForjSettingsStruct) Set(source, instance, key, value string) {
	switch instance {
	case "default":
		s.Default.Set(source, key, value)
		s.forge.dirty()
		return
	case "default-repo-apps":
		if s.RepoApps == nil {
			s.RepoApps = make(map[string]string)
		}
		s.RepoApps[key] = value
		relApp := strings.Split(value, ":")
		if len(relApp) != 1 {
			key = relApp[0]
			value = relApp[1]
		}
		s.forge.ForjCore.Repos.SetRelapps(key, value)
		for _, deploy := range s.forge.Deployments {
			if deploy.Details != nil {
				deploy.Details.Repos.SetRelapps(key, value)
			}
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
	s.sources = s.sources.Set(source, key, value)
}

func (g *ForjSettingsStruct) set_forge(f *ForgeYaml) {
	g.forge = f
	g.Default.set_forge(f)
}
