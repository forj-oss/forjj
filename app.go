package main

import (
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/goforjj"
	"github.com/forj-oss/forjj-modules/trace"
	"net/url"
	"os"
	"os/exec"
	"regexp"
)

// TODO: Support multiple contrib sources.
// TODO: Add flag for branch name to ensure local git branch is correct.

// Struct for args/flags for an action
type ActionOpts struct {
	name     string                         // Action name
	flags    map[string]*kingpin.FlagClause // list of additional flags loaded.
	flagsv   map[string]*string             // list of additional flags value loaded.
	args     map[string]*kingpin.ArgClause  // List of Arguments by name
	argsv    map[string]*string             // List of Arguments value by name
	repoList *ReposList                     // List of values for --(add-)?repos flag.
	Cmd      *kingpin.CmdClause             // Command object
}

type DriverCmdOptions struct {
	flags map[string]DriverCmdOptionFlag // list of flags values
	//    args  map[string]string // list of args values
}

type DriverCmdOptionFlag struct {
	driver_flag_name string
	value            string
}

type Driver struct {
	DriverType    string                         // driver type name
	InstanceName  string                         // Instance name.
	Name          string                         // Name of driver to load Yaml.Name is the real internal driver name.
	cmds          map[string]DriverCmdOptions    // List of flags per commands
	flags         map[string]*kingpin.FlagClause // list of additional flags loaded at app level.
	flagsv        map[string]*string             // list of additional flags value loaded at app level.
	plugin        goforjj.PluginDef              // Plugin Data
	InfraRepo     bool                           // True if this driver instance is the one hosting the infra repository.
	FlagFile      string                         // Path to the predefined plugin or generic forjj plugin flag file.
	ForjjFlagFile bool                           // true if the flag_file is set by forjj.
	app_request   bool                           // true if the driver is loaded by a apps create/update/maintain task (otherwise requested by Repos or flows request.)
	Runtime       *goforjj.YamlPluginRuntime     // Reference to the plugin runtime information given by the plugin yaml file.
	// When a driver is initially loaded, it will be saved here, and used it as ref every where.
	// So we are sure that :
	// - any change in plugin is not failing a running environment.
	// - If no plugin is referenced from cli, we can start it without loading it from the plugin.yaml.
	// - We can manage plugins versions and update when needed or requested.
}

// DriverModel: Structure used as template context. The way to get it: Driver.Model()
type DriverModel struct {
	InstanceName string
	Name         string
}

// DriverOptions: List of maintain drivers options required by each plugin.
type DriverOptions struct {
	driver_type string
	Options     map[string]goforjj.PluginOption // List of options with helps given by the plugin through create/update phase.
}

type DriversOptions struct {
	Drivers map[string]DriverOptions // List of options for each drivers
}

type Forj struct {
	// Collections of fields regarding flags given
	drivers_list DriversList            // List of drivers passed to the command line argument from --app.
	Actions      map[string]*ActionOpts // map of Commands with their arguments/flags

	flags_loaded map[string]string // key/values for flags laoded. Used when doing a create AND maintain at the same time (create case)

	drivers         map[string]*Driver // List of drivers data/flags/... per instance name (key)
	drivers_options DriversOptions     // forjj-maintain.yml See infra-maintain.go

	cli *cli.ForjCli // ForjCli data

	// Flags values
	CurrentCommand *ActionOpts // Loaded CurrentCommand reference.

	CurrentPluginDriver *Driver // Driver executing
	InfraPluginDriver   *Driver // Driver used by upstream

	// Forjj Core values, saved at create time, updated at update time. maintain should save also.
	Infra_repo *string // Infra repository name flag value
	Orga_name  *string // Infra repository name flag value

	creds_file           *string  // Credential file
	Branch               string   // Update feature branch name
	ContribRepo_uri      *url.URL // URL to github raw files for plugin files.
	RepotemplateRepo_uri *url.URL // URL to github raw files for RepoTemplates.
	FlowRepo_uri         *url.URL // URL to github raw files for Flows.
	no_maintain          *bool    // At create time. true to not start maintain task at the end of create.
	// TODO: enhance infra README.md with a template.

	infra_readme string // Initial infra repo README.md text.

	w Workspace    // Data structure to stored in the workspace. See workspace.go
	o ForjjOptions // Data structured stored in the root of the infra repo. See forjj-options.go
	r ReposList    // Collection of Repositories managed. Data structured stored in the root of the infra repo. See repos.go
}

const (
	ssh_dir = "ssh-dir"
)

const (
	cr_act    string = "create"
	add_act   string = "add"
	upd_act   string = "update"
	rem_act   string = "remove"
	ren_act   string = "rename"
	list_act  string = "list"
	maint_act string = "maintain"
)

