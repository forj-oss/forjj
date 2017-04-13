package forjfile

import "fmt"

type AppStruct struct {
	forge *ForgeYaml
	name   string
	AppYamlStruct `yaml:",inline"`
}

type AppYamlStruct struct {
	Type   string
	Driver string
	// TODO: Support for object dedicated to the application instance (not shared)
	// Objects map[string]map[string]string
	More   map[string]string `yaml:",inline"`
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

	a.AppYamlStruct = app
	return nil
}

func (a *AppStruct) MarshalYAML() (interface{}, error) {
	// Ensure we write only Yaml Data.
	return a.AppYamlStruct, nil
}

func (a *AppStruct)Name() string {
	return a.name
}

func (a *AppStruct)Get(flag string) (value string, found bool) {
	switch flag {
	case "name":
		return a.name, true
	case "type":
		return a.Type, true
	case "driver":
		return a.Driver, true
	default:
		value, found = a.More[flag]
		return
	}
}

func (r *AppStruct)SetHandler(from func(field string)(string, bool), keys...string) {
	for _, key := range keys {
		if v, found := from(key) ; found {
			r.Set(key, v)
		}
	}
}

func (a *AppStruct)Set(flag, value string) {
	switch flag {
	case "name":
		a.name = value
	case "type":
		a.Type = value
	case "driver":
		a.Driver = value
	default:
		if a.More == nil { a.More = make(map[string]string) }
		a.More[flag] = value
	}
	a.forge.dirty()
	return
}

func (g *AppStruct) set_forge(f *ForgeYaml) {
	g.forge = f
}
