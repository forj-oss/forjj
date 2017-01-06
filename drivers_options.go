package main

import (
	"bytes"
	"fmt"
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/goforjj"
	"github.com/forj-oss/forjj-modules/trace"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"text/template"
)

// Load driver options to a Command requested.

// Currently there is no distinction about setting different options for a specific task on the driver.
func (a *Forj) load_driver_options(instance_name string) error {
	if err := a.read_driver(instance_name); err != nil {
		return err
	}

	if a.drivers[instance_name].plugin.Yaml.Name != "" { // if true => Driver Def loaded
		a.init_driver_flags(instance_name)
	}

	return nil
}

func (d *Driver) Model() (m *DriverModel) {
	m = &DriverModel{
		InstanceName: d.InstanceName,
		Name:         d.Name,
	}
	return
}

// TODO: Check if forjj-options, plugins runtime are valid or not.

func (a *Forj) load_missing_drivers() error {
	gotrace.Trace("Number of registered instances %d", len(a.o.Drivers))
	gotrace.Trace("Number of loaded instances %d", len(a.drivers))
	for instance, d := range a.o.Drivers {
		if _, found := a.drivers[instance]; !found {
			gotrace.Trace("Loading missing instance %s", instance)
			a.drivers[instance] = d
			d.cmds = map[string]DriverCmdOptions{ // List of Driver actions supported.
				"common":   {make(map[string]DriverCmdOptionFlag)},
				"create":   {make(map[string]DriverCmdOptionFlag)},
				"update":   {make(map[string]DriverCmdOptionFlag)},
				"maintain": {make(map[string]DriverCmdOptionFlag)},
			}

			gotrace.Trace("Loading '%s'", instance)
			if err := a.load_driver_options(instance); err != nil {
				log.Printf("Unable to load plugin information for instance '%s'. %s", instance, err)
				continue
			}
			/*            if err := d.plugin.PluginLoadFrom(instance, d.Runtime) ; err != nil {
			              log.Printf("Unable to load Runtime information from forjj-options for instance '%s'. Forjj may not work properly. You can fix it with 'forjj update --apps %s:%s:%s'. %s", instance, d.DriverType, d.Name, d.InstanceName, err)
			          }*/
			d.plugin.PluginSetWorkspace(a.w.Path())
			d.plugin.PluginSetSource(path.Join(a.w.Path(), a.w.Infra.Name, "apps", d.DriverType))
			d.plugin.PluginSocketPath(path.Join(a.w.Path(), "lib"))
		}
	}
	return nil
}

// Read Driver yaml document
func (a *Forj) read_driver(instance_name string) (err error) {
	var (
		yaml_data []byte
		driver    *Driver
	)
	if d, ok := a.drivers[instance_name]; ok {
		driver = d
	}

	if driver.Name == "" {
		return
	}

	ContribRepoUri := *a.ContribRepo_uri
	ContribRepoUri.Path = path.Join(ContribRepoUri.Path, driver.DriverType, driver.Name, driver.Name+".yaml")

	if yaml_data, err = read_document_from(&ContribRepoUri); err != nil {
		return
	}

	if err = driver.plugin.PluginDefLoad(yaml_data); err != nil {
		return
	}

	// Set defaults value for undefined parameters
	var ff string
	if driver.plugin.Yaml.CreatedFile == "" {
		ff = "." + driver.InstanceName + ".created"
		driver.ForjjFlagFile = true // Forjj will test the creation success itself, as the driver did not created it automatically.
	} else {
		ff = driver.plugin.Yaml.CreatedFile
	}

	// Initialized defaults value from templates
	var doc bytes.Buffer

	if t, err := template.New("plugin").Parse(ff); err != nil {
		return fmt.Errorf("Unable to interpret plugin yaml definition. '/created_flag_file' has an invalid template string '%s'. %s", driver.plugin.Yaml.CreatedFile, err)
	} else {
		t.Execute(&doc, driver.Model())
	}
	driver.FlagFile = doc.String()
	driver.Runtime = &driver.plugin.Yaml.Runtime
	gotrace.Trace("Created flag file name Set to default for plugin instance '%s' to %s", driver.InstanceName, driver.plugin.Yaml.CreatedFile)

	return

}