const (
	workspace = "workspace"
	repo      = "repo"
	infra     = "infra"
	app       = "app"
	flow      = "flow"
)

const (
	debug_f = "debug"
	infra_f = "infra"
	cred_f  = "credentials-file"
	orga_f  = "organization"
)

//
// Define application cli options
//
// Defines the list of valid cli options
// - cli predefined flags/actions/Arguments
// - Load plugin specific flags. (from the plugin yaml file)
func (a *Forj) init() {
	// Define options
	opts_required := cli.Opts().Required()
	opts_ssh_dir := cli.Opts().Default(fmt.Sprintf("%s/.ssh", os.Getenv("HOME")))
	opts_contribs_repo := cli.Opts().Envar("CONTRIBS_REPO")
	opts_flows_repo := cli.Opts().Envar("FLOWS_REPO")
	opts_repotmpl := cli.Opts().Envar("REPOTEMPLATES_REPO")
	opts_infra_repo := cli.Opts().Short('I').Default("<organization>-infra")
	opts_creds_file := cli.Opts().Short('C')
	opts_orga_name := cli.Opts().Short('O')
	opts_workspace := cli.Opts().Required().Envar("FORJJ_WORKSPACE").Short('W')

	a.cli = cli.NewForjCli(kingpin.New(os.Args[0], forjj_help).UsageTemplate(DefaultUsageTemplate))
	a.cli.App.Version("forjj V0.0.1 (POC)").Author("Christophe Larsonneur <christophe.larsonneur@hpe.com>")

	a.cli.AddAppFlag(cli.Bool, debug_f, forjj_debug_help, nil)
	a.cli.AddAppFlag(cli.String, cred_f, forjj_creds_help, opts_creds_file)

	u, _ := url.Parse("https://github.hpe.com/forj/forjj-contribs/raw/master")
	a.ContribRepo_uri = u
	u, _ = url.Parse("https://github.hpe.com/forj/forjj-repotemplates/raw/master")
	a.RepotemplateRepo_uri = u
	u, _ = url.Parse("https://github.hpe.com/forj/forjj-flows/raw/master")
	a.FlowRepo_uri = u

	a.drivers = make(map[string]*Driver)
	a.Actions = make(map[string]*ActionOpts)
	a.o.Drivers = make(map[string]*Driver)

	// ACTIONS ************
	// Create kingpin actions layer in kingpin.
	// ex: forjj add
	a.cli.AddActions(cr_act, create_action_help, "Create %s", true)
	a.cli.AddActions(add_act, add_action_help, "Add or more %s.", false)
	a.cli.AddActions(upd_act, update_action_help, "Update one or more %s.", false)
	a.cli.AddActions(rem_act, remove_action_help, "remove one or more %s.", false)
	a.cli.AddActions(ren_act, rename_action_help, "Rename %s.", false)
	a.cli.AddActions(list_act, list_action_help, "List %s.", false)
	a.cli.AddActions(maint_act, maintain_action_help, "Maintain %s.", true)

	// OBJECTS ************
	// Create Object layer in kingpin on top of each actions.
	// ex: forjj add repo
	a.cli.AddObjectActions(workspace, "forjj workspace", true, upd_act, maint_act)
	a.cli.AddObjectActions(infra, "your infrastructure", true, cr_act, upd_act)
	a.cli.AddObjectActions(repo, "GIT repositories", true, add_act, upd_act, rem_act, ren_act, list_act)
	a.cli.AddObjectActions(app, "application driver", true, add_act, rem_act, list_act)
	a.cli.AddObjectActions(flow, "Flow over applications", true, add_act, rem_act, list_act)

	// FLAGS **************
	// Add flags/Args to Action/Object couple
	//
	// ex: forjj add workspace --docker-exe-path ...
	a.cli.AddObjectActionsParam(cli.Flag, cli.String, workspace, "docker-exe-path", docker_exe_path_help, nil, upd_act, rem_act)
	a.cli.AddObjectActionsParam(cli.Flag, cli.String, workspace, "contribs-repo", contribs_repo_help, opts_contribs_repo, upd_act, rem_act)
	a.cli.AddObjectActionsParam(cli.Flag, cli.String, workspace, "flows-repo", flows_repo_help, opts_flows_repo, upd_act, rem_act)
	a.cli.AddObjectActionsParam(cli.Flag, cli.String, workspace, "repotemplates-repo", repotemplates_repo_help, opts_repotmpl, upd_act, rem_act)
	a.cli.AddObjectActionsParam(cli.Flag, cli.String, workspace, infra_f, forjj_infra_name_help, opts_infra_repo, upd_act, rem_act)
	a.cli.AddObjectActionsParam(cli.Flag, cli.String, workspace, orga_f, forjj_orga_name_help, opts_orga_name, upd_act, rem_act)

	// infra - Mostly built by plugins or other objects list with update action only.

	a.cli.AddObjectActionsParam(cli.Arg, cli.String, repo, "instance", repo_instance_name_help, nil, add_act)
	a.cli.AddObjectActionsParam(cli.Arg, cli.String, repo, "name", repo_name_help, opts_required, add_act, upd_act, rem_act, ren_act)
	a.cli.AddObjectActionsParam(cli.Arg, cli.String, repo, "flow", repo_flow_help, nil, add_act, upd_act)
	a.cli.AddObjectActionsParam(cli.Arg, cli.String, repo, "repo-template", repo_template_help, nil, add_act, upd_act)
	a.cli.AddObjectActionsParam(cli.Arg, cli.String, repo, "title", repo_title_help, nil, add_act, upd_act)

	a.cli.AddObjectActionsParam(cli.Arg, cli.String, app, "type", app_type_help, opts_required, add_act, upd_act, rem_act, ren_act)
	a.cli.AddObjectActionsParam(cli.Arg, cli.String, app, "driver", app_driver_help, opts_required, add_act)
	a.cli.AddObjectActionsParam(cli.Arg, cli.String, app, "driver", app_driver_help, nil, upd_act, rem_act, ren_act)
	a.cli.AddObjectActionsParam(cli.Arg, cli.String, app, "name", app_name_help, nil, add_act, upd_act, rem_act, ren_act)
	a.cli.AddObjectActionsParam(cli.Arg, cli.String, app, "name", app_name_help, opts_required, upd_act, rem_act, ren_act)

	// Flow - Not fully defined.

	// Regular Word for lists
	word := `([a-z]+[a-z0-9_-]*)`
	free_text := `([A-Za-z0-9_ !:/.-]+)`

	// LISTS ************

	// Define Repo list in create
	// Ex: forjj add repos "github/myrepo:::My Repo" "other_repo:::Another repo"
	//     forjj add repos "github/myrepo:::My Repo,other_repo:::Another repo"
	//     forjj create --repos "github/myrepo:::My Repo,other_repo:::Another repo"
	repo_list_def := fmt.Sprintf("(%s/)?%s(:%s(:%s(:%s)?)?)?", word, word, word, word, free_text)
	a.cli.CreateObjectList(repo, "to_create", ",", repo_list_def).
		Field(2, "instance").Field(3, "name").Field(5, "flow").Field(7, "repo-template").Field(9, "title").
		AddActions(add_act)
	// Add the list to "create" action and "update infra"
	a.cli.AddActionsParam(cli.Flag, cli.List, repo+"_to_create", repo_list_help, nil, cr_act)
	a.cli.AddObjectActionsParam(cli.Flag, cli.List, infra, repo+"_to_create", repo_list_help, nil, upd_act)

	// Define App list
	a.cli.CreateObjectList(app, "to_create", ",", fmt.Sprintf("%s:%s(:%s)?", word, word, word)).
		Field(1, "type").Field(2, "driver").Field(4, "name").
		AddActions(add_act)
	// Add the list to "create" action and "update infra"
	a.cli.AddActionsParam(cli.Flag, cli.List, app+"_to_create", app_list_help, nil, cr_act)
	a.cli.AddObjectActionsParam(cli.Flag, cli.List, infra, app+"_to_create", app_list_help, nil, upd_act)

	// Add --workspace/$FORJJ_WORKSPACE/-W on all actions except create.
	// forjj <add/...> --workspace
	a.cli.AddActionsParam(cli.Flag, cli.String, workspace, update_orga_help, opts_workspace,
		upd_act, add_act, list_act, rem_act, ren_act)

	// ACTIONS/OBJECTS enhancement ***********

	// Enhance 'create' action
	// forjj create <workspace>
	a.cli.AddActionsParam(cli.Arg, cli.String, workspace, create_orga_help, opts_required, cr_act)
	a.cli.AddActionsParam(cli.Flag, cli.String, ssh_dir, create_ssh_dir_help, opts_ssh_dir, cr_act)
	a.cli.AddActionsParam(cli.Flag, cli.String, "flow", flow_help, nil, cr_act)
	a.cli.AddActionsParam(cli.Flag, cli.Bool, "no-maintain", create_no_maintain_help, nil, cr_act)

	// Add Update workspace flags to Create action, not prefixed.
	a.cli.AddObjectActionParamsToAction(workspace, upd_act, cr_act, false)

	// Enhance 'update infra':
	// Add Update workspace flags to Update Infra, not prefixed.
	a.cli.AddObjectActionParamsToObjectAction(workspace, upd_act, infra, upd_act, false)

	// Enhance Maintain
	a.cli.AddActionsParam(cli.Flag, cli.String, "file", maintain_option_file, nil, maint_act)

	// Next to revisit
	// a.GetDriversFlags(os.Args[1:])

	_, err := exec.LookPath("git")
	kingpin.FatalIfError(err, "Unable to find 'git' command. Ensure it available in your PATH and retry.\n")
}

