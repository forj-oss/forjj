package forjfile

import (
	"fmt"
	"forjj/sources_info"

	"github.com/forj-oss/goforjj"
)

const (
	appName   = "name"
	appType   = "type"
	appDriver = "driver"
)

type AppStruct struct {
	forge         *ForgeYaml
	name          string
	sources       *sourcesinfo.Sources
	AppYamlStruct `yaml:",inline"`
}

type AppYamlStruct struct {
	Type    string
	Driver  string
	Version string
	// TODO: Support for object dedicated to the application instance (not shared)
	// Objects map[string]map[string]string
	Flows map[string]AppFlowYaml `yaml:",omitempty"`
	more  ForjValues
	More  map[string]string `yaml:",inline"`
}

type AppFlowYaml struct {
	name    string
	Service string `yaml:"used-as,omitempty"`
	Options map[string]string
}

func NewAppStruct() (ret *AppStruct) {
	ret = new(AppStruct)
	return
}

func (a *AppStruct) Flags() (flags []string) {
	flags = make([]string, 3, 3+len(a.more))
	flags[0] = appName
	flags[1] = appType
	flags[2] = appDriver
	for k := range a.more {
		flags = append(flags, k)
	}
	return
}

func (r *AppStruct) Model() AppModel {
	model := AppModel{
		app: r,
	}

	return model
}

// TODO: Add struct unit tests

func (a *AppStruct) UnmarshalYAML(unmarchal func(interface{}) error) error {
	var app AppYamlStruct

	if err := unmarchal(&app); err != nil {
		return err
	}
	if app.Type == "" {
		return fmt.Errorf("Application type value is required.")
	}

	app.more = make(ForjValues)
	for key, value := range app.More {
		app.more[key] = ForjValue{value: value}
	}

	a.AppYamlStruct = app
	return nil
}

func (a *AppStruct) MarshalYAML() (interface{}, error) {
	// Ensure we write only Yaml Data.
	a.AppYamlStruct.More = a.AppYamlStruct.more.Map()
	return a.AppYamlStruct, nil
}

func (a *AppStruct) Name() string {
	return a.name
}

// Get return the flag value.
// found is true if value exist in more or if the value is not empty
func (a *AppStruct) Get(flag string) (value *goforjj.ValueStruct, found bool, source string) {
	source = a.sources.Get(flag)
	switch flag {
	case appName:
		value, found = value.Set(a.name), (a.name != "")
	case appType:
		value, found = value.Set(a.Type), (a.Type != "")
	case appDriver:
		value, found = value.Set(a.Driver), (a.Driver != "")
	default:
		v, f := a.more[flag]
		value, found = value.SetIfFound(v.Get(), f)
	}
	return
}

func (r *AppStruct) SetHandler(source string, from func(field string) (string, bool), set func(*ForjValue, string) bool, keys ...string) {
	for _, key := range keys {
		if v, found := from(key); found {
			r.Set(source, key, v, set)
		}
	}
}

func (a *AppStruct) Set(source, flag, value string, set func(*ForjValue, string) bool) {
	switch flag {
	case "name":
		if a.name != value {
			a.forge.dirty()
			a.name = value
		}
	case "type":
		if a.Type != value {
			a.Type = value
			a.forge.dirty()
		}
	case "driver":
		if a.Driver != value {
			a.Driver = value
			a.forge.dirty()
		}
	default:
		if a.more == nil {
			a.more = make(ForjValues)
		}
		if v, found := a.more[flag]; found && value == "" {
			a.forge.dirty()
			delete(a.more, flag)
		} else {
			if set(&v, value) {
				a.forge.dirty()
				a.more[flag] = v
			}
		}
	}
	a.sources = a.sources.Set(source, flag, value)
	return
}

func (g *AppStruct) set_forge(f *ForgeYaml) {
	g.forge = f
}

func (a *AppStruct) mergeFrom(from *AppStruct) {
	for _, flag := range from.Flags() {
		if v, found, source := from.Get(flag); found {
			a.Set(source, flag, v.GetString(), (*ForjValue).Set)
		}
	}
}
