package drivers

import (
	"fmt"
	"forjj/utils"
	"os"
	"path"

	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

type Driver struct {
	DriverType    string                      // driver type name
	DriverVersion string                      // Driver version to use. If not set, use latest.
	InstanceName  string                      // Instance name.
	Name          string                      // Name of driver to load Yaml.Name is the real internal driver name.
	cmds          map[string]DriverCmdOptions // List of flags per commands
	Plugin        *goforjj.Plugin             // Plugin Data
	InfraRepo     bool                        // True if this driver instance is the one hosting the infra repository.
	FlagFile      string                      // Path to the predefined plugin or generic forjj plugin flag file.
	ForjjFlagFile bool                        // true if the flag_file is set by forjj.
	app_request   bool                        // true if the driver is loaded by a apps create/update/maintain task (otherwise requested by Repos or flows request.)
	Runtime       *goforjj.YamlPluginRuntime  // Reference to the plugin runtime information given by the plugin yaml file.
	// When a driver is initially loaded, it will be saved here, and used it as ref every where.
	// So we are sure that :
	// - any change in plugin is not failing a running environment.
	// - If no plugin is referenced from cli, we can start it without loading it from the plugin.yaml.
	// - We can manage plugins versions and update when needed or requested.
	DriverAPIUrl string // Recognized application API url shared between plugins
}

func NewDriver(driver, driver_type, instance string, cli_requested bool) *Driver {
	d := new(Driver)
	d.Name = driver
	d.DriverType = driver_type
	d.InstanceName = instance
	d.app_request = cli_requested
	d.Init()
	return d
}

func (d *Driver) Init() {
	if d.cmds != nil {
		return
	}
	d.cmds = map[string]DriverCmdOptions{ // List of Driver actions supported.
		"common":   {make(map[string]DriverCmdOptionFlag)},
		"create":   {make(map[string]DriverCmdOptionFlag)},
		"update":   {make(map[string]DriverCmdOptionFlag)},
		"maintain": {make(map[string]DriverCmdOptionFlag)},
	}
}

func (d *Driver) IsValidCommand(command string) bool {
	_, ok := d.cmds[command]
	return ok
}

func (d *Driver) InitCmdFlag(command, key_name, option_name string) {
	d.cmds[command].flags[key_name] = DriverCmdOptionFlag{driver_flag_name: option_name}
}

func (d *Driver) AppRequest() bool {
	return d.app_request
}

// CheckFlagBefore Check if the flag exist to avoid creating the resource a second time. It must use update instead.
func (d *Driver) CheckFlagBefore(instance, action string) error {
	flag_file := path.Join("apps", d.DriverType, d.FlagFile)

	if d.ForjjFlagFile { // Default setup made by Forjj
		if _, err := os.Stat(flag_file); err == nil {
			if action == "create" {
				return fmt.Errorf("The driver instance '%s' has already created the resources. Use 'Update' to update it, and maintain to instanciate it as soon as your infra repo flow is completed.", instance)
			}
		} else {
			gotrace.Trace("Flag file '%s' NOT found. `forjj create` is authorized.", flag_file)
		}
	} else {
		if _, err := os.Stat(flag_file); err != nil {
			// if an update is requested on the driver host the infra, then we will need to go further to restore the workspace. No error in that case.
			if action == "update" && !d.InfraRepo {
				return fmt.Errorf("The driver instance '%s' do not have the resource requested. Use 'Create' to create it, and maintain to instanciate it as soon as your infra repo flow is completed.", instance)
			}
		} else {
			gotrace.Trace("Flag file '%s' found. `forjj create` is NOT authorized. It must be update or maintain.", flag_file)
		}
	}
	return nil
}

func (d *Driver) CheckFlagAfter() (err error) {
	flag_file := path.Join("apps", d.DriverType, d.FlagFile)

	// Check the flag file
	if _, err = os.Stat(flag_file); err == nil {
		return err
	}

	gotrace.Warning("Driver '%s' has not created the expected flag file (%s). Probably a driver bug. Contact the plugin maintainer to fix it.", d.Name, flag_file)

	// Create a forjj flag file instead.
	if err = utils.Touch(flag_file); err != nil {
		return err
	}
	gotrace.Trace("Forjj has flagged (%s) for driver '%s(%s)'", flag_file, d.Name, d.DriverType)

	return
}

// HasNoFiles Return True if no file sis registered in the driver response.
func (d *Driver) HasNoFiles() bool {
	return (len(d.Plugin.Result.Data.Files) == 0)
}
