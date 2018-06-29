package main

import (
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
	"log"
	"regexp"
	"sort"
	"forjj/drivers"
	"forjj/utils"
)

// initDriverObjectFlags internally used by init_driver_flags()
type initDriverObjectFlags struct {
	// Initialized by init_driver_flags()
	d             *drivers.Driver
	a             *Forj
	instance_name string
	d_opts        *drivers.DriverOptions

	// Initialized by determine_object()
	object_name string
	object_det  *goforjj.YamlObject
	obj         *cli.ForjObject

	// Initialized by prepare_actions_list()
	validActions        []string
	validCommandActions []string
	allActions          bool
	defineActions       map[string]bool

	// Used by add_object_actions_flags()
	object_instance_name string
}

// set_task_flags read the driver in task_flags/create and get the list of flags to create as cli action flag.
func (id *initDriverObjectFlags) set_task_flags(command string, flags map[string]goforjj.YamlFlag) {
	service_type := id.d.DriverType

	if ok := id.a.drivers[id.instance_name].IsValidCommand(command); !ok {
		log.Printf("FORJJ Driver '%s': Invalid tag '%s'. valid one are 'common', 'create', 'update', 'maintain'. Ignored.",
			service_type, command)
	}

	// Sort Flags for readability:
	keys := make([]string, 0, len(flags))
	for k := range flags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	no_maintain := false
	if command == maint_act {
		no_maintain, _ = id.a.cli.GetActionBoolValue(cr_act, no_maintain_f)
	}

	for _, option_name := range keys {
		flag_options := flags[option_name]

		// drivers flags starting with --forjj are a way to communicate some forjj internal data to the driver.
		// They are not in the list of possible drivers options from the cli.
		if ok, _ := regexp.MatchString("forjj-.*", option_name); ok {
			// No value by default. Will be set later after complete parse.
			id.d.InitCmdFlag(command, option_name, option_name)
			gotrace.Trace("'%s' action Flag '%s' is an internal forjj flag request. Not added to kingpin.",
				command, option_name)
			continue
		}

		forjj_option_name := id.instance_name + "-" + option_name
		flag_opts := id.d_opts.SetFlagOptions(option_name, &flag_options.Options, id.task_has_value)
		if command == common_acts {
			// loop on create/update/maintain to create flag on each command
			gotrace.Trace("Create common flags '%s' to App layer.", forjj_option_name)
			id.a.init_driver_flags_for(id.d, option_name, "", forjj_option_name, flag_options.Help, flag_opts)
		} else {
			id.a.init_driver_flags_for(id.d, option_name, command, forjj_option_name, flag_options.Help, flag_opts)
			if  command == maint_act && !no_maintain {
				gotrace.Trace("Adding `maintain` flag '%s' to `create` action.", option_name)
				id.a.init_driver_flags_for(id.d, option_name, cr_act, forjj_option_name, flag_options.Help, flag_opts)
			} else {
				gotrace.Trace("`maintain` flag '%s' NOT added to `create` action. --no-maintain is true.", option_name)
			}
		}
	}
}

// task_has_value will determine which default value to add to a cli flag.
// It is called in the context of a plugin `task_flags`
// It is searching in creds (creds.YamlSecure) and in Forjfile (forjfile.Forge)
//
// If a value is found in both creds and Forjfile, creds is chosen.
func (id *initDriverObjectFlags) task_has_value(flag string) (value string, found bool) {
	value, found, _ = id.a.s.GetString(id.object_name, id.object_instance_name, flag)
	if found { // Any credential data are simply ignored
		return
	}
	return id.a.f.GetString("settings", "", flag)
}

// determine_object identify an existing object or create a new one with a key if not single.
func (id *initDriverObjectFlags) determine_object(object_name string, object_det *goforjj.YamlObject) (new bool) {
	id.object_det = object_det
	id.object_name = object_name
	flag_key := id.object_det.Identified_by_flag

	if o := id.a.cli.GetObject(object_name); o != nil {
		if o.HasRole() == "internal" {
			gotrace.Trace("'%s' object definition is invalid. This is an internal forjj object. Ignored.", object_name)
			id.obj = nil
			return
		}
		id.obj = o
		gotrace.Trace("Updating object '%s'", object_name)
	} else {
		// New Object and get the key.
		new = true
		id.obj = id.a.cli.NewObject(object_name, id.object_det.Help, "")
		if flag_key == "" {
			id.obj.Single()
			gotrace.Trace("New single object '%s'", object_name)
		} else {
			if v, found := id.object_det.Flags[flag_key]; !found {
				gotrace.Trace("Unable to create the object '%s' identified by '%s'. '%s' is not defined.",
					object_name, flag_key, flag_key)
			} else {
				flag_opts := id.d_opts.SetFlagOptions(flag_key, &v.Options, id.d_opts.HasValue)
				if v.FormatRegexp == "" {
					id.obj.AddKey(cli.String, flag_key, v.Help, ".*", flag_opts)
				} else {
					id.obj.AddKey(cli.String, flag_key, v.Help, v.FormatRegexp, flag_opts)
				}
			}
			gotrace.Trace("New object '%s' with key '%s'", object_name, flag_key)
		}
	}
	return
}

