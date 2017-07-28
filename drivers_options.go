package main

import (
	"bytes"
	"fmt"
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
	"path"
	"text/template"
	"forjj/drivers"
	"forjj/forjfile"
)

// Load driver options to a Command requested.

// Currently there is no distinction about setting different options for a specific task on the driver.
func (a *Forj) load_driver_options(instance_name string) error {
	if err := a.read_driver(instance_name); err != nil {
		return err
	}

	if a.drivers[instance_name].Plugin.Yaml.Name != "" { // if true => Driver Def loaded
		a.init_driver_flags(instance_name)
	}

	return nil
}

// TODO: Check if forjj-options, plugins runtime are valid or not.

// prepare_registered_drivers get the list of drivers identified in the Repository (Forjfile) and prepare it (Driver).
func (a *Forj) prepare_registered_drivers() error {
	for _, app := range a.f.Apps() {
		a.add_defined_driver(app)
	}
	return nil
}

// GetDriversFlags - Prepare drivers to load identified by cli App context hook.
// This function is provided as cli app object Parse hook
func (a *Forj) GetDriversFlags(o *cli.ForjObject, c *cli.ForjCli, _ interface{}) (error, bool) {
	list := a.cli.GetObjectValues(o.Name())
	// Loop on drivers to pre-initialized drivers flags.
	gotrace.Trace("Number of plugins provided from parameters: %d", len(list))
	for _, d := range list {
		if err := a.add_driver(d.GetString("driver"), d.GetString("type"), d.GetString("name"), true) ; err != nil {
			gotrace.Trace("%s", err)
			continue
		}

	}
	return nil, true
}

