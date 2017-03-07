package drivers

import (
	"github.com/forj-oss/goforjj"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/forjj-modules/cli"
)

type DriversOptions struct {
	Drivers map[string]DriverOptions // List of options for each drivers
}

// DriverOptions: List of maintain drivers options required by each plugin.
type DriverOptions struct {
	Driver_type string
	Options     map[string]goforjj.PluginOption // List of options with helps given by the plugin through create/update phase.
}

type DriverCmdOptions struct {
	flags map[string]DriverCmdOptionFlag // list of flags values
										 //    args  map[string]string // list of args values
}

type DriverCmdOptionFlag struct {
	driver_flag_name string
	value            string
}

func (d *DriversOptions) AddForjjPluginOptions(name string, options map[string]goforjj.PluginOption, driver_type string) {
	if d.Drivers == nil {
		d.Drivers = make(map[string]DriverOptions)
	}

	d.Drivers[name] = DriverOptions{driver_type, options}
}



// GetDriversMaintainParameters Used by api service or cli driver to add options values requested by the driver from
// creds & def.
// Currently this function add all values for all drivers to the args. So, we need to:
// TODO: Revisit how args are built for drivers, when flow will be introduced.
// We may need to select which one is required for each driver to implement the flow. TBD
func (d *DriversOptions) GetDriversMaintainParameters(plugin_args map[string]string, action string) {
	for n, v := range d.Drivers {
		for k, o := range v.Options {
			if o.Value == "" {
				gotrace.Trace("Instance '%s' parameter '%s' has no value.", n, k)
			}
			plugin_args[k] = o.Value
		}
	}
}

// Set options on a new flag created.
//
// It currently assigns defaults or required.
//
func (d *DriverOptions) SetFlagOptions(option_name string, params *goforjj.YamlFlagOptions) (opts *cli.ForjOpts) {
	if params == nil {
		return
	}

	var preloaded_data bool
	opts = cli.Opts()

	if d != nil {
		if option_value, found := d.Options[option_name]; found && option_value.Value != "" {
			// Do not set flag in any case as required or with default, if a value has been set in the driver loaded options (creds-forjj.yml)
			preloaded_data = true
			if params.Secure {
				// We do not set a secure data as default in kingpin default flags to avoid displaying them from forjj help.
				gotrace.Trace("Option value found for '%s' : -- set as hidden default value. --", option_name)
				// The data will be retrieved by
			} else {
				gotrace.Trace("Option value found for '%s' : %s -- Default value. --", option_name, option_value.Value)
				// But here, we can show through kingpin default what was loaded.
				opts.Default(option_value.Value)
			}
		}
	}

	if !preloaded_data {
		// No preloaded data from forjj-creds.yaml (or equivalent files) -- Normal plugin driver set up
		if params.Required {
			opts.Required()
		}
		if params.Default != "" {
			opts.Default(params.Default)
		}
	}

	if params.Envar != "" {
		opts.Envar(params.Envar)
	}
	return
}