// Initialize command drivers flags with plugin definition loaded from plugin yaml file.
func (a *Forj) init_driver_flags(instance_name string) {
	d := a.drivers[instance_name]
	service_type := d.DriverType
	commands := d.plugin.Yaml.Actions
	d_opts := a.drivers_options.Drivers[instance_name]

	gotrace.Trace("Setting flags from plugin type '%s' (%s)", service_type, d.plugin.Yaml.Name)
	for command, def := range commands {
		if _, ok := a.drivers[instance_name].cmds[command]; !ok {
			fmt.Printf("FORJJ Driver '%s': Invalid tag '%s'. valid one are 'common', 'create', 'update', 'maintain'. Ignored.", service_type, command)
		}

		// Sort Flags for readability:
		keys := make([]string, 0, len(def.Flags))

		for k := range def.Flags {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		search_re, _ := regexp.Compile("^(.*[_-])?(" + d.plugin.Yaml.Name + ")([_-].*)?$")
		for _, option_name := range keys {
			flag_options := def.Flags[option_name]

			// drivers flags starting with --forjj are a way to communicate some forjj internal data to the driver.
			// They are not in the list of possible drivers options from the cli.
			if ok, _ := regexp.MatchString("forjj-.*", option_name); ok {
				d.cmds[command].flags[option_name] = DriverCmdOptionFlag{driver_flag_name: option_name} // No value by default. Will be set later after complete parse.
				continue
			}

			forjj_option_name := SetAppropriateflagName(option_name, instance_name, search_re)
			flag_opts := d_opts.set_flag_options(option_name, &flag_options)
			if command == "common" {
				// loop on create/update/maintain to create flag on each command
				gotrace.Trace("Create common flags '%s' to each commands...", forjj_option_name)
				for _, cmd := range []string{"create", "update", "maintain"} {
					d.init_driver_flags_for(a, option_name, cmd, forjj_option_name, flag_options.Help, flag_opts)
				}
			} else {
				d.init_driver_flags_for(a, option_name, command, forjj_option_name, flag_options.Help, flag_opts)
			}
		}
	}
}

// Set options on a new flag created.
//
// It currently assigns defaults or required.
//
func (d *DriverOptions) set_flag_options(option_name string, params *goforjj.YamlFlagsOptions) (opts *cli.ForjOpts) {
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
	return
}

// Create the flag to a kingpin Command. (create/update/maintain)
func (d *Driver) init_driver_flags_for(a *Forj, option_name, command, forjj_option_name, forjj_option_help string, opts *cli.ForjOpts) {
	// No value by default. Will be set later after complete parse.
	d.cmds[command].flags[forjj_option_name] = DriverCmdOptionFlag{driver_flag_name: option_name}

	// Create flag 'option_name' on kingpin cmd or app
	if forjj_option_name != option_name {
		gotrace.Trace("Set action '%s' flag for '%s(%s)'", command, forjj_option_name, option_name)
	} else {
		gotrace.Trace("Set action '%s' flag for '%s'", command, forjj_option_name)
	}
	a.cli.OnActions(command).AddFlag(cli.String, forjj_option_name, forjj_option_help, opts)
	return
}

// SetAppropriateflagName Define the Forjj flag name for the plugin selected.
// If instanceName is not equal to service_type, then
// any flag without <service_type> in flag name will returned with <instance_name>-<FlagName>
// Any flag having at least .*[_-]<service_type> or <service_type>[_-].*or both, <service_type is replaced by <instance_name>
func SetAppropriateflagName(flag_name, instance_name string, search_re *regexp.Regexp) string {

	res := search_re.FindStringSubmatch(flag_name)
	if res == nil {

		return instance_name + "-" + flag_name
	}

	return search_re.ReplaceAllString(flag_name, "${1}"+instance_name+"${3}")
}

// GetDriversFlags - cli App context hook. Load drivers requested (app object)
// This function is provided as cli app object Parse hook
func (a *Forj) GetDriversFlags(o *cli.ForjObject, c *cli.ForjCli, d interface{}) error {
	//gotrace.Trace("c: %s", c)

	list := a.cli.GetObjectValues(o.Name())
	// Loop on drivers to pre-initialized drivers flags.
	gotrace.Trace("Number of plugins provided from parameters: %d", len(list))
	for _, d := range list {
		driver := d.GetString("driver")
		driver_type := d.GetString("driver_type")
		instance := d.GetString("instance")
		if instance == "" {
			instance = driver
		}
		if driver == "" || driver_type == "" {
			gotrace.Trace("Invalid plugin definition. driver:%s, driver_type:%s", driver, driver_type)
			continue
		}

		a.drivers[instance] = &Driver{
			Name:         driver,
			DriverType:   driver_type,
			InstanceName: instance,
			app_request:  true,
			cmds: map[string]DriverCmdOptions{ // List of Driver actions supported.
				"common":   {make(map[string]DriverCmdOptionFlag)},
				"create":   {make(map[string]DriverCmdOptionFlag)},
				"update":   {make(map[string]DriverCmdOptionFlag)},
				"maintain": {make(map[string]DriverCmdOptionFlag)},
			},
		}
		gotrace.Trace("Selected '%s' app driver: %s\n", driver_type, driver)

		if err := a.load_driver_options(instance); err != nil {
			fmt.Printf("Error: %#v\n", err)
			os.Exit(1)
		}
	}

	// Automatically load all other drivers not requested by --apps but listed in forjj-options.yaml.
	// Those drivers are all used by all services that forjj should manage.
	a.load_missing_drivers()
	return nil
}
