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
	if a == nil {
		return
	}
	flags = make([]string, 3, 3+len(a.more))
	flags[0] = appName
	flags[1] = appType
	flags[2] = appDriver
	for k := range a.more {
		flags = append(flags, k)
	}
	return
}

func (a *AppStruct) Model() AppModel {
	if a == nil {
		return AppModel{}
	}
	model := AppModel{
		app: a,
	}

	return model
}

// TODO: Add struct unit tests

func (a *AppStruct) UnmarshalYAML(unmarchal func(interface{}) error) error {
	if a == nil {
		return nil
	}
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
	if a == nil {
		return nil, nil
	}
	// Ensure we write only Yaml Data.
	a.AppYamlStruct.More = a.AppYamlStruct.more.Map()
	return a.AppYamlStruct, nil
}

func (a *AppStruct) Name() string {
	if a == nil {
		return ""
	}
	return a.name
}

// Get return the flag value.
// found is true if value exist in more or if the value is not empty
func (a *AppStruct) Get(flag string) (value *goforjj.ValueStruct, found bool, source string) {
	if a == nil {
		return
	}
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

func (a *AppStruct) SetHandler(source string, from func(field string) (string, bool), set func(*ForjValue, string) bool, keys ...string) (updated bool) {
	if a == nil {
		return
	}
	for _, key := range keys {
		if v, found := from(key); found {
			updated = a.set(source, key, v, set) || updated
		}
	}
	return
}

func (a *AppStruct) Set(source, flag, value string) (updated bool) {
	if a == nil {
		return
	}

	return a.set(source, flag, value, (*ForjValue).Set)
}

func (a *AppStruct) set(source, flag, value string, set func(*ForjValue, string) bool) (updated bool) {
	if a == nil {
		return
	}
	switch flag {
	case "name":
		if a.name != value {
			a.forge.dirty()
			a.name = value
			updated = true
		}
	case "type":
		if a.Type != value {
			a.Type = value
			a.forge.dirty()
			updated = true
		}
	case "driver":
		if a.Driver != value {
			a.Driver = value
			a.forge.dirty()
			updated = true
		}
	default:
		if a.more == nil {
			a.more = make(ForjValues)
		}
		if v, found := a.more[flag]; found && value == "" {
			a.forge.dirty()
			delete(a.more, flag)
			updated = true
		} else {
			if set(&v, value) {
				a.forge.dirty()
				a.more[flag] = v
				updated = true
			}
		}
	}
	a.sources = a.sources.Set(source, flag, value)
	return
}

func (a *AppStruct) set_forge(f *ForgeYaml) {
	if a == nil {
		return
	}
	a.forge = f
}

func (a *AppStruct) mergeFrom(from *AppStruct) {
	if a == nil {
		return
	}
	for _, flag := range from.Flags() {
		if v, found, source := from.Get(flag); found {
			a.set(source, flag, v.GetString(), (*ForjValue).Set)
		}
	}
}