// determine_object_instances will update object cli with the list of instances found in Forjfile.
func (id *initDriverObjectFlags) determine_object_instances(object_name string) {

}

// prepare_actions_list set validActions/defineActions
func (id *initDriverObjectFlags) prepare_actions_list() {
	// Determine which actions can be configured for drivers object flags.
	id.validActions, id.validCommandActions = id.a.get_valid_driver_actions()
	id.defineActions = make(map[string]bool)
	for _, action_name := range id.validActions {
		id.defineActions[action_name] = false
	}
}

// is_object_scope determine if the field scope should be global or not
// It decides in the following order:
// - at yaml flag level (/objects/<object_name>/flags/<flag_name>)
// - at yaml object level (/objects/<object_name>)
// - from object role
func (id *initDriverObjectFlags) is_field_object_scope(flag_det *goforjj.YamlFlag) (bool) {
	if v := flag_det.FieldScope ; v != "" {
		return (v == "object")
	}
	if v := id.object_det.FieldsScope ; v != "" {
		return (v == "object")
	}
	return (id.obj.HasRole() == "object-scope")
}

// add_object_fields_to_cmds will declare flag 'cli-exported-for-actions' app object fields attached to actions (kingpin.Cmd)
func (id *initDriverObjectFlags) add_object_field_to_cmds(flag_name string, flag_det *goforjj.YamlFlag) {

	// We can add only APP object fields. A warning is displayed if set to some other objects.
	if o := id.object_name ; o != app {
		 if len(flag_det.CliCmdActions) > 0 {
			 gotrace.Warning("FORJJ Driver '%s-%s': The object '%s' flag '%s' cli-exported-for-actions ignored. This parameter is " +
				 "only supported on '%s' object type",
				 id.d.DriverType, id.d.Name, o, flag_name, app)
		 }
		return
	}

	flag_opts := id.d_opts.SetFlagOptions(flag_name, &flag_det.Options, id.task_has_value)

	// remove kingpin Required option if requested. forjj will test it later.
	if flag_opts.IsRequired() {
		flag_opts.NotRequired()
	}

	// We can export only to recognized Command actions (create/update/maintain)
	for _, action := range flag_det.CliCmdActions {
		if utils.InStringList(action, id.validCommandActions...) == "" {
			gotrace.Warning("FORJJ Driver '%s-%s': object '%s' flag '%s'. cli-exported-for-actions declares invalid action '%s'. ignored.",
				id.d.DriverType, id.d.Name, id.object_name, flag_name, action)
			continue
		}
		if id.a.cli.OnActions(action).WithObjectInstance(id.object_name, id.instance_name).
			AddActionFlagFromObjectField(flag_name, flag_opts) == nil {
			gotrace.Error("Unable to declare field '%s' to Command action '%s'. %s",
				flag_name, action, id.a.cli.Error())
		}
		gotrace.Trace("Flag '%s' added to action object instance action '%s/%s %s'.", id.object_name, id.instance_name, action)
	}
}

func (id *initDriverObjectFlags) add_object_fields(flag_name string, flag_det *goforjj.YamlFlag, default_actions []string) (flag_updated bool) {
	if id.obj.HasField(flag_name) {
		gotrace.Trace("Object '%s': Field '%s' has already been defined as an object field. Ignored.",
			id.obj.Name(), flag_name)
		return
	}

	flag_opts := id.d_opts.SetFlagOptions(flag_name, &flag_det.Options, id.task_has_value)

	if id.is_field_object_scope(flag_det) {
		if flag_det.FormatRegexp == "" {
			id.obj.AddField(cli.String, flag_name, flag_det.Help, ".*", flag_opts)
		} else {
			id.obj.AddField(cli.String, flag_name, flag_det.Help, flag_det.FormatRegexp, flag_opts)
		}

		gotrace.Trace("Object '%s' field '%s' added.", id.obj.Name(), flag_name)
	} else {
		for _, instance_name := range id.obj.GetInstances() {
			id.obj.AddInstanceField(instance_name, cli.String, flag_name, flag_det.Help, flag_det.FormatRegexp, flag_opts)
			gotrace.Trace("Object Instance '%s-%s': Field '%s' added.", id.obj.Name(), instance_name, flag_name)
		}
	}

	// Checking flag actions definition.
	if flag_det.Actions != nil {
		for _, action := range flag_det.Actions {
			action_name := utils.InStringList(action, id.validActions...)
			if action_name == "" {
				log.Printf("FORJJ Driver '%s-%s': Invalid action '%s' for field '%s'. Accept only '%s'. Ignored.",
					id.d.DriverType, id.d.Name, action, flag_name, id.validActions)
				// Remove this bad Action name from yaml loaded driver.
				flag_det.Actions = utils.ArrayStringDelete(flag_det.Actions, action)
				flag_updated = true
				continue
			}
		}
	}

	if flag_det.Actions == nil || len(flag_det.Actions) == 0 {
		flag_det.Actions = default_actions
		flag_updated = true
		gotrace.Trace("Object '%s': Field '%s' is defined for DEFAULT actions '%s'",
			id.obj.Name(), flag_name, default_actions)
	} else {
		gotrace.Trace("Object '%s': Field '%s' is defined for actions '%s'",
			id.obj.Name(), flag_name, flag_det.Actions)
	}

	// Determine all actions for the object. Required by cli.ForjObject.DefineActions() on new objects.
	if !id.allActions {
		for _, action_name := range default_actions {
			id.defineActions[action_name] = true
		}
		id.allActions = true
		for key := range id.defineActions {
			if !id.defineActions[key] {
				id.allActions = false
				break
			}
		}
		gotrace.Trace("Object '%s': will be defined with all actions.", id.object_name)
	}
	return
}

