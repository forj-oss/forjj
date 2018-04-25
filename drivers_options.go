package main

import (
	"bytes"
	"fmt"
	"forjj/drivers"
	"forjj/forjfile"
	"forjj/scandrivers"
	"forjj/utils"
	"path"
	"strings"
	"text/template"

	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

// Load driver options to a Command requested.

// Currently there is no distinction about setting different options for a specific task on the driver.
func (a *Forj) load_driver_options(instance_name string) error {
	if err := a.read_driver(instance_name); err != nil {
		return err
	}

	if a.cli.GetCurrentCommand()[0].FullCommand() == val_act {
		return nil // Do not set plugin flags in validate mode.
	}
	if a.drivers[instance_name].Plugin.Yaml.Name != "" { // if true => Driver Def loaded
		if err := a.init_driver_flags(instance_name); err != nil {
			return err
		}
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
		if err := a.add_driver(d.GetString("driver"), d.GetString("type"), d.GetString("name"), true); err != nil {
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

	repos := []string{"forjj-" + driver.Name, driver.Name, "forjj-contribs"}
	reposSubPaths := []string{"", "", path.Join(driver.DriverType, driver.Name)}
	if yaml_data, err = utils.ReadDocumentFrom(a.ContribRepoURIs, repos, reposSubPaths, driver.Name+".yaml"); err != nil {
		return
	}

	if err = driver.Plugin.PluginDefLoad(yaml_data); err != nil {
		return
	}

	// Set defaults value for undefined parameters
	var ff string
	if driver.Plugin.Yaml.CreatedFile == "" {
		ff = "{{ .InstanceName }}/{{.Name}}.yaml" // Default Flag file setting.
		driver.ForjjFlagFile = true               // Forjj will test the creation success itself, as the driver did not created it automatically.
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

// TODO: Replace this split by cli function to list valid actions from cli definition (cli.NewAction)
// This will avoid hard coded list here.

// get_valid_driver_actions split al cli actions in
// Object actions, like add/remove and
// Command actions, like create/maintain
// It returns 2 list of valid kingpin Object/Command actions
func (a *Forj) get_valid_driver_actions() (validObjectActions, validCommandActions []string) {
	actions := a.cli.GetAllActions()
	validObjectActions = make([]string, 0, len(actions))
	validCommandActions = make([]string, 0, len(actions))
	for action_name := range actions {
		if utils.InStringList(action_name, cr_act, upd_act, maint_act) == "" {
			validObjectActions = append(validObjectActions, action_name)
		} else {
			validCommandActions = append(validCommandActions, action_name)
		}
	}
	return
}

func (a *Forj) init_driver_set_cli_app_instance(instance_name string) error {
	app_obj := a.cli.GetObject(app)

	if app_obj == nil {
		return fmt.Errorf("Unable to find internal 'app' object.")
	}
	app_obj.AddInstances(instance_name)
	return nil
}

// Initialize kingpin/cli command drivers flags with plugin definition loaded from plugin yaml file.
func (a *Forj) init_driver_flags(instance_name string) error {
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

	// Add Application(app) instance name to cli, if not set itself by cli input.
	if err := a.init_driver_set_cli_app_instance(instance_name); err != nil {
		return fmt.Errorf("Unable to initialize driver instance '%s'. %s", instance_name, err)
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

			id.add_object_field_to_cmds(flag_name, &flag_det)
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
	return nil
}

func (a *Forj) add_defined_driver(app *forjfile.AppStruct) error {
	if _, found := a.drivers[app.Name()]; !found {
		driver := new(drivers.Driver)
		driver.InstanceName = app.Name()
		driver.Name = app.Driver
		driver.DriverType = app.Type
		driver.DriverVersion = app.Version
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
	if _, found := a.drivers[instance]; found {
		return nil
	}
	a.drivers[instance] = drivers.NewDriver(driver, driver_type, instance, true)
	gotrace.Trace("Identified driver to load: %s\n", driver_type, driver)
	return nil
}

// GetForjjFlags build the Forjj list of parameters requested by the plugin for a specific action name.
func (a *Forj) GetForjjFlags(r *goforjj.PluginReqData, d *drivers.Driver, action string) {
	var action_data string

	if action == maint_act && a.from_create {
		gotrace.Trace("Getting flags from create action instead of maintain, as started from create.")
		action_data = cr_act
	} else {
		action_data = action
	}
	if tc, found := d.Plugin.Yaml.Tasks[action]; found {
		for flag_name := range tc {
			if v, found := a.GetDriversActionsParameter(d, flag_name, action_data); found {
				r.Forj[flag_name] = v
			}
		}
	}
}

func (a *Forj) moveSecureAppData(ffd *forjfile.DeployForgeYaml, flag_name string, missing_required bool) error {
	if v, found := ffd.GetString("settings", "", flag_name); found {
		a.s.SetForjValue(a.scanDeploy, flag_name, v)
		ffd.Remove("settings", "", flag_name)
		gotrace.Trace("Moving secure flag data '%s' from Forjfile to creds.yaml", flag_name)
		return nil
	}
	if v, error := a.cli.GetAppStringValue(flag_name); error == nil {
		gotrace.Trace("Setting Forjfile flag '%s' from cli", flag_name)
		a.s.SetForjValue(a.scanDeploy, flag_name, v)
		return nil
	}
	if missing_required {
		return fmt.Errorf("Missing required setting flag '%s' value.", flag_name)
	}
	return nil
}

func (a *Forj) copyCliData(flag_name, def_value string) {
	if v, error := a.cli.GetAppStringValue(flag_name); error == nil && v != "" {
		gotrace.Trace("Setting Forjfile flag '%s' from cli", flag_name)
		a.s.SetForjValue(a.scanDeploy, flag_name, v)
	} else {
		if def_value != "" {
			gotrace.Trace("Setting Forjfile flag '%s' default value to '%s'", flag_name, def_value)
			a.s.SetForjValue(a.scanDeploy, flag_name, def_value)
		}
	}
}

func (a *Forj) moveSecureObjectData(ffd *forjfile.DeployForgeYaml, object_name, instance, flag_name string, missing_required bool) error {
	if v, found := ffd.Get(object_name, instance, "secret_"+flag_name); found {
		// each key can have a secret_<key> value defined, stored in secret and can be refered in the Forjfile
		// with {{ Current.Creds.<flag_name> }}
		a.s.SetObjectValue(a.scanDeploy, object_name, instance, flag_name, v)
		ffd.Remove(object_name, instance, flag_name)

		gotrace.Trace("Removing and setting secure Object (%s/%s) flag data '%s' from Forjfile to creds.yaml",
			object_name, instance, "secret_"+flag_name)
	}
	if v, found := ffd.Get(object_name, instance, flag_name); found {
		a.s.SetObjectValue(a.scanDeploy, object_name, instance, flag_name, v)
		// When no template value is set in Forjfile flag value, (default case in next code line)
		// forjj will consider this string '{{ .Current.Creds.<flag_name> }}' as way to extract it
		// The Forjfile can define that flag value to a different template. A simple string is not
		// permitted for such secured data.
		ffd.Remove(object_name, instance, flag_name) // To let Forjj get default way.
		gotrace.Trace("Moving secure Object (%s/%s) flag data '%s' from Forjfile to creds.yaml",
			object_name, instance, flag_name)
		return nil
	}
	if v, found, _, _ := a.cli.GetStringValue(object_name, instance, flag_name); found {
		a.s.SetObjectValue(a.scanDeploy, object_name, instance, flag_name, new(goforjj.ValueStruct).Set(v))
		gotrace.Trace("Set %s/%s:%s value to Forjfile from cli.", object_name, instance, flag_name)
		return nil
	}
	if _, found3 := a.s.Get(object_name, instance, flag_name); !found3 && missing_required {
		return fmt.Errorf("Missing required %s %s flag '%s' value.", object_name, instance, flag_name)
	}
	return nil
}

func (a *Forj) setSecureObjectData(ffd *forjfile.DeployForgeYaml, object_name, instance, flag_name string, missing_required bool) error {
	if v, found := ffd.Get(object_name, instance, "secret-"+flag_name); found {
		// each key can have a secret_<key> value defined, stored in secret and can be referred in the Forjfile
		// with {{ Current.Creds.<flag_name> }}
		a.s.SetObjectValue(a.scanDeploy, object_name, instance, flag_name, v)
		ffd.Remove(object_name, instance, "secret-"+flag_name)
		gotrace.Trace("Moving secret Object (%s/%s) flag data '%s' from Forjfile to creds.yaml",
			object_name, instance, "secret-"+flag_name)
	}
	return nil
}

// objectGetInstances returns the merge of instances of an object found in cli and Forjfile
func (a *Forj) objectGetInstances(object_name string) (ret []string) {
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

func (a *Forj) copyCliObjectData(ffd *forjfile.DeployForgeYaml, object_name, instance, flag_name, def_value string) {
	if v, found, _, _ := a.cli.GetStringValue(object_name, instance, flag_name); found && v != "" {
		a.f.Set(object_name, instance, flag_name, v)
		gotrace.Trace("Set %s/%s:%s value to Forjfile from cli.", object_name, instance, flag_name)
	} else {
		if def_value != "" {
			ffd.SetDefault(object_name, instance, flag_name, def_value)
			gotrace.Trace("Setting Forjfile flag '%s/%s:%s' default value to '%s'",
				object_name, instance, flag_name, def_value)
		}
	}
}

// scanCreds scan each tasks/objects flags defined in loaded plugins to:
// - move Forjfile creds to creds.yml
// - copy cli creds data to creds.yml
//
// Used by create and update
func (a *Forj) scanCreds(ffd *forjfile.DeployForgeYaml, deploy string, missing bool) error {
	s := scandrivers.NewScanDrivers(ffd, a.drivers)

	s.SetScanTaskFlagsFunc(
		func(name string, flag goforjj.YamlFlag) error {
			if flag.Options.Secure {
				if err := a.moveSecureAppData(ffd, name, missing && flag.Options.Required); err != nil {
					return err
				}
			}
			return nil
		})

	s.SetScanObjFlag(
		func(objectName, instanceName, flagPrefix, flagName string, flag goforjj.YamlFlag) (err error) {
			if flag.Options.Secure {
				if err = a.moveSecureObjectData(ffd, objectName, instanceName, flagPrefix+flagName, missing && flag.Options.Required); err != nil {
					return err
				}
			}
			return nil
		})

	s.SetScanGetObjInstFunc(
		func(objectName string) []string {
			return ffd.GetInstances(objectName)
		})

	return s.DoScanDriversObject(deploy)
}

func (a *Forj) scanAndSetDefaults(ffd *forjfile.DeployForgeYaml, deploy string) error {
	return nil
}

// ScanAndSetObjectData scan each object defined in loaded plugins:
// - move Forjfile creds to creds.yml
// - copy cli creds data to creds.yml
// - copy cli non creds data to Forjfile
func (a *Forj) ScanAndSetObjectData(ffd *forjfile.DeployForgeYaml, deploy string, missing bool) error {
	if ffd == nil { // No Forjfile loaded.
		return nil
	}
	a.scanDeploy = deploy
	for _, driver := range a.drivers {
		// Tasks flags
		for _, task := range driver.Plugin.Yaml.Tasks {
			for flag_name, flag := range task {
				if !flag.Options.Secure {
					a.copyCliData(flag_name, flag.Options.Default)
					continue
				}
				if err := a.moveSecureAppData(ffd, flag_name, missing && flag.Options.Required); err != nil {
					return err
				}
			}
		}

		for object_name, obj := range driver.Plugin.Yaml.Objects {
			// Instances of Forj objects
			instances := a.objectGetInstances(object_name)
			for _, instance_name := range instances {
				// Do not set app object values for a driver of a different application.
				if object_name == "app" && instance_name != driver.InstanceName {
					continue
				}

				// Object flags
				if err := a.DispatchObjectFlags(ffd, object_name, instance_name, "", missing, obj.Flags); err != nil {
					return err
				}
				// Object group flags
				for group_name, group := range obj.Groups {
					if err := a.DispatchObjectFlags(ffd, object_name, instance_name, group_name+"-", missing, group.Flags); err != nil {
						return err
					}
				}

			}
		}
	}
	return nil
}

// DispatchObjectFlags is dispatching Forjfile template data between Forjfile and creds
// All plugin defined flags set with secret ON, are moving to creds
// All plugin undefined flags named with "secret_" as prefix are considered as required to be moved to
// creds
//
// The secret transfered flag value is moved to creds functions
// while in Forjfile the moved value is set to {{ .creds.<flag_name> }}
// a golang template is then used for Forfile to get the data from the default credential structure.
func (a *Forj) DispatchObjectFlags(ffd *forjfile.DeployForgeYaml, object_name, instance_name, flag_prefix string, missing bool, flags map[string]goforjj.YamlFlag) (err error) {
	for flag_name, flag := range flags {
		// treat 'secret_' flag type.
		if err = a.setSecureObjectData(ffd, object_name, instance_name, flag_prefix+flag_name,
			missing && flag.Options.Required); err != nil {
			return err
		}
		if !flag.Options.Secure {
			// Forjfile flag already loaded. We can update it from cli or default otherwise
			a.copyCliObjectData(ffd, object_name, instance_name, flag_prefix+flag_name, flag.Options.Default)
			continue
		}
		if err = a.moveSecureObjectData(ffd, object_name, instance_name, flag_prefix+flag_name,
			missing && flag.Options.Required); err != nil {
			return err
		}
	}
	return
}

// IsRepoManaged check if the upstream driver is the repository owner.
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
	if v, found := a.f.GetString(object_name, instance_name, "apps:upstream"); found {
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
			if object_name == "app" && instance_name != d.InstanceName {
				continue
			}

			// Filter on repo to be supported by the driver instance.
			if object_name == "repo" {
				if _, is_owner := a.IsRepoManaged(d, object_name, instance_name); !is_owner {
					continue
				}
			}

			keys := make(goforjj.InstanceKeys)

			flags := Obj.FlagsRange("setup")

			for key, flag := range flags {
				if v, found := a.GetInternalForjData(key); found {
					if value := new(goforjj.ValueStruct).Set(v); value != nil {
						keys[key] = value
					}
					continue
				}

				value := new(goforjj.ValueStruct)
				if flag.Options.Secure {
					// From creds.yml
					def_value := "{{ (index .Current.Creds \"" + key + "\").GetString }}"
					if v, found := a.f.Get(object_name, instance_name, key); found {
						if s := v.GetString(); strings.HasPrefix("{{", s) {
							def_value = s
						}
					}
					value.Set(def_value)
				} else {
					// From Forjfile
					if v, found := a.f.Get(object_name, instance_name, key); !found {
						continue
					} else {
						value.Set(v)
					}
				}
				if err := value.Evaluate(a.Model(object_name, instance_name, key)); err != nil {
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

// AddReqDeployment create a new deployment-env key in the forj-settings section of a plugin payload.
func (a *Forj) AddReqDeployment(req *goforjj.PluginReqData) (err error) {
	if deploy := a.f.GetDeployment(); deploy != "" {
		req.Forj["deployment-env"] = deploy
		if deployObj, found := a.f.GetADeployment(deploy); found {
			req.Forj["deployment-type"] = deployObj.Type
		} else {
			return fmt.Errorf("Internal error! Deploy object '%s' not found in deployments", deploy)
		}
		return
	}
	return fmt.Errorf("Cannot deploy to an unknown environment. Your Forjfile is missing `forj-settings/deploy-environment`")
}

func (a *Forj) DefineMissingDeployRepositories(warning bool) (err error) {
	// Define Deployments repositories and
	for deployName := range a.f.GetDeployments() {
		// Ensure deploy repo is identified.
		repoName := a.w.Organization + "-" + deployName
		if r, found := a.f.GetRepo(repoName); found {
			if r.IsInfra() {
				return fmt.Errorf("Deployment repository can't be your Infra Source repository '%s'. Fix your Forjfile", repoName)
			}
			continue
		}
		repo := a.f.DeployForjfile().NewRepoStruct(repoName)
		repo.Set(forjfile.FieldRepoTitle, strings.Title(deployName)+" deployment code generated by Forjj from "+a.f.GetInfraName()+".")
		repo.Set("issue_tracker", "false") // if the upstream detect this parameter disable the issue tracker.
		repo.Set(forjfile.FieldRepoFlow, "default")
		repo.Set(forjfile.FieldRepoDeployName, deployName)
		// TODO: Set defaults and internals automatically.
		repo.Register()
		if v, found := a.f.GetUpstreamApps(); found && len(v) != 1 {
			repo.Set(forjfile.FieldRepoUpstream, a.w.Instance)
		}
		if warning {
			gotrace.Warning("Deployment Repository '%s' added automatically. We suggest you to declare it in your main Forjfile.", repoName)
		} else {
			gotrace.Info("Deployment Repository '%s' added to your main Forjfile.", repoName)
		}

	}
	return
}
