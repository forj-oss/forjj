package forjfile

import (
	"fmt"
	"github.com/forj-oss/goforjj"
)

type AppStruct struct {
	forge *ForgeYaml
	name   string
	AppYamlStruct `yaml:",inline"`
}

type AppYamlStruct struct {
	Type   string
	Driver string
	Version string
	// TODO: Support for object dedicated to the application instance (not shared)
	// Objects map[string]map[string]string
	Flows map[string]AppFlowYaml `yaml:",omitempty"`
	more   ForjValues
	More   map[string]string `yaml:",inline"`
}

type AppFlowYaml struct {
	name string
	Service string `yaml:"used-as,omitempty"`
	Options map[string]string
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
		app.more[key] = ForjValue{value:value}
	}

	a.AppYamlStruct = app
	return nil
}

func (a *AppStruct) MarshalYAML() (interface{}, error) {
	// Ensure we write only Yaml Data.
	a.AppYamlStruct.More = a.AppYamlStruct.more.Map()
	return a.AppYamlStruct, nil
}

func (a *AppStruct)Name() string {
	return a.name
}

func (a *AppStruct)Get(flag string) (value *goforjj.ValueStruct, _ bool) {
	switch flag {
	case "name":
		return value.Set(a.name), true
	case "type":
		return value.Set(a.Type), true
	case "driver":
		return value.Set(a.Driver), true
	default:
		v, f := a.more[flag]
		return value.SetIfFound(v.Get(), f)
	}
	return
}

func (r *AppStruct)SetHandler(from func(field string)(string, bool), set func(*ForjValue, string) (bool), keys...string) {
	for _, key := range keys {
		if v, found := from(key) ; found {
			r.Set(key, v, set)
		}
	}
}

func (a *AppStruct)Set(flag, value string, set func(*ForjValue, string) (bool)) {
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
		if a.more == nil { a.more = make(ForjValues) }
		if v, found := a.more[flag] ; found && value == "" {
			a.forge.dirty()
			delete(a.more, flag)
		} else {
			if set(&v, value) {
				a.forge.dirty()
				a.more[flag] = v
			}
		}
	}
	return
}

func (g *AppStruct) set_forge(f *ForgeYaml) {
	g.forge = f
}
