package main

import (
	"bytes"
	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/cli/kingpinCli"
	"github.com/forj-oss/forjj-modules/trace"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"text/template"
	"fmt"
	"forjj/forjfile"
	"forjj/drivers"
	"forjj/repo"
	"forjj/creds"
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

type Forj struct {
	// Collections of fields regarding flags given
	drivers_list DriversList // List of drivers passed to the command line argument from --app.
	//Actions      map[string]*ActionOpts // map of Commands with their arguments/flags

	//flags_loaded map[string]string // key/values for flags loaded. Used when doing a create AND maintain at the same time (create case)

	drivers         map[string]*drivers.Driver // List of drivers data/flags/... per instance name (key)
	drivers_options drivers.DriversOptions     // forjj-maintain.yml See infra-maintain.go

	cli *cli.ForjCli // ForjCli data
	app *kingpin.Application

	//	CurrentCommand clier.CmdClauser // Current Command
	//	CurrentObject  clier.CmdClauser // Current Object

	CurrentPluginDriver *drivers.Driver // Driver executing
	InfraPluginDriver   *drivers.Driver // Driver used by upstream

	// Forjj Core values, saved at create time, updated at update time. maintain should save also.

	InternalForjData     map[string]string
	creds_file           *string  // Credential file
	Branch               string   // Update feature branch name
	ContribRepo_uri      *url.URL // URL to github raw files for plugin files.
	RepotemplateRepo_uri *url.URL // URL to github raw files for RepoTemplates.
	FlowRepo_uri         *url.URL // URL to github raw files for Flows.
	appMapEntries        map[string]AppMapEntry
	no_maintain          *bool    // At create time. true to not start maintain task at the end of create.
	// TODO: enhance infra README.md with a template.

	infra_readme string // Initial infra repo README.md text.

	f forjfile.Forge           // Forge Data stored in the Repository (Loaded from Forjfile)
	w forjfile.Workspace       // Data structure to stored in the workspace. See workspace.go
	s creds.YamlSecure         // credential file support.
	o ForjjOptions             // Data structured stored in the root of the infra repo. See forjj-options.go

	i repository.GitRepoStruct // Infra Repository management.
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
	infra_path_f     = "infra-path"        // Path where infra repository gets cloned.
	infra_name_f     = "infra-name"        // Name of the infra repository in upstream system
	infra_upstream_f = "infra-upstream"    // Name of the infra upstream service instance name (github for example)
	cred_f           = "credentials-file"
	orga_f           = "organization"      // Organization name for the Forge. Could be used to set upstream organization.
	// create flags
	forjfile_path_f  = "forjfile-path"     // Path where the Forjfile template resides.
	forjfile_f       = "forjfile-name"     // Name of the forjfile where the Forjfile template resides.
	ssh_dir_f        = "ssh-dir"
	no_maintain_f    = "no-maintain"
	message_f        = "message"
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
	opts_infra_path := cli.Opts().Envar("FORJJ_INFRA").Short('W')
	opts_forjfile := cli.Opts().Short('F').Default(".")
	opts_message := cli.Opts().Short('m')

	a.app = kingpin.New(os.Args[0], forjj_help).UsageTemplate(DefaultUsageTemplate)

	version := "forjj V0.0.1 Alpha"
	if build_branch != "master" {
		version += fmt.Sprintf(" branch %s - %s - %s", build_branch, build_date, build_commit)
	}

	a.app.Version(version).Author("Christophe Larsonneur <clarsonneur@gmail.com>")
	// kingpin is driven by cli module.
	a.cli = cli.NewForjCli(kingpinCli.New(a.app, "forjj"))

	a.cli.ParseAfterHook(a.ParseContext)
	// Regular filter for lists
	// Used by list capture function parameter
	a.cli.AddFieldListCapture("w", `[a-z]+[a-z0-9_-]*`)
	a.cli.AddFieldListCapture("ft", `[A-Za-z0-9_ !:/.-]+`)

	a.cli.AddAppFlag(cli.String, cred_f, forjj_creds_help, opts_creds_file)

	u, _ := url.Parse("https://github.com/forj-oss/forjj-contribs/raw/master")
	a.ContribRepo_uri = u
	u, _ = url.Parse("https://github.com/forj-oss/forjj-repotemplates/raw/master")
	a.RepotemplateRepo_uri = u
	u, _ = url.Parse("https://github.com/forj-oss/forjj-flows/raw/master")
	a.FlowRepo_uri = u

	a.drivers = make(map[string]*drivers.Driver)
	//a.Actions = make(map[string]*ActionOpts)
	//a.o.Drivers = make(map[string]*drivers.Driver)

	// ACTIONS ************
	// Create kingpin actions layer in kingpin.
	// ex: forjj add
	a.cli.NewActions(cr_act, create_action_help, "Create %s.", true)
	a.cli.NewActions(upd_act, update_action_help, "Update %s.", true)
	a.cli.NewActions(maint_act, maintain_action_help, "Maintain %s.", true)
	a.cli.NewActions(add_act, add_action_help, "Add %s to your software factory.", false)
	a.cli.NewActions(chg_act, update_action_help, "Update %s of your software factory.", false)
	a.cli.NewActions(rem_act, remove_action_help, "Remove %s from your software factory.", false)
	a.cli.NewActions(ren_act, rename_action_help, "Rename %s of your software factory.", false)
	a.cli.NewActions(list_act, list_action_help, "List %s of your software factory.", false)

	// OBJECTS ************
	// Create Object layer in kingpin on top of each actions.
	// ex: forjj add repo
	if a.cli.NewObject(workspace, "any forjj workspace parameters", "internal").
		Single().
		AddField(cli.String, "docker-exe-path", docker_exe_path_help, "#w", nil).
		AddField(cli.String, "contribs-repo", contribs_repo_help, "#w", nil).
		AddField(cli.String, "flows-repo", flows_repo_help, "#w", nil).
		AddField(cli.String, "repotemplates-repo", repotemplates_repo_help, "#w", nil).
		AddField(cli.String, orga_f, forjj_orga_name_help, "#w", nil).
		DefineActions(chg_act, rem_act).OnActions().
		AddFlag("docker-exe-path", nil).
		AddFlag("contribs-repo", opts_contribs_repo).
		AddFlag("flows-repo", opts_flows_repo).
		AddFlag("repotemplates-repo", opts_repotmpl).
		AddFlag(orga_f, opts_orga_name) == nil {
		log.Printf("Workspace : %s", a.cli.GetObject(workspace).Error())
	}

	if a.cli.NewObject(repo, "a GIT repository", "object-scope").
		AddKey(cli.String, "name", repo_name_help, "#w", nil).
		AddField(cli.String, "instance", repo_instance_name_help, "#w", nil).
		AddField(cli.String, "flow", repo_flow_help, "#w", nil).
		AddField(cli.String, "repo_template", repo_template_help, "#w", nil).
		AddField(cli.String, "title", repo_title_help, "#ft", nil).
		AddField(cli.String, "new_name", new_repo_name_help, "#w", nil).
		DefineActions(add_act, chg_act, rem_act, ren_act, list_act).
		OnActions(add_act).
		AddFlag("instance", nil).
		OnActions(add_act, chg_act).
		AddFlag("flow", nil).
		AddFlag("repo_template", nil).
		AddFlag("title", nil).
		OnActions(add_act, chg_act, rem_act, ren_act).
		AddArg("name", opts_required).
		OnActions(ren_act).
		AddArg("new_name", opts_required).
		OnActions(add_act, chg_act, rem_act, list_act).
		AddFlagsFromObjectAction(workspace, chg_act) == nil {
		log.Printf("Repo: %s", a.cli.GetObject(repo).Error())
	}

	// Define create repo list
	if a.cli.GetObject(repo).CreateList("to_create", ",",
		"[instance/]name[:flow[:repo_template[:title]]]",
		"one or more GIT repositories").
		// Ex: forjj add/change repos "github/myrepo:::My Repo" "other_repo:::Another repo"
		//     forjj add/change repos "github/myrepo:::My Repo,other_repo:::Another repo"
		AddActions(add_act, chg_act).
		AddFlagsFromObjectAction(workspace, chg_act) == nil {
		log.Printf("repo: to_create list: %s", a.cli.GetObject(repo).Error())
	}

	// Define remove repo list
	if a.cli.GetObject(repo).CreateList("to_remove", ",", "name", "one or more GIT repositories").
		AddActions(rem_act).
		AddFlagsFromObjectAction(workspace, chg_act) == nil {
		log.Printf("repo: to_remove list: %s", a.cli.GetObject(repo).Error())
	}

	if a.cli.NewObject(app, "an application driver", "instance-scope").
		AddKey(cli.String, "name", app_name_help, "#w", nil).
		AddField(cli.String, "type", app_type_help, "#w", nil).
		AddField(cli.String, "driver", app_driver_help, "#w", nil).
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
		ParseHook(a.GetDriversFlags).
		OnActions(add_act, chg_act, rem_act, list_act).
		AddFlagsFromObjectAction(workspace, chg_act) == nil {
		log.Printf("app: %s", a.cli.GetObject(app).Error())
	}

	// Define app list
	if a.cli.GetObject(app).CreateList("to_create", ",", "type:driver[:name]", "one or more application drivers").
		AddValidateHandler(func(l *cli.ForjListData) error {
		if l.Data["name"] == "" {
			driver := l.Data["driver"]
			gotrace.Trace("Set default instance name to '%s'.", driver)
			l.Data["name"] = driver
		}
		return nil
	}).
		// Ex: forjj add/change apps <type>:<driver>[:<instance>] ...
		AddActions(add_act, chg_act).
		AddFlagsFromObjectAction(workspace, chg_act) == nil {
		log.Printf("app: to_create: %s", a.cli.GetObject(app).Error())
	}

	if a.cli.GetObject(app).CreateList("to_remove", ",", "name", "one or more application drivers").
		// Ex: forjj remove apps <instance> ...
		AddActions(rem_act).
		AddFlagsFromObjectAction(workspace, chg_act) == nil {
		log.Printf("%s", a.cli.GetObject(app).Error())
	}

	// infra - Mostly built by plugins or other objects list with update action only.
	if a.cli.NewObject(infra, "the global settings", "internal").
		Single().
		AddField(cli.String, infra_path_f, infra_path_help, "#w", nil).
		AddField(cli.String, infra_name_f, forjj_infra_name_help, "#w", nil).
		AddField(cli.String, infra_upstream_f, forjj_infra_upstream_help, "#w", nil).
		AddField(cli.String, "flow", default_flow_help, "#w", nil).
		AddField(cli.String, message_f, create_message_help, "#w", opts_message).
		DefineActions(chg_act).
		OnActions().
		AddFlag(infra_path_f, opts_infra_path).
		AddFlag(infra_name_f, opts_infra_repo).
		AddFlag(infra_upstream_f, nil).
		AddFlag(message_f, nil).
		AddFlag("flow", nil) == nil {
		log.Printf("infra: %s", a.cli.GetObject(infra).Error())
	}

	// Flow - Not fully defined.
	if a.cli.NewObject(flow, "flow over applications", "internal").NoFields().
		DefineActions(add_act, rem_act, list_act) == nil {
		log.Printf("infra: %s", a.cli.GetObject(flow).Error())
	}

	// Enhance create action
	if a.cli.OnActions(cr_act).
		// Add Update workspace flags to Create action, not prefixed.
		// ex: forjj create --docker-exe-path ...
		AddActionFlagsFromObjectAction(workspace, chg_act).
		// Add Update workspace flags to Create action, not prefixed.
		// ex: forjj create --infra-repo ...
		AddActionFlagsFromObjectAction(infra, chg_act).
		AddFlag(cli.String, ssh_dir_f, create_ssh_dir_help, nil).
	// TODO: Support for a different Forjfile name. (using forjfile_name_f constant)
		AddFlag(cli.String, forjfile_path_f, create_forjfile_help, opts_forjfile).
		AddFlag(cli.Bool, no_maintain_f, create_no_maintain_help, nil) == nil {
		log.Printf("action create: %s", a.cli.Error())
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
		log.Printf("action update: %s", a.cli.Error())
	}

	// Enhance Maintain
	if a.cli.OnActions(maint_act).
		AddActionFlagsFromObjectAction(workspace, chg_act).
		AddFlag(cli.String, "file", maintain_option_file, nil) == nil {
		log.Printf("action maintain: %s", a.cli.Error())
	}

	_, err := exec.LookPath("git")
	kingpin.FatalIfError(err, "Unable to find 'git' command. Ensure it available in your PATH and retry.\n")

	// Add Forjfile/cli mapping for simple forj data getter
	a.AddMap(orga_f, workspace, "", orga_f, "settings", "", orga_f)
	a.AddMap(infra_name_f, infra, "", infra_name_f, infra, "", "name")
	a.AddMap(infra_upstream_f, infra, "", infra_upstream_f, infra, "", "upstream")
	// TODO: Add git-remote cli mapping
}

// LoadInternalData()
func (a *Forj) LoadInternalData() {
	a.InternalForjData = make(map[string]string)
	ldata := []string{"organization", "infra", "infra-upstream", "instance-name", "source-mount", "workspace-mount"}
	for _, param := range ldata {
		a.InternalForjData[param] = a.getInternalData(param)
	}
}

// GetInternalData
//
// Provide value for some forjj internal parameters. Used by InitializeDriversFlag to provide values to plugins as they requested it.
func (a *Forj) getInternalData(param string) (result string) {
	switch param {
	case "organization":
		result = a.w.Organization
	case "infra":
		result = a.w.Infra.Name
	case "infra-upstream":
		if a.w.Instance == "" || a.w.Instance == "none" {
			result = ""
		} else {
			result = a.w.Infra.GetUpstream()
		}
	case "instance-name":
		if a.CurrentPluginDriver != nil {
			result = a.CurrentPluginDriver.InstanceName
		} else {
			gotrace.Trace("Warning. instance_name requested outside plugin context.")
		}
	case "source-mount": // where the plugin has source mounted in the container
		if a.CurrentPluginDriver != nil {
			result = a.CurrentPluginDriver.Plugin.SourceMount
		} else {
			gotrace.Trace("Warning. source-mount requested outside plugin context.")
		}
	case "workspace-mount": // where the plugin has source mounted to the container from caller
		if a.CurrentPluginDriver != nil {
			result = a.CurrentPluginDriver.Plugin.WorkspaceMount
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
func (a *Forj) GetDriversActionsParameter(d *drivers.Driver, flag_name string) (string, bool) {
	forjj_regexp, _ := regexp.Compile("forjj-(.*)")
	forjj_interpret, _ := regexp.Compile(`\{\{.*\}\}`)

	forjj_vars := forjj_regexp.FindStringSubmatch(flag_name)
	if forjj_vars == nil {
		gotrace.Trace("'%s' candidate as parameters.", flag_name)
		parameter_name := d.InstanceName + "-" + flag_name
		if v, err := a.cli.GetAppStringValue(parameter_name); err != nil {
			if forjj_interpret.MatchString(v) {
				gotrace.Trace("Interpreting '%s' from '%s'", v, parameter_name)
				// Initialized defaults value from templates
				var doc bytes.Buffer

				if t, err := template.New("forj-data").Parse(v); err != nil {
					gotrace.Trace("Unable to interpret Parameter '%s' value '%s'. %s", parameter_name, v, err)
					return "", false
				} else {
					t.Execute(&doc, a.InternalForjData)
				}

				v = doc.String()
				gotrace.Trace("'%s' interpreted to '%s'", parameter_name, v)
			}
			gotrace.Trace("Set: '%s' <= '%s'", flag_name, v)
			return v, true
		} else {
			gotrace.Trace("%s", err)
			return "", false
		}
	} else {
		return a.InternalForjData[forjj_vars[1]], true
	}
}
