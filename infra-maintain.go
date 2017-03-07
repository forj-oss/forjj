package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"forjj/git"
)

const (
	drivers_def_options_file  = ".forjj-maintain-options.yml"
	drivers_data_options_file = "forjj-creds.yml"
)

// FIXME: Def file keep just one driver at a time...
// FIXME: Fix maintain process from cli.

// This function save in the infra-repo a forjj configuration about
// plugins options required at maintain phase. forjj-maintain.yml
// It saves also values found at create/update phase in the workspace
// Forjj usually will requires a way to load this definition and load the appropriate value
// Passed to each plugins
// The file is saved, added and commited in the infra-repo.
type forjjPlugins struct {
	Plugins map[string]forjjPluginsOptions `,inline`
}

type forjjPluginsOptions struct {
	Type    string
	Options map[string]string `,omitempty`
}

// Save maintain credentials and definition files:
// - drivers_def_options_file in infra repo
// - drivers_data_options_file in a.w.Path(), drivers_data_options_file
func (a *Forj) SaveForjjPluginsOptions() error {
	if a.drivers_options.Drivers == nil {
		return nil
	}

	def := forjjPlugins{make(map[string]forjjPluginsOptions)}

	for driver, opts := range a.drivers_options.Drivers {
		o := forjjPluginsOptions{opts.Driver_type, make(map[string]string)}
		for option, v := range opts.Options {
			o.Options[option] = v.Help
		}
		def.Plugins[driver] = o
	}
	gotrace.Trace("Plugin options definition file content: %#v", def)

	if err := def.Save(drivers_def_options_file); err != nil {
		return fmt.Errorf("Unable to write '%s'. %s", drivers_def_options_file, err)
	}

	git.Do("add", drivers_def_options_file)

	for driver, opts := range a.drivers_options.Drivers {
		for option, v := range opts.Options {
			def.Plugins[driver].Options[option] = v.Value
		}
	}

	gotrace.Trace("Plugin options data file content: %#v", def)
	workspace_file := path.Join(a.w.Path(), drivers_data_options_file)
	if err := def.Save(workspace_file); err != nil {
		return fmt.Errorf("Unable to write '%s'. %s", workspace_file, err)
	}
	return nil
}

func (f *forjjPlugins) Save(file string) error {
	yaml_data, err := yaml.Marshal(f)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(file, yaml_data, 0644); err != nil {
		return err
	}
	gotrace.Trace("File name saved: %s", file)
	return nil
}

// This functions loads the forjj plugins options definitions in 'Maintain' phase context.
// 2 files have to be loaded. The definition in forj-repo and the one given at forjj cli.
func (a *Forj) LoadForjjPluginsOptions(creds_file string) error {
	// Read definition file from repo.
	var fpdef forjjPlugins // Plugins/<plugin>/Options/<option>=help

	filedef := path.Clean(path.Join(a.w.Path(), a.w.Infra.Name, drivers_def_options_file))
	if err := fpdef.LoadFile(filedef); err != nil {
		return err
	}
	gotrace.Trace("Plugin data definition file '%s' loaded.", filedef)

	// Load plugins Options data file, given to forjj
	var fpdata forjjPlugins // Plugins/<plugin>/Options/<option>=value

	if creds_file == "" {
		creds_file = path.Clean(path.Join(a.w.Path(), drivers_data_options_file))
		gotrace.Trace("Use default credential file '%s'.", creds_file)
	}
	if err := fpdata.LoadFile(creds_file); err != nil {
		return err
	}
	gotrace.Trace("Plugin credentials file '%s' loaded.", creds_file)

	// Load values in Forj.driver_options keys/values pair
	for name, p_opts := range fpdef.Plugins { // each plugin
		pluginOptions := make(map[string]goforjj.PluginOption)

		for opt_name, help := range p_opts.Options { // each options
			_, ok := fpdata.Plugins[name]
			value, ok2 := fpdata.Plugins[name].Options[opt_name]

			if ok && ok2 {
				pluginOptions[opt_name] = goforjj.PluginOption{Value: value}
			} else {
				return fmt.Errorf("Missing driver '%s' option '%s'. driver_type : '%s'. You must create and set it in '%s'\nBasic help: %s - %s",
					name, opt_name, p_opts.Type, creds_file, opt_name, help)
			}
		}
		a.drivers_options.AddForjjPluginOptions(name, pluginOptions, p_opts.Type)
	}
	return nil
}

func (fp *forjjPlugins) LoadFile(file string) error {
	yaml_data, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Unable to read '%s'. %s", drivers_data_options_file, err)
	}

	if err := yaml.Unmarshal(yaml_data, fp); err != nil {
		return fmt.Errorf("Unable to decode the required plugin options from yaml format for maintain phase. %s.", err)
	}
	return nil
}