func (id *initDriverObjectFlags) add_object_actions() []string {
	// Adding Actions to the object.
	actionsToAdd := make([]string, 0, len(id.defineActions))
	for action_name, toAdd := range id.defineActions {
		if toAdd {
			actionsToAdd = append(actionsToAdd, action_name)
		}
	}
	id.obj.DefineActions(actionsToAdd...)
	gotrace.Trace("Object '%s': Actions %s added.", id.obj.Name(), actionsToAdd)
	return actionsToAdd
}

// TODO: Revisit how we create commands/flags/args from the driver. https://github.com/forj-oss/forjj/issues/58

// add_object_actions_flags
// warning: Ensure flag_det.Actions is properly set before calling this function.
func (id *initDriverObjectFlags) add_object_actions_flags(
	flag_name string,
	flag_det *goforjj.YamlFlag,
	actionsToAdd []string,
) {
	if ok, _ := regexp.MatchString("forjj-.*", flag_name); ok {
		gotrace.Trace("Object '%s': '%s' is an internal FORJJ variable. Not added in any object actions.",
			id.object_name, flag_name)
		return
	}

	id.add_object_actions_flag(flag_det.Actions, flag_name, flag_det)
}

// is_object_scope determine if the field scope should be global or not
// It decides in the following order:
// - at yaml flag level (/objects/<object_name>/flags/<flag_name>)
// - at yaml object level (/objects/<object_name>)
// - from object role
func (id *initDriverObjectFlags) is_flag_object_scope(flag_name string, flag_det *goforjj.YamlFlag) (bool) {
	if v := flag_det.FlagScope; v != "" {
		// if the flag exist as a global field, so we can't create an instance flag. No instance.
		if found, asObjectField := id.obj.IsObjectField(flag_name) ; v == "instance" && found && asObjectField{
			return true
		}
		return (v == "object")
	}
	if v := id.object_det.FlagsScope; v != "" {
		return (v == "object")
	}
	return (id.obj.HasRole() == "object-scope")
}

func (id *initDriverObjectFlags) add_object_actions_flag(actions []string, flag_name string, flag_det *goforjj.YamlFlag) {
	gotrace.Trace("Object '%s': Adding flag '%s' for actions '%s'.", id.object_name, flag_name, flag_det.Actions)

	if id.is_flag_object_scope(flag_name, flag_det) {
		id.obj.OnActions(actions...)
		id.object_instance_name = ""
		flag_opts := id.d_opts.SetFlagOptions(flag_name, &flag_det.Options, id.object_instance_has_value)
		id.obj.AddFlag(flag_name, flag_opts)
		return
	}
	for _, id.object_instance_name = range id.obj.GetInstances() {
		flag_opts := id.d_opts.SetFlagOptions(flag_name, &flag_det.Options, id.object_instance_has_value)
		id.a.cli.OnActions(actions...).
			WithObjectInstance(id.object_name, id.object_instance_name).
			AddActionFlagFromObjectField(flag_name, flag_opts)
	}
}

// object_instance_has_value will determine which default value to add to a cli flag.
// It is called in the context of a plugin `object_fields`
// It is searching in creds (drivers.DriverOptions) and in Forjfile (forjfile.Forge)
//
// If a value is found in both creds and Forjfile, creds is chosen.
//
// In Forjfile, an object flag is set in several yaml sections. The mapping is:
// Objects:
// - app => /applications
// - repo => /repositories
// - user => /forj/users
// - group => /forj/groups
//
// Any other objects are in /forj/<object>
//
func (id *initDriverObjectFlags) object_instance_has_value(flag string) (value string, found bool) {
	value, found = id.d_opts.HasValue(flag)
	if found {
		return
	}
	return id.a.f.GetString(id.object_name, id.object_instance_name ,flag)
}