// Read Driver yaml document
func (a *Forj) read_driver(instance_name string) (err error) {
	var (
		yaml_data []byte
		driver    *drivers.Driver
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

	if err = driver.Plugin.PluginDefLoad(yaml_data); err != nil {
		return
	}

	// Set defaults value for undefined parameters
	var ff string
	if driver.Plugin.Yaml.CreatedFile == "" {
		ff = "." + driver.InstanceName + ".created"
		driver.ForjjFlagFile = true // Forjj will test the creation success itself, as the driver did not created it automatically.
	} else {
		ff = driver.Plugin.Yaml.CreatedFile
	}

	// Initialized defaults value from templates
	var doc bytes.Buffer

	if t, err := template.New("plugin").Parse(ff); err != nil {
		return fmt.Errorf("Unable to interpret plugin yaml definition. '/created_flag_file' has an invalid template string '%s'. %s", driver.Plugin.Yaml.CreatedFile, err)
	} else {
		t.Execute(&doc, driver.Model())
	}
	driver.FlagFile = doc.String()
	driver.Runtime = &driver.Plugin.Yaml.Runtime
	gotrace.Trace("Created flag file name Set to default for plugin instance '%s' to %s", driver.InstanceName, driver.Plugin.Yaml.CreatedFile)

	return

}

func (a *Forj) get_valid_driver_actions() (validActions []string) {
	actions := a.cli.GetAllActions()
	validActions = make([]string, 0, len(actions))
	for action_name := range actions {
		if inStringList(action_name, cr_act, upd_act, maint_act) == "" {
			validActions = append(validActions, action_name)
		}
	}
	return
}

// Initialize command drivers flags with plugin definition loaded from plugin yaml file.
func (a *Forj) init_driver_flags(instance_name string) {
	d := a.drivers[instance_name]
	service_type := d.DriverType
	opts := a.drivers_options.Drivers[instance_name]
	id := initDriverObjectFlags{
		a:             a,
		d:             d,
		instance_name: instance_name,
		d_opts:        &opts,
	}

	gotrace.Trace("Setting create/update/maintain flags from plugin type '%s' (%s)", service_type, d.Plugin.Yaml.Name)
	for command, flags := range d.Plugin.Yaml.Tasks {
		id.set_task_flags(command, flags)
	}

	// Create an object or enhance an existing one.
	// Then create the object key if needed.
	// Then add fields, define actions and create flags.
	gotrace.Trace("Setting Objects...")
	for object_name, object_det := range d.Plugin.Yaml.Objects {
		new := id.determine_object(object_name, &object_det)
		if id.obj == nil {
			continue
		}

		// Determine which actions can be configured for drivers object flags.
		id.prepare_actions_list()

		gotrace.Trace("Object '%s': Adding fields", object_name)
		// Adding fields to the object.
		for flag_name, flag_det := range object_det.Flags {
			if flag_det.FormatRegexp == "" { // Default flag regexp to eliminate cli warning.
				flag_det.FormatRegexp = ".*"
			}

			if id.add_object_fields(flag_name, &flag_det, id.validActions) {
				object_det.Flags[flag_name] = flag_det
			}
		}

		gotrace.Trace("Object '%s': Adding groups fields", object_name)
		for group_name, group_det := range object_det.Groups {
			default_actions := id.validActions
			if group_det.Actions != nil && len(group_det.Actions) > 0 {
				default_actions = group_det.Actions
				gotrace.Trace("Object '%s' - Group '%s': Default group actions defined to '%s'", default_actions)
			}

			for flag_name, flag_det := range group_det.Flags {
				if id.add_object_fields(group_name+"-"+flag_name, &flag_det, default_actions) {
					object_det.Groups[group_name].Flags[flag_name] = flag_det
				}
			}
		}

		if new {
			gotrace.Trace("Object '%s': Setting Object supported Actions...", object_name)
			// Adding Actions to the object.
			if len(id.add_object_actions()) == 0 {
				gotrace.Warning("No actions to add flags.")
				continue
			}
		} else {
			gotrace.Trace("Object '%s': Supported Actions already set - Not a new object.", object_name)
		}

		gotrace.Trace("Object '%s': Adding Object Action flags...", object_name)
		// Adding flags to object actions
		for flag_name, flag_dets := range object_det.Flags {
			id.add_object_actions_flags(flag_name, &flag_dets, id.validActions)
		}
		gotrace.Trace("Object '%s': Adding Object Action groups flags", object_name)
		for group_name, group_det := range object_det.Groups {
			default_actions := id.validActions
			if group_det.Actions != nil && len(group_det.Actions) > 0 {
				default_actions = group_det.Actions
			}
			for flag_name, flag_det := range group_det.Flags {

				id.add_object_actions_flags(group_name+"-"+flag_name, &flag_det, default_actions)
			}
		}

	}

	// TODO: Give plugin capability to manipulate new plugin object instances as list (ex: role => roles)
	// TODO: integrate new plugins objects list in create/update task
}


func (a *Forj) add_defined_driver(app *forjfile.AppStruct) error {
	if _, found := a.drivers[app.Name()]; !found {
		driver := new(drivers.Driver)
		driver.InstanceName = app.Name()
		driver.Name = app.Driver
		driver.DriverType = app.Type
		a.drivers[app.Name()] = driver
		driver.Init()
	}
	gotrace.Trace("Registered driver to load: %s - %s", app.Type, app.Name())
	return nil
}

// add_driver add a new driver to the list of drivers to load is none were already identified.
func (a *Forj) add_driver(driver, driver_type, instance string, cli_requested bool) error {
	if driver == "" || driver_type == "" {
		return fmt.Errorf("Invalid plugin definition. driver:%s, driver_type:%s", driver, driver_type)
	}
	if instance == "" {
		instance = driver
	}
	if _, found := a.drivers[instance] ; found {
		return nil
	}
	a.drivers[instance] = drivers.NewDriver(driver, driver_type, instance, true)
	gotrace.Trace("Identified driver to load: %s\n", driver_type, driver)
	return nil
}

// GetForjjFlags build the Forjj list of parameters requested by the plugin for a specific action name.
func (a *Forj) GetForjjFlags(r *goforjj.PluginReqData, d *drivers.Driver, action string) {
	if tc, found := d.Plugin.Yaml.Tasks[action]; found {
		for flag_name := range tc {
			if v, found := a.GetDriversActionsParameter(d, flag_name); found {
				r.Forj[flag_name] = v
			}
		}
	}
}

func (a *Forj)moveSecureAppData(flag_name string) {
	if v, found := a.f.GetString("settings", "", flag_name) ; found {
		a.s.SetForjValue(flag_name, v)
		a.f.Remove("settings", "", flag_name)
		gotrace.Trace("Moving secure flag data '%s' from Forjfile to creds.yaml", flag_name)
	}
	if v, error := a.cli.GetAppStringValue(flag_name) ; error == nil {
		gotrace.Trace("Setting Forjfile flag '%s' from cli", flag_name)
		a.s.SetForjValue(flag_name, v)
	}
}

func (a *Forj)copyCliData(flag_name, def_value string) {
	if v, error := a.cli.GetAppStringValue(flag_name) ; error == nil && v != "" {
		gotrace.Trace("Setting Forjfile flag '%s' from cli", flag_name)
		a.s.SetForjValue(flag_name, v)
	} else {
		if def_value != "" {
			gotrace.Trace("Setting Forjfile flag '%s' default value to '%s'", flag_name, def_value)
			a.s.SetForjValue(flag_name, def_value)
		}
	}
}

func (a *Forj)moveSecureObjectData(object_name, instance, flag_name string) {
	if v, found := a.f.Get(object_name, instance, flag_name) ; found {
		a.s.SetObjectValue(object_name, instance, flag_name, v)
		a.f.Remove(object_name, instance, flag_name)
		gotrace.Trace("Moving secure Object (%s/%s) flag data '%s' from Forjfile to creds.yaml",
			object_name, instance, flag_name)
	}
	if v, found, _, _ := a.cli.GetStringValue(object_name, instance, flag_name) ; found {
		a.s.SetObjectValue(object_name, instance, flag_name, new(goforjj.ValueStruct).Set(v))
		gotrace.Trace("Set %s/%s:%s value to Forjfile from cli.", object_name, instance, flag_name)
	}
}

// objectGetInstances returns the merge of instances of an object found in cli and Forjfile
func (a *Forj)objectGetInstances(object_name string) (ret []string) {
	cli_obj := a.cli.GetObject(object_name)
	if cli_obj == nil {
		return
	}
	instances := make(map[string]int)
	if cli_obj.IsSingle() {
		instances[object_name] = 0
	} else {
		for _, instance := range cli_obj.GetInstances() {
			instances[instance] = 0
		}
		for _, instance := range a.f.GetInstances(object_name) {
			instances[instance] = 0
		}
	}
	ret = make([]string, len(instances))
	i := 0
	for instance := range instances {
		ret[i] = instance
		i++
	}
	return
}

func (a *Forj)copyCliObjectData(object_name, instance, flag_name, def_value string) {
	if v, found, _, _ := a.cli.GetStringValue(object_name, instance, flag_name) ; found && v != "" {
		a.f.Set(object_name, instance, flag_name, v)
		gotrace.Trace("Set %s/%s:%s value to Forjfile from cli.", object_name, instance, flag_name)
	} else {
		if def_value != "" {
			a.f.SetDefault(object_name, instance, flag_name, def_value)
			gotrace.Trace("Setting Forjfile flag '%s/%s:%s' default value to '%s'",
				object_name, instance, flag_name, def_value)
		}
	}
}

// ScanAndSetObjectData scan each object defined in loaded plugins:
// - move Forjfile creds to creds.yml
// - copy cli creds data to creds.yml
// - copy cli non creds data to Forjfile
func (a *Forj)ScanAndSetObjectData() {
	for _, driver := range a.drivers {
		// Tasks flags
		for _, task := range driver.Plugin.Yaml.Tasks {
			for flag_name, flag := range task {
				if !flag.Options.Secure {
					a.copyCliData(flag_name, flag.Options.Default)
					continue
				}
				a.moveSecureAppData(flag_name)
			}
		}

		for object_name, obj := range driver.Plugin.Yaml.Objects {
			// Instances of Forj objects
			instances := a.objectGetInstances(object_name)
			for _, instance_name := range instances {
				// Do not set app object values for a driver of a different application.
				if object_name == "app" && instance_name != driver.InstanceName { continue }

				if object_name == "repo" {
					if instance_owner, is_owner := a.IsRepoManaged(driver, object_name, instance_name) ; is_owner {
						Repo := a.f.GetObjectInstance(object_name, instance_name).(*forjfile.RepoStruct)
						// Getting the owner from the upstream plugins result
						if v, found := a.drivers[instance_owner] ; found && v.Plugin.Result != nil {
							if v2, found2 := v.Plugin.Result.Data.Repos[instance_name] ; found2 {
								Repo.SetInstanceOwner(v2.Owner)
							} else {
								gotrace.Warning("Unable to set Repository Owner. Unable to find '%s' repository from " +
								"upstream '%s' driver data Result.", instance_name, instance_owner)
							}
						} else {
							gotrace.Warning("Unable to set Repository Owner. Unable to find '%s' repository from " +
								"upstream '%s' driver data Result.", instance_name, instance_owner)
						}
					} else {
						// Do not set repo values not managed by the driver
						continue
					}
				}

				// Object flags
				for flag2_name, flag2 := range obj.Flags {
					if ! flag2.Options.Secure {
						a.copyCliObjectData(object_name, instance_name, flag2_name, flag2.Options.Default)
						continue
					}
					a.moveSecureObjectData(object_name, instance_name, flag2_name)
				}
				// Object group flags
				for group_name, group := range obj.Groups {
					for flag3_name, flag3 := range group.Flags {
						if ! flag3.Options.Secure {
							a.copyCliObjectData(object_name, instance_name, group_name + "-" + flag3_name, flag3.Options.Default)
							continue
						}
						a.moveSecureObjectData(object_name, instance_name, group_name + "-" + flag3_name)
					}
				}

			}
		}
	}
}

// IsRepoManaged check is the upstream driver is the repository owner.
// It returns the repo owner declared and true if the upstream driver is that owner.
func (a *Forj) IsRepoManaged(d *drivers.Driver, object_name, instance_name string) (repo_upstream string, is_owner bool) {
	// Determine if the upstream instance is set to this instance.
	repo_upstream = a.RepoManagedBy(object_name, instance_name)
	if repo_upstream == "" {
		return
	}
	if repo_upstream != d.InstanceName {
		gotrace.Trace("Repo '%s' ignored for driver '%s'. Expect Repo to be managed by '%s'.",
			instance_name, d.InstanceName, repo_upstream)
		return
	}
	is_owner = true
	return
}

// RepoManagedBy return the upstream instance name that is identified to have the ownership
func (a *Forj) RepoManagedBy(object_name, instance_name string) (_ string) {
	// Determine if the upstream instance is set to this instance.
	if v, found := a.f.GetString(object_name, instance_name, "git-remote"); found && v != "" {
		return
	}
	if v, found := a.f.GetString(object_name, instance_name, "upstream"); found {
		return v
	}
	if v, found := a.f.GetString("settings", "default", "upstream-instance"); found {
		return v
	}
	return
}

// GetObjectsData build the list of Object required by the plugin provided from the cli flags.
func (a *Forj) GetObjectsData(r *goforjj.PluginReqData, d *drivers.Driver, action string) error {
	// Loop on each plugin object
	for object_name, Obj := range d.Plugin.Yaml.Objects {
		Obj_instances := a.f.GetInstances(object_name)
		for _, instance_name := range Obj_instances {
			// filter on current app
			if object_name == "app" && instance_name != d.InstanceName { continue }

			// Filter on repo to be supported by the driver instance.
			if object_name == "repo" {
				if _, is_owner := a.IsRepoManaged(d, object_name, instance_name) ; ! is_owner {
					continue
				}
			}

			keys := make(goforjj.InstanceKeys)

			flags := Obj.FlagsRange("setup")

			for key, flag := range flags {
				if v, found := a.GetInternalForjData(key) ; found {
					if value := new(goforjj.ValueStruct).Set(v) ; value != nil {
						keys[key] = value
					}
					continue
				}

				value := new(goforjj.ValueStruct)
				if flag.Options.Secure {
					// From creds.yml
					if v, found := a.s.Get(object_name, instance_name, key) ; !found {
						continue
					} else {
						value.Set(v)
					}
				} else {
					// From Forjfile
					if v, found := a.f.Get(object_name, instance_name, key) ; !found {
						continue
					} else {
						value.Set(v)
					}
				}
				if err := value.Evaluate(a.Model(object_name, instance_name)) ; err != nil {
					return fmt.Errorf("Unable to evaluate '%s'. %s", value.GetString(), err)
				}
				gotrace.Trace("%s/%s: Key '%s' has been set to '%s'", object_name, instance_name, key, value.GetString())
				keys[key] = value
			}
			r.AddObjectActions(object_name, instance_name, keys)
		}
	}
	return nil
}
