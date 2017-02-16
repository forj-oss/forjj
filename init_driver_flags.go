package main

import (
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
	"log"
	"regexp"
	"sort"
)

// initDriverObjectFlags internally used by init_driver_flags()
type initDriverObjectFlags struct {
	// Initialized by init_driver_flags()
	d             *Driver
	a             *Forj
	instance_name string
	d_opts        *DriverOptions

	// Initialized by determine_object()
	object_name string
	object_det  *goforjj.YamlObject
	obj         *cli.ForjObject

	// Initialized by prepare_actions_list()
	validActions  []string
	allActions    bool
	defineActions map[string]bool
}

func (id *initDriverObjectFlags) set_task_flags(command string, flags map[string]goforjj.YamlFlag) {
	service_type := id.d.DriverType

	if _, ok := id.a.drivers[id.instance_name].cmds[command]; !ok {
		log.Printf("FORJJ Driver '%s': Invalid tag '%s'. valid one are 'common', 'create', 'update', 'maintain'. Ignored.",
			service_type, command)
	}

	// Sort Flags for readability:
	keys := make([]string, 0, len(flags))
	for k := range flags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, option_name := range keys {
		flag_options := flags[option_name]

		// drivers flags starting with --forjj are a way to communicate some forjj internal data to the driver.
		// They are not in the list of possible drivers options from the cli.
		if ok, _ := regexp.MatchString("forjj-.*", option_name); ok {
			id.d.cmds[command].flags[option_name] = DriverCmdOptionFlag{driver_flag_name: option_name} // No value by default. Will be set later after complete parse.
			continue
		}

		forjj_option_name := id.instance_name + "-" + option_name
		flag_opts := id.d_opts.set_flag_options(option_name, &flag_options.Options)
		if command == "common" {
			// loop on create/update/maintain to create flag on each command
			gotrace.Trace("Create common flags '%s' to App layer.", forjj_option_name)
			id.d.init_driver_flags_for(id.a, option_name, "", forjj_option_name, flag_options.Help, flag_opts)
		} else {
			id.d.init_driver_flags_for(id.a, option_name, command, forjj_option_name, flag_options.Help, flag_opts)
		}
	}

}

func (id *initDriverObjectFlags) determine_object(object_name string, object_det *goforjj.YamlObject) (new bool) {
	id.object_det = object_det
	id.object_name = object_name
	flag_key := id.object_det.Identified_by_flag

	if o := id.a.cli.GetObject(object_name); o != nil {
		if o.IsInternal() {
			gotrace.Trace("'%s' object definition is invalid. This is an internal forjj object. Ignored.", object_name)
			return
		}
		id.obj = o
		gotrace.Trace("Updating object '%s'", object_name)
	} else {
		// New Object and get the key.
		new = true
		id.obj = id.a.cli.NewObject(object_name, id.object_det.Help, false)
		if flag_key == "" {
			id.obj.Single()
			gotrace.Trace("New single object '%s'", object_name)
		} else {
			if v, found := id.object_det.Flags[flag_key]; !found {
				gotrace.Trace("Unable to create the object '%s' identified by '%s'. '%s' is not defined.",
					object_name, flag_key, flag_key)
			} else {
				flag_opts := id.d_opts.set_flag_options(flag_key, &v.Options)
				id.obj.AddKey(cli.String, flag_key, v.Help, v.FormatRegexp, flag_opts)
			}
			gotrace.Trace("New object '%s' with key '%s'", object_name, flag_key)
		}
	}
	return
}

// prepare_actions_list set validActions/defineActions
func (id *initDriverObjectFlags) prepare_actions_list() {
	// Determine which actions can be configured for drivers object flags.
	id.validActions = id.a.get_valid_driver_actions()
	id.defineActions = make(map[string]bool)
	for _, action_name := range id.validActions {
		id.defineActions[action_name] = false
	}
}

func (id *initDriverObjectFlags) add_object_fields(flag_name string, flag_det *goforjj.YamlFlag, default_actions []string) (flag_updated bool) {
	d := id.a.drivers[id.instance_name]
	d_opts := id.a.drivers_options.Drivers[id.instance_name]
	service_type := d.DriverType
	if id.obj.HasField(flag_name) {
		gotrace.Trace("Object '%s': Field '%s' has already been defined as an object field. Ignored.",
			id.obj.Name(), flag_name)
		return
	}
	flag_opts := d_opts.set_flag_options(flag_name, &flag_det.Options)
	id.obj.AddInstanceField(id.instance_name, cli.String, flag_name, flag_det.Help, flag_det.FormatRegexp, flag_opts)
	gotrace.Trace("Object Instance '%s-%s': Field '%s' added.", id.obj.Name(), id.instance_name, flag_name)

	// Checking flag actions definition.
	if flag_det.Actions != nil {
		for _, action := range flag_det.Actions {
			action_name := inStringList(action, id.validActions...)
			if action_name == "" {
				log.Printf("FORJJ Driver '%s-%s': Invalid action '%s' for field '%s'. Accept only '%s'. Ignored.",
					service_type, d.Name, action, flag_name, id.validActions)
				// Remove this bad Action name from yaml loaded driver.
				flag_det.Actions = arrayStringDelete(flag_det.Actions, action)
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

// add_object_actions_flags
// warning: Ensure flag_det.Actions is properly set before calling this function.
func (id *initDriverObjectFlags) add_object_actions_flags(flag_name string, flag_det goforjj.YamlFlag, actionsToAdd []string) {
	if ok, _ := regexp.MatchString("forjj-.*", flag_name); ok {
		gotrace.Trace("Object '%s': '%s' is an internal FORJJ variable. Not added in any object actions.",
			id.object_name, flag_name)
		return
	}

	id.obj.OnActions(flag_det.Actions...)
	gotrace.Trace("Object '%s': Adding flag '%s' for actions '%s'.", id.object_name, flag_name, flag_det.Actions)
	id.obj.AddFlag(flag_name, nil)

	if flag_det.Options.Secure {
		flag_opts := id.d_opts.set_flag_options(flag_name, &flag_det.Options)
		id.a.cli.OnActions(maint_act).
			WithObjectInstance(id.object_name, id.instance_name).
			AddActionFlagFromObjectField(flag_name, flag_opts)
		gotrace.Trace("Object '%s': Secure field '%s' added to maintain task.", id.object_name, flag_name)
	}

}