//
// Generic Application function settings
//

// GetActionOptsFromCli
//
// Get the ActionsOpts of the selected Command clause in kingpin (ie create/update or maintain)
func (a *Forj) GetActionOptsFromCli(cmd *kingpin.CmdClause) *ActionOpts {
	return a.GetActionOptsFromString(cmd.FullCommand())
}

// GetActionOptsFromString
//
// Get the ActionsOpts of a command string (ie create/update or maintain)
func (a *Forj) GetActionOptsFromString(cmd string) *ActionOpts {
	if v, found := a.Actions[cmd]; found {
		return v
	}
	kingpin.Fatalf("FORJJ Internal error. No matching '%s' in declared commands", cmd)
	return nil
}

// InitializeDriversFlag
//
// Function initializing driver flags with values.
// From values found in the commandline, extract them
// From forjj-* values, get it from Forjj internal data.
func (a *Forj) InitializeDriversFlag() {

	forjj_regexp, _ := regexp.Compile("forjj-(.*)")

	for instance_name, driverOpts := range a.drivers {
		if driverOpts.plugin.Yaml.Name == "" {
			continue
		}

		gotrace.Trace("driver: '%s(%s)', command: '%s'", driverOpts.DriverType, instance_name, a.CurrentCommand.name)
		for _, command := range []string{"common", a.CurrentCommand.name} {
			gotrace.Trace(" From '%s' flags list", command)
			for flag_name := range driverOpts.cmds[command].flags {
				gotrace.Trace("  Flag_name => '%s'", flag_name)
				forjj_vars := forjj_regexp.FindStringSubmatch(flag_name)
				f, _ := a.drivers[instance_name].cmds[command].flags[flag_name]
				if forjj_vars == nil {
					if flag_value, ok := a.CurrentCommand.flagsv[flag_name]; ok && flag_value != nil {
						f.value = *flag_value
						gotrace.Trace("   %s := %s", flag_name, *flag_value)
					}
				} else {
					flag_value := a.GetInternalData(forjj_vars[1])
					f.value = flag_value
					gotrace.Trace("   forjj(%s) => %s := %s", forjj_vars[1], flag_name, flag_value)
				}
				a.drivers[instance_name].cmds[command].flags[flag_name] = f
			}
		}
	}
}

