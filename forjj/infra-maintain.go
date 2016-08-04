package main

import (
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "fmt"
    "path"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "github.hpe.com/christophe-larsonneur/goforjj"
)

const (
    drivers_def_options_file = "forjj-maintain-options.yml"
    drivers_data_options_file = "forjj-options.yml"
)

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
    Type string
    Options map[string]string `,inline,omitempty`
}

func (a *Forj)SaveForjjPluginsOptions() error {
    if a.drivers_options.Drivers == nil {
        return nil
    }

    def := forjjPlugins{ make(map[string]forjjPluginsOptions) }

    for driver, opts := range a.drivers_options.Drivers {
        o := forjjPluginsOptions{ opts.driver_type, make(map[string]string) }
        for option, v := range opts.Options {
            o.Options[option] = v.Help
        }
        def.Plugins[driver] = o
    }
    gotrace.Trace("Plugin options definition file content: %#v", def)

    if err := def.Save(drivers_def_options_file) ; err != nil {
        return fmt.Errorf("Unable to write '%s'. %s", drivers_def_options_file, err)
    }

    git("add", drivers_def_options_file)

    for driver, opts := range a.drivers_options.Drivers {
        for option, v := range opts.Options {
            def.Plugins[driver].Options[option] = v.Value
        }
    }

    gotrace.Trace("Plugin options data file content: %#v", def)
    workspace_file := path.Join(a.Workspace_path, a.Workspace, drivers_data_options_file)
    if err := def.Save(workspace_file) ; err != nil {
        return fmt.Errorf("Unable to write '%s'. %s", workspace_file, err)
    }
    return nil
}

func (f *forjjPlugins)Save(file string) error {
    yaml_data, err := yaml.Marshal(f)
    if err != nil {
        return err
    }

    if err := ioutil.WriteFile(file, yaml_data , 0644 ) ; err != nil {
        return err
    }
    gotrace.Trace("Plugin file name saved: %s", file)
    return nil
}

// FIXME: Should read the repo file and the file given as parameter of forjj maintain cli.

// This functions loads the forjj plugins options definitions
// 2 files have to be loaded. The definition in forj-repo and the one given at forjj cli.
func (a *Forj)LoadForjjPluginsOptions() error {
    file := path.Clean(path.Join(a.Workspace_path, a.Workspace, a.w.Infra, drivers_data_options_file))

    yaml_data, err := ioutil.ReadFile(file)
    if err != nil {
        return fmt.Errorf("Unable to read '%s'. %s", drivers_data_options_file, err)
    }

    if err := yaml.Unmarshal(yaml_data, &a.drivers_options) ; err != nil {
        return fmt.Errorf("Unable to decode the required plugin options from yaml format for maintain phase. %s.", err)
    }
    return nil
}

func (d *DriversOptions)AddForjjPluginOptions(name string, options map[string]goforjj.PluginOption, driver_type string) {
    if d.Drivers == nil {
        d.Drivers = make(map[string]DriverOptions)
    }

    d.Drivers[name] = DriverOptions{ driver_type, options }
}

func (d *DriversOptions)GetDriversMaintainParameters(plugin_args map[string]string, action string) {
    if action != "maintain" {
        return
    }
    for _, v := range d.Drivers {
        for k, o := range v.Options {
            plugin_args[k] = o.Value
        }
    }
}
