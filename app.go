package main

import (
	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/cli"
	//	"github.com/forj-oss/forjj-modules/cli/interface"
	"github.com/forj-oss/forjj-modules/cli/kingpinCli"
	"github.com/forj-oss/goforjj"
	"github.com/forj-oss/forjj-modules/trace"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
)

// TODO: Support multiple contrib sources.
// TODO: Add flag for branch name to ensure local git branch is correct.

// ActionOpts: Struct for args/flags for an action
//type ActionOpts struct {
//	name     string                         // Action name
//	flags    map[string]*kingpin.FlagClause // list of additional flags loaded.
//	flagsv   map[string]*string             // list of additional flags value loaded.
//	args     map[string]*kingpin.ArgClause  // List of Arguments by name
//	argsv    map[string]*string             // List of Arguments value by name
//	repoList *ReposList                     // List of values for --(add-)?repos flag.
//	Cmd      *kingpin.CmdClause             // Command object
//}

type DriverCmdOptions struct {
	flags map[string]DriverCmdOptionFlag // list of flags values
	//    args  map[string]string // list of args values
}

type DriverCmdOptionFlag struct {
	driver_flag_name string
	value            string
}

type Driver struct {
	DriverType   string                      // driver type name
	InstanceName string                      // Instance name.
	Name         string                      // Name of driver to load Yaml.Name is the real internal driver name.
	cmds         map[string]DriverCmdOptions // List of flags per commands
	//flags         map[string]*kingpin.FlagClause // list of additional flags loaded at app level.
	//flagsv        map[string]*string             // list of additional flags value loaded at app level.
	plugin        goforjj.PluginDef          // Plugin Data
	InfraRepo     bool                       // True if this driver instance is the one hosting the infra repository.
	FlagFile      string                     // Path to the predefined plugin or generic forjj plugin flag file.
	ForjjFlagFile bool                       // true if the flag_file is set by forjj.
	app_request   bool                       // true if the driver is loaded by a apps create/update/maintain task (otherwise requested by Repos or flows request.)
	Runtime       *goforjj.YamlPluginRuntime // Reference to the plugin runtime information given by the plugin yaml file.
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
	drivers_list DriversList // List of drivers passed to the command line argument from --app.
	//Actions      map[string]*ActionOpts // map of Commands with their arguments/flags

	flags_loaded map[string]string // key/values for flags loaded. Used when doing a create AND maintain at the same time (create case)

	drivers         map[string]*Driver // List of drivers data/flags/... per instance name (key)
	drivers_options DriversOptions     // forjj-maintain.yml See infra-maintain.go

	cli *cli.ForjCli // ForjCli data
	app *kingpin.Application

	//	CurrentCommand clier.CmdClauser // Current Command
	//	CurrentObject  clier.CmdClauser // Current Object

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

/*const (
	ssh_dir = "ssh-dir"
)*/

const (
	cr_act    string = "create"
	chg_act   string = "change"
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
	//opts_ssh_dir := cli.Opts().Default(fmt.Sprintf("%s/.ssh", os.Getenv("HOME")))
	opts_contribs_repo := cli.Opts().Envar("CONTRIBS_REPO")
	opts_flows_repo := cli.Opts().Envar("FLOWS_REPO")
	opts_repotmpl := cli.Opts().Envar("REPOTEMPLATES_REPO")
	opts_infra_repo := cli.Opts().Short('I').Default("<organization>-infra")
	opts_creds_file := cli.Opts().Short('C')
	opts_orga_name := cli.Opts().Short('O')
	opts_workspace := cli.Opts().Envar("FORJJ_WORKSPACE").Short('W')

	a.app = kingpin.New(os.Args[0], forjj_help).UsageTemplate(DefaultUsageTemplate)
	a.app.Version("forjj V0.0.1 (POC)").Author("Christophe Larsonneur <christophe.larsonneur@hpe.com>")
	// kingpin is driven by cli module.
	a.cli = cli.NewForjCli(kingpinCli.New(a.app))

	a.cli.ParseHook(a.ParseContext)
	// Regular filter for lists
	// Used by list capture function parameter
	a.cli.AddFieldListCapture("w", `[a-z]+[a-z0-9_-]*`)
	a.cli.AddFieldListCapture("ft", `[A-Za-z0-9_ !:/.-]+`)

	a.cli.AddAppFlag(cli.String, cred_f, forjj_creds_help, opts_creds_file)

	u, _ := url.Parse("https://github.hpe.com/forj/forjj-contribs/raw/master")
	a.ContribRepo_uri = u
	u, _ = url.Parse("https://github.hpe.com/forj/forjj-repotemplates/raw/master")
	a.RepotemplateRepo_uri = u
	u, _ = url.Parse("https://github.hpe.com/forj/forjj-flows/raw/master")
	a.FlowRepo_uri = u

	a.drivers = make(map[string]*Driver)
	//a.Actions = make(map[string]*ActionOpts)
	a.o.Drivers = make(map[string]*Driver)

	// ACTIONS ************
	// Create kingpin actions layer in kingpin.
	// ex: forjj add
	a.cli.NewActions(cr_act, create_action_help, "Create %s.", true)
	a.cli.NewActions(upd_act, update_action_help, "Update %s.", false)
	a.cli.NewActions(maint_act, maintain_action_help, "Maintain %s.", true)
	a.cli.NewActions(add_act, add_action_help, "Add %s to your software factory.", false)
	a.cli.NewActions(chg_act, update_action_help, "Update %s of your software factory.", false)
	a.cli.NewActions(rem_act, remove_action_help, "Remove %s from your software factory.", false)
	a.cli.NewActions(ren_act, rename_action_help, "Rename %s of your software factory.", false)
	a.cli.NewActions(list_act, list_action_help, "List %s of your software factory.", false)

	// OBJECTS ************
	// Create Object layer in kingpin on top of each actions.
	// ex: forjj add repo
	if a.cli.NewObject(workspace, "any forjj workspace parameters", true).
		AddKey(cli.String, workspace, workspace_path_help).
		AddField(cli.String, "docker-exe-path", docker_exe_path_help).
		AddField(cli.String, "contribs-repo", contribs_repo_help).
		AddField(cli.String, "flows-repo", flows_repo_help).
		AddField(cli.String, "repotemplates-repo", repotemplates_repo_help).
		AddField(cli.String, infra_f, forjj_infra_name_help).
		AddField(cli.String, orga_f, forjj_orga_name_help).
		DefineActions(chg_act, rem_act).
		OnActions(chg_act, rem_act).
		AddFlag(workspace, opts_workspace).
		AddFlag("docker-exe-path", nil).
		AddFlag("contribs-repo", opts_contribs_repo).
		AddFlag("flows-repo", opts_flows_repo).
		AddFlag("repotemplates-repo", opts_repotmpl).
		AddFlag(infra_f, opts_infra_repo).
		AddFlag(orga_f, opts_orga_name) == nil {
		log.Printf("%s", a.cli.GetObject(workspace).Error())
	}

	if a.cli.NewObject(repo, "a GIT repository", true).
		AddKey(cli.String, "name", repo_name_help).
		AddField(cli.String, "instance", repo_instance_name_help).
		AddField(cli.String, "flow", repo_flow_help).
		AddField(cli.String, "repo-template", repo_template_help).
		AddField(cli.String, "title", repo_title_help).
		AddField(cli.String, "new-name", new_repo_name_help).
		DefineActions(add_act, chg_act, rem_act, ren_act, list_act).
		OnActions(add_act).
		AddFlag("instance", nil).
		OnActions(add_act, chg_act).
		AddFlag("flow", nil).
		AddFlag("repo-template", nil).
		AddFlag("title", nil).
		OnActions(add_act, chg_act, rem_act, ren_act).
		AddArg("name", opts_required).
		OnActions(ren_act).
		AddArg("new-name", opts_required) == nil {
		log.Printf("%s", a.cli.GetObject(repo).Error())
	}

	// Define create repo list
	if a.cli.GetObject(repo).CreateList("to_create", ",", "(#w/)?#w(:#w(:#w(:#ft)?)?)?", "one or more GIT repositories").
		Field(2, "instance").Field(3, "name").Field(5, "flow").Field(7, "repo-template").Field(9, "title").
		// Ex: forjj add/change repos "github/myrepo:::My Repo" "other_repo:::Another repo"
		//     forjj add/change repos "github/myrepo:::My Repo,other_repo:::Another repo"
		AddActions(add_act, chg_act) == nil {
		log.Printf("%s", a.cli.GetObject(repo).Error())
	}

	// Define remove repo list
	if a.cli.GetObject(repo).CreateList("to_remove", ",", "#w", "one or more GIT repositories").
		Field(1, "name").
		AddActions(rem_act) == nil {
		log.Printf("%s", a.cli.GetObject(repo).Error())
	}

	if a.cli.NewObject(app, "an application driver", true).
		AddKey(cli.String, "name", app_name_help).
		AddField(cli.String, "type", app_type_help).
		AddField(cli.String, "driver", app_driver_help).
		DefineActions(add_act, chg_act, rem_act, list_act).
		OnActions(add_act).
		AddArg("type", opts_required).
		AddArg("driver", opts_required).
		AddArg("name", nil).
		OnActions(chg_act, rem_act).
		AddArg("name", opts_required).
		OnActions(list_act).
		AddFlag("type", nil).
		AddFlag("driver", nil).
		AddFlag("name", nil).
		ParseHook(a.GetDriversFlags) == nil {
		log.Printf("%s", a.cli.GetObject(app).Error())
	}

	// Define app list
	if a.cli.GetObject(app).CreateList("to_create", ",", "#w:#w(:#w)?", "one or more application drivers").
		Field(1, "type").Field(2, "driver").Field(4, "name").
		// Ex: forjj add/change apps <type>:<driver>[:<instance>] ...
		AddActions(add_act, chg_act) == nil {
		log.Printf("%s", a.cli.GetObject(app).Error())
	}

	if a.cli.GetObject(app).CreateList("to_remove", ",", "#w", "one or more application drivers").
		Field(1, "name").
		// Ex: forjj remove apps <instance> ...
		AddActions(rem_act) == nil {
		log.Printf("%s", a.cli.GetObject(app).Error())
	}

	// infra - Mostly built by plugins or other objects list with update action only.
	if a.cli.NewObject(infra, "the global settings", true).
		AddKey(cli.String, "infra-repo", "Infra repository name.").
		AddField(cli.String, "infra-upstream", "Infra repository upstream instance name.").
		AddField(cli.String, "flow", default_flow_help).
		DefineActions(chg_act).
		OnActions().
		AddFlag("infra-repo", cli.Opts().Required()).
		AddFlag("infra-upstream", nil).
		AddFlag("flow", nil) == nil {
		log.Printf("%s", a.cli.GetObject(infra).Error())
	}

	// Flow - Not fully defined.
	if a.cli.NewObject(flow, "flow over applications", true).NoFields().
		DefineActions(add_act, rem_act, list_act) == nil {
		log.Printf("%s", a.cli.GetObject(flow).Error())
	}

	// Enhance create action
	if a.cli.OnActions(cr_act).
		// Ex: forjj create --repos "github/myrepo:::My Repo,other_repo:::Another repo"
		AddActionFlagFromObjectListAction(cr_act, repo, "to_create", add_act).
		// Ex: forjj create --apps "upstream:github"
		AddActionFlagFromObjectListAction(cr_act, app, "to_create", add_act).
		// Add Update workspace flags to Create action, not prefixed.
		// ex: forjj create --docker-exe-path ...
		AddActionFlagsFromObjectAction(workspace, chg_act).
		// Add Update workspace flags to Create action, not prefixed.
		// ex: forjj create --infra-repo ...
		AddActionFlagsFromObjectAction(infra, chg_act).
		AddFlag(cli.String, "ssh-dir", create_ssh_dir_help, nil).
		AddFlag(cli.Bool, "no-maintain", create_no_maintain_help, nil) == nil {
		log.Printf("%s", a.cli.Error())
	}

	// Enhance Update
	if a.cli.OnActions(upd_act).
		// Ex: forjj update infra --add-repos "github/myrepo:::My Repo,other_repo:::Another repo"...
		AddActionFlagsFromObjectListActions(repo, "to_create", add_act).
		// Ex: forjj update infra --remove-repos "myrepo" ... # This will disable the repo only. No real remove.
		AddActionFlagsFromObjectListActions(repo, "to_remove", rem_act).
		// Ex: forjj update infra --add-apps "upstream:github" --github-...
		AddActionFlagsFromObjectListActions(app, "to_create", add_act).
		// Ex: forjj update infra --remove-apps "github" ...
		AddActionFlagsFromObjectListActions(app, "to_remove", rem_act).
		// Add Update workspace flags to Create action, not prefixed.
		// ex: forjj update --docker-exe-path ...
		AddActionFlagsFromObjectAction(workspace, chg_act).
		// Add Update workspace flags to Create action, not prefixed.
		// ex: forjj update --infra-repo ...
		AddFlag(cli.String, "ssh-dir", create_ssh_dir_help, nil) == nil {
		log.Printf("%s", a.cli.Error())
	}

	// Enhance Maintain
	if a.cli.OnActions(maint_act).
		AddFlag(cli.String, "file", maintain_option_file, nil) == nil {
		log.Printf("%s", a.cli.Error())
	}

	// Next to revisit
	//a.GetDriversFlags(os.Args[1:])

	_, err := exec.LookPath("git")
	kingpin.FatalIfError(err, "Unable to find 'git' command. Ensure it available in your PATH and retry.\n")
}

//
// Generic Application function settings
//

// GetActionOptsFromCli
//
// Get the ActionsOpts of the selected Command clause in kingpin (ie create/update or maintain)
/*func (a *Forj) GetActionOptsFromCli(cmd []clier.CmdClauser) {
	if len(cmd) >= 1 {
		a.CurrentCommand = cmd[0]
	}
	if len(cmd) >= 1 {
		a.CurrentObject = cmd[1]
	}
}*/

// InitializeDriversAPI
//
// Function initializing driver flags with values.
// From values found in the commandline, extract them
// From forjj-* values, get it from Forjj internal data.
func (a *Forj) InitializeDriversAPI() {
	// TODO: Use cli : To re-apply
	// We will need to apply value to the Driver REST API and do call to each object requested to transmit the
	// object action. Common are systematically given. And maintain is a different use case.
	/*	forjj_regexp, _ := regexp.Compile("forjj-(.*)")

		for instance_name, driverOpts := range a.drivers {
			if driverOpts.plugin.Yaml.Name == "" {
				continue
			}

			cur_cmd := a.cli.GetCurrentCommand()
			gotrace.Trace("driver: '%s(%s)', command: '%s'", driverOpts.DriverType, instance_name, cur_cmd)
			for _, command := range []string{"common", cur_cmd} { // a.CurrentCommand.FullCommand()} {
				gotrace.Trace(" From '%s' flags list", command)
				for flag_name := range driverOpts.cmds[command].flags {
					gotrace.Trace("  Flag_name => '%s'", flag_name)
					forjj_vars := forjj_regexp.FindStringSubmatch(flag_name)
					f, _ := a.drivers[instance_name].cmds[command].flags[flag_name]
					if forjj_vars == nil {
						if flag_value, ok := a.cli.GetStringValue(flag_name); ok {
							f.value = flag_value
							gotrace.Trace("   %s := %s", flag_name, flag_value)
						}
					} else {
						flag_value := a.GetInternalData(forjj_vars[1])
						f.value = flag_value
						gotrace.Trace("   forjj(%s) => %s := %s", forjj_vars[1], flag_name, flag_value)
					}
					a.drivers[instance_name].cmds[command].flags[flag_name] = f
				}
			}
		}*/
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