// GetInternalData
//
// Provide value for some forjj internal parameters. Used by InitializeDriversFlag to provide values to plugins as they requested it.
func (a *Forj) GetInternalData(param string) (result string) {
	switch param {
	case "organization":
		result = a.w.Organization
	case "infra":
		result = a.w.Infra.Name
	case "instance-name":
		if a.CurrentPluginDriver != nil {
			result = a.CurrentPluginDriver.InstanceName
		} else {
			gotrace.Trace("Warning. instance_name requested outside plugin context.")
		}
	case "source-mount": // where the plugin has source mounted
		if a.CurrentPluginDriver != nil {
			result = a.CurrentPluginDriver.plugin.SourceMount
		} else {
			gotrace.Trace("Warning. source-mount requested outside plugin context.")
		}
	case "workspace-mount": // where the plugin has source mounted
		if a.CurrentPluginDriver != nil {
			result = a.CurrentPluginDriver.plugin.WorkspaceMount
		} else {
			gotrace.Trace("Warning. workspace-mount requested outside plugin context.")
		}
	}
	gotrace.Trace("'%s' requested. Value returned '%s'", param, result)
	return
}

// GetDriversActionsParameters
//
// Build the list of plugin shell parameters for dedicated action.
// It will be created as a Hash of values
func (a *Forj) GetDriversActionsParameters(cmd_args map[string]string, cmd string) {
	forjj_regexp, _ := regexp.Compile("forjj-(.*)")

	if a.flags_loaded == nil {
		a.flags_loaded = make(map[string]string)
	}

	for _, pluginOpts := range a.drivers {
		for k, v := range pluginOpts.cmds[cmd].flags {
			forjj_vars := forjj_regexp.FindStringSubmatch(k)
			if forjj_vars == nil {
				gotrace.Trace("'%s' candidate as parameters.", k)
				if v_saved, ok := a.flags_loaded[k]; ok {
					v.value = v_saved
				}
				if v.value != "" {
					cmd_args[v.driver_flag_name] = v.value
					a.flags_loaded[k] = v.value
					gotrace.Trace("Set: '%s' <= '%s'", k, v.value)
				}
			} else {
				cmd_args[k] = a.GetInternalData(forjj_vars[1])
			}
		}
	}
}
