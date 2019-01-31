package main

import (
	"bytes"
	"fmt"
	"forjj/creds"
	"forjj/drivers"
	"forjj/flow"
	"forjj/forjfile"
	"forjj/repo"
	forjjWorkspace "forjj/workspace"
	"forjj/secrets"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/cli/kingpinCli"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
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

	drivers         drivers.Drivers // List of drivers data/flags/... per instance name (key)
	plugins         *goforjj.Plugins           // List of plugins loaded
	drivers_options drivers.DriversOptions     // forjj-maintain.yml See infra-maintain.go

	cli *cli.ForjCli // ForjCli data
	app *kingpin.Application

	// cli commands modules
	secrets   secrets.Secrets
	workspace forjjWorkspace.Workspace

	contextAction string // Context action defined in ParseContext.
	// Can be create/update or maintain. But it can be any others, like secrets...

	actionDispatch map[string]func(string)

	CurrentPluginDriver *drivers.Driver // Driver executing
	InfraPluginDriver   *drivers.Driver // Driver used by upstream

	// Forjj Core values, saved at create time, updated at update time. maintain should save also.

	InternalForjData     map[string]string
	creds_file           *string // Credential file
	forjfile_tmpl_path   string
	Branch               string     // Update feature branch name
	ContribRepoURIs      []*url.URL // URL to github raw files for plugin files.
	RepotemplateRepo_uri *url.URL   // URL to github raw files for RepoTemplates.
	appMapEntries        map[string]AppMapEntry
	no_maintain          *bool    // At create time. true to not start maintain task at the end of create.
	debug_instances      []string // List of instances in debug mode
	from_create          bool     // true when start running maintain from create
	validation_issue     bool     // true if validation of Forjfile has failed.

	// TODO: enhance infra README.md with a template.

	infra_readme string // Initial infra repo README.md text.

	f forjfile.Forge                 // Forge Data stored in the Repository (Loaded from Forjfile)
	w forjfile.Workspace             // Data structure to stored in the workspace. See workspace.go
	s creds.Secure                   // credential file support.
	o ForjjOptions                   // Data structured stored in the root of the infra repo. See forjj-options.go
	d *forjfile.DeploymentCoreStruct // deployment information

	flows flow.Flows

	i repository.GitRepoStruct // Infra Repository management.

	deployContext forjDeployContext
}

/*const (
	ssh_dir = "ssh-dir"
)*/

const (
	val_act     string = "validate"
	cr_act      string = "create"
	chg_act     string = "change"
	add_act     string = "add"
	upd_act     string = "update"
	rem_act     string = "remove"
	ren_act     string = "rename"
	list_act    string = "list"
	maint_act   string = "maintain"
	common_acts string = "common" // Refer to all other actions
)

const (
	workspace = "workspace"
	repo      = "repo"
	infra     = "infra"
	app       = "app"
	flow_obj  = "flow"
)

const (
	infra_path_f     = "infra-path"     // Path where infra repository gets cloned.
	infra_name_f     = "infra-name"     // Name of the infra repository in upstream system
	infra_upstream_f = "infra-upstream" // Name of the infra upstream service instance name (github for example)
	cred_f           = "credentials-file"
	debug_instance_f = "run-plugin-debugger"
	orga_f           = "organization" // Organization name for the Forge. Could be used to set upstream organization.
	// create flags
	forjfile_path_f = "forjfile-path" // Path where the Forjfile model resides.
	// deployTo is the name of the deployment environment to update/maintain.
	deployToArg   = "deploy-to"
	forjfile_f    = "forjfile-name" // Name of the forjfile where the Forjfile model resides.
	ssh_dir_f     = "ssh-dir"
	no_maintain_f = "no-maintain"
	message_f     = "message"
)

const (
	defaultRepoBranch   = "master"
	defaultContribsRepo = "https://github.com/forj-oss/<repo>/raw/" + defaultRepoBranch
	defaultFlowRepo     = "https://github.com/forj-oss/forjj-flows/raw/" + defaultRepoBranch
	defaultRepoTemplate = "https://github.com/forj-oss/forjj-repotemplates/raw/" + defaultRepoBranch
)

// ForjModel is used by template mechanism
type ForjModel struct {
	Forjfile    *forjfile.DeployForgeModel
	Deployments *forjfile.DeploymentsModel
	Current     ForjCurrentModel
	Secret      string
}

// ForjCurrentModel is a sub struct of ForjModel
type ForjCurrentModel struct {
	Type       string
	Name       string
	Deployment string
	Data       interface{}
	Creds      map[string]*goforjj.ValueStruct
}

// Model is used to build a Model to use by text/templates
// Used to interpret data from the Forjfile in memory.
func (a *Forj) Model(object_name, instance_name, key string) *ForjModel {
	ffd := a.f.InMemForjfile()
	data := ForjModel{
		Forjfile:    forjfile.NewDeployForgeModel(ffd),
		Deployments: forjfile.NewDeploymentsModel(a.f.GetDeployments()),
		Current: ForjCurrentModel{
			Type:       object_name,
			Name:       instance_name,
			Deployment: a.f.GetDeployment(),
			Data:       ffd.GetObjectInstance(object_name, instance_name),
			Creds:      a.s.GetObjectInstance(object_name, instance_name),
		},
	}
	data.Secret = ""
	if v, found := data.Current.Creds[key]; found {
		data.Secret = v.GetString()
	}
	return &data
}

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
	opts_contribs_repo := cli.Opts().Envar("CONTRIBS_REPO").Default(defaultContribsRepo)
	a.w.SetDefault(forjfile.ContribRepoPathField, defaultContribsRepo)
	opts_flows_repo := cli.Opts().Envar("FLOWS_REPO").Default(defaultFlowRepo)
	a.w.SetDefault("flow-repo-path", defaultFlowRepo)
	opts_repotmpl := cli.Opts().Envar("REPOTEMPLATES_REPO").Default(defaultRepoTemplate)
	a.w.SetDefault("repotemplate-repo-path", defaultRepoTemplate)
	optsDirsPath := cli.Opts().Envar("PLUGINS_SOCKET_DIR_NAME")
	obsoleteOptsDirsPath := cli.Opts().Envar("PLUGINS_SOCKET_DIRS_PATH") // TODO: Obsolete setup to remove
	opts_infra_repo := cli.Opts().Short('I').Default("<organization>-infra")
	opts_creds_file := cli.Opts().Short('C')
	opts_orga_name := cli.Opts().Short('O')
	opts_infra_path := cli.Opts().Envar("FORJJ_INFRA").Short('W')
	opts_forjfile := cli.Opts().Short('F').Default(".")
	opts_message := cli.Opts().Short('m')

	a.app = kingpin.New(os.Args[0], forjj_help).UsageTemplate(DefaultUsageTemplate)

	a.secrets.Init(a.app, &a.f, &a.drivers, &a.s, a.cli.IsParsePhase, func(context *secrets.Context, cmd *kingpin.CmdClause) {
		// Define Common flags required by ParseContext

		// TODO: Find a way to avoid redefining such common flags option here and re-use cli.Opts
		// Following flags are parseable by cli, and used by ParseContext (so required), but we do not need them on secrets.

		// ISSUE: Default() affect only cli after ParseContext. Default value is retrieved thanks to a fix in GetLocalPrefs()
		context.Flag("contribs-repo",
			cmd.Flag("contribs-repo", contribs_repo_help).Envar("CONTRIBS_REPO").Default(defaultContribsRepo)).String()
		context.Flag("flows-repo",
			cmd.Flag("flows-repo", flows_repo_help).Envar("FLOWS_REPO").Default(defaultFlowRepo)).String()
		context.Flag("repotemplates-repo",
			cmd.Flag("repotemplates-repo", repotemplates_repo_help).Envar("REPOTEMPLATES_REPO").Default(defaultRepoTemplate)).String()
		context.Flag(infra_path_f,
			cmd.Flag(infra_path_f, infra_path_help)).Envar("FORJJ_INFRA").Short('W').String()
	})
	a.workspace.Init(a.app, &a.w, a.cli.IsParsePhase, func(context *forjjWorkspace.Context, cmd *kingpin.CmdClause) {
		// Define Common flags required by ParseContext

		// TODO: Find a way to avoid redefining such common flags option here and re-use cli.Opts
		// Following flags are parseable by cli, and used by ParseContext (so required), but we do not need them on workspace.

		// ISSUE: Default() affect only cli after ParseContext. Default value is retrieved thanks to a fix in GetLocalPrefs()
		context.Flag("contribs-repo",
			cmd.Flag("contribs-repo", contribs_repo_help).Envar("CONTRIBS_REPO").Default(defaultContribsRepo)).String()
		context.Flag("flows-repo",
			cmd.Flag("flows-repo", flows_repo_help).Envar("FLOWS_REPO").Default(defaultFlowRepo)).String()
		context.Flag("repotemplates-repo",
			cmd.Flag("repotemplates-repo", repotemplates_repo_help).Envar("REPOTEMPLATES_REPO").Default(defaultRepoTemplate)).String()
		context.Flag(infra_path_f,
			cmd.Flag(infra_path_f, infra_path_help)).Envar("FORJJ_INFRA").Short('W').String()

	})

	var version string
	if PRERELEASE {
		version = "forjj pre-release V" + VERSION
	} else {
		version = "forjj V" + VERSION
	}

	fmt.Printf("branch %s, build_date %s, build_commit %s, build_tag %s.\n",
		build_branch, build_date, build_commit, build_tag)

	if build_branch != "master" {
		version += fmt.Sprintf(" branch %s", build_branch)
	}
	if build_tag == "false" {
		version += fmt.Sprintf(" patched - %s - %s", build_date, build_commit)
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
	a.cli.AddAppFlag(cli.String, debug_instance_f, "List of plugin instances in debug mode, comma separated.",
		nil)

	a.actionDispatch = make(map[string]func(string))
	a.actionDispatch[cr_act] = a.createAction
	a.actionDispatch[upd_act] = a.updateAction
	a.actionDispatch[maint_act] = a.maintainAction
	a.actionDispatch[val_act] = a.validateAction
	a.actionDispatch["secrets"] = a.secrets.Action
	a.actionDispatch["workspace"] = a.workspace.Action

	a.plugins = goforjj.NewPlugins()
	//a.Actions = make(map[string]*ActionOpts)
	//a.o.Drivers = make(map[string]*drivers.Driver)

	// ACTIONS ************
	// Create kingpin actions layer in kingpin.
	// ex: forjj add
	a.cli.NewActions(cr_act, create_action_help, "Create %s.", true)
	a.cli.NewActions(upd_act, update_action_help, "Update %s.", true)
	a.cli.NewActions(maint_act, maintain_action_help, "Maintain %s.", true)
	a.cli.NewActions(val_act, val_act_help, "", true)
	a.cli.NewActions(add_act, add_action_help, "Add %s to your software factory.", false)
	a.cli.NewActions(chg_act, update_action_help, "Update %s of your software factory.", false)
	a.cli.NewActions(rem_act, remove_action_help, "Remove/disable %s from your software factory.", false)
	a.cli.NewActions(ren_act, rename_action_help, "Rename %s of your software factory.", false)
	a.cli.NewActions(list_act, list_action_help, "List %s of your software factory.", false)

	// OBJECTS ************
	// Create Object layer in kingpin on top of each actions.
	// ex: forjj add repo
	if a.cli.NewObject(workspace, "any forjj workspace parameters", "internal").
		Single().
		AddField(cli.String, infra_path_f, infra_path_help, "#w", opts_infra_path).
		AddField(cli.String, "docker-exe-path", docker_exe_path_help, "#w", nil).
		AddField(cli.String, "contribs-repo", contribs_repo_help, "#w", opts_contribs_repo).
		AddField(cli.String, "flows-repo", flows_repo_help, "#w", opts_flows_repo).
		AddField(cli.String, "repotemplates-repo", repotemplates_repo_help, "#w", opts_repotmpl).
		AddField(cli.String, forjfile.PluginsSocketDirField, socketDirsPathHelp, "#w", optsDirsPath).
		AddField(cli.String, "plugins-sockets-dirs-path", obsoleteSocketPathHelp, "#w", obsoleteOptsDirsPath). // TODO: Obsolete. To remove.
		AddField(cli.String, orga_f, forjj_orga_name_help, "#w", nil).
		DefineActions(chg_act, rem_act).OnActions().
		AddFlag(infra_path_f, nil).
		AddFlag("docker-exe-path", nil).
		AddFlag("contribs-repo", nil).
		AddFlag("flows-repo", nil).
		AddFlag("repotemplates-repo", nil).
		AddFlag(forjfile.PluginsSocketDirField, nil).
		AddFlag("plugins-sockets-dirs-path", nil). // TODO: Obsolete. To remove.
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
		AddField(cli.String, infra_name_f, forjj_infra_name_help, "#w", nil).
		AddField(cli.String, infra_upstream_f, forjj_infra_upstream_help, "#w", nil).
		AddField(cli.String, "flow", default_flow_help, "#w", nil).
		AddField(cli.String, message_f, create_message_help, "#w", opts_message).
		DefineActions(chg_act).
		OnActions().
		AddFlag(infra_name_f, opts_infra_repo).
		AddFlag(infra_upstream_f, nil).
		AddFlag(message_f, nil).
		AddFlag("flow", nil) == nil {
		log.Printf("infra: %s", a.cli.GetObject(infra).Error())
	}

	// Flow - Not fully defined.
	if a.cli.NewObject(flow_obj, "flow over applications", "internal").NoFields().
		DefineActions(add_act, rem_act, list_act) == nil {
		log.Printf("infra: %s", a.cli.GetObject(flow_obj).Error())
	}

	// Enhance create action. Plugins can add options to create with `only-for-actions`
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

	if a.cli.OnActions(val_act).
		// Add Update workspace flags to Create action, not prefixed.
		// ex: forjj create --docker-exe-path ...
		AddActionFlagsFromObjectAction(workspace, chg_act).
		AddFlag(cli.String, forjfile_path_f, create_forjfile_help, opts_forjfile) == nil {
		log.Printf("action create: %s", a.cli.Error())
	}

	// Enhance Update. Plugins can add options to update with `only-for-actions`
	if a.cli.OnActions(upd_act).
		// Add Update workspace flags to Create action, not prefixed.
		// ex: forjj update --docker-exe-path ...
		AddActionFlagsFromObjectAction(workspace, chg_act).
		AddArg(cli.String, deployToArg, updateDeployToHelp, nil).
		AddFlag(cli.Bool, "deploy-publish", updateDeployPublishHelp, nil).
		AddFlag(cli.String, "ssh-dir", create_ssh_dir_help, nil) == nil {
		log.Printf("action update: %s", a.cli.Error())
	}

	// Enhance Maintain. Plugins can add options to maintain with `only-for-actions`
	if a.cli.OnActions(maint_act).
		AddActionFlagsFromObjectAction(workspace, chg_act).
		AddActionFlagFromObjectAction(infra, chg_act, infra_path_f).
		AddArg(cli.String, deployToArg, maintainDeployToHelp, opts_required).
		AddFlag(cli.String, "file", maintain_option_file, nil) == nil {
		log.Printf("action maintain: %s", a.cli.Error())
	}

	_, err := exec.LookPath("git")
	kingpin.FatalIfError(err, "Unable to find 'git' command. Ensure it available in your PATH and retry.\n")

	// Add Forjfile/cli mapping for simple forj data getter
	a.AddMap(orga_f, workspace, "", orga_f, "settings", "", orga_f)
	a.AddMap(infra_name_f, infra, "", infra_name_f, infra, "", "name")
	a.AddMap(infra_upstream_f, infra, "", infra_upstream_f, infra, "", "apps:upstream")
	a.AddMap(deployToArg, "_app", "forjj", deployToArg, "settings", "default", "dev-deploy")
	a.AddMapFunc("secrets", deployToArg, a.secrets.Context.GetStringValue)

	a.AddMap(infra_path_f, workspace, "", infra_path_f, workspace, "", infra_path_f)
	a.AddMapFunc("secrets", infra_path_f, a.secrets.GetStringValue)
	a.AddMapFunc("workspace", infra_path_f, a.workspace.GetStringValue)

	a.AddMap("contribs-repo", workspace, "", "contribs-repo", "", "", forjfile.ContribRepoPathField)
	a.AddMapFunc("secrets", "contribs-repo", a.secrets.GetStringValue)
	a.AddMapFunc("workspace", "contribs-repo", a.workspace.GetStringValue)

	a.AddMap("flows-repo", workspace, "", "flows-repo", "", "", "flow-repo-path")
	a.AddMapFunc("secrets", "flows-repo", a.secrets.GetStringValue)
	a.AddMapFunc("workspace", "flows-repo", a.workspace.GetStringValue)

	a.AddMap("repotemplates-repo", workspace, "", "repotemplates-repo", "", "", "repotemplate-repo-path")
	a.AddMapFunc("secrets", "repotemplates-repo", a.secrets.GetStringValue)
	a.AddMapFunc("workspace", "repotemplates-repo", a.workspace.GetStringValue)
	// TODO: Add git-remote cli mapping
}

// LoadInternalData()
func (a *Forj) LoadInternalData() {
	a.InternalForjData = make(map[string]string)
	ldata := []string{"organization", "infra", "infra-upstream", "instance-name", "source-mount", "workspace-mount", "deploy-mount", "username", "secrets"}
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
		result = a.w.GetString(param)
	case "infra":
		if infra := a.w.Infra(); infra != nil {
			result = infra.Name
		}
	case "infra-upstream", "infra-upstream-url":
		if instanceName := a.w.GetString("infra-instance-name"); instanceName == "" || instanceName == "none" {
			result = ""
		} else {
			if infra := a.w.Infra(); infra != nil {
				result = infra.GetUpstream(param == "infra-upstream")
			}
		}
	case "instance-name":
		if a.CurrentPluginDriver != nil {
			result = a.CurrentPluginDriver.InstanceName
		} else {
			gotrace.Trace("Warning. instance_name requested outside plugin context.")
		}
	case "source-mount": // where the plugin has source source mounted in the container
		if a.CurrentPluginDriver != nil {
			result = a.CurrentPluginDriver.Plugin.SourceMount
		} else {
			gotrace.Trace("Warning. source-mount requested outside plugin context.")
		}
	case "workspace-mount": // where the plugin has source workspace mounted to the container from caller
		if a.CurrentPluginDriver != nil {
			result = a.CurrentPluginDriver.Plugin.WorkspaceMount
		} else {
			gotrace.Trace("Warning. workspace-mount requested outside plugin context.")
		}
	case "deploy-mount": // where the plugin has soure deployment mounted to the container
		if a.CurrentPluginDriver != nil {
			result = a.CurrentPluginDriver.Plugin.DestMount
		} else {
			gotrace.Trace("Warning. deploy-mount requested outside plugin context.")
		}
	case "username": // username running forjj command.
		result = os.Getenv("LOGNAME")
	case "secrets": // all secrets that forjj has global and for current deployment. The text is encrypted.
		// TODO: get encrypted data
		result = ""
	}
	gotrace.Trace("'%s' requested. Value returned '%s'", param, result)
	return
}

// GetDriversActionsParameters
//
// Build the list of plugin shell parameters for dedicated action.
// It will be created as a Hash of values
//
// Common task flags are defined at cli Application layer. All other flags are defined as action flags.
func (a *Forj) GetDriversActionsParameter(d *drivers.Driver, flag_name, action string) (value string, found bool) {
	forjj_interpret, _ := regexp.Compile(`\{\{.*\}\}`)

	if value, found = a.GetInternalForjData(flag_name); found {
		return
	}
	gotrace.Trace("'%s' candidate as parameters.", flag_name)
	parameter_name := d.InstanceName + "-" + flag_name

	// Define where to get the flag
	var getStringValue func(string) (string, error)
	if action == common_acts {
		// Common flags at App layer
		getStringValue = func(parameter string) (string, error) {
			return a.cli.GetAppStringValue(parameter_name)
		}
	} else {
		// From action
		getStringValue = func(parameter string) (string, error) {
			if act := a.cli.GetAction(action); act == nil {
				return "", fmt.Errorf("Unable to find '%s' in action '%s'.", parameter, action)
			} else if v := act.GetStringAddr(parameter); v == nil {
				return "", fmt.Errorf("'%s' as not been initialized as *string in action '%s'.", parameter, action)
			} else {
				return *v, nil
			}
		}
	}

	if v, err := getStringValue(parameter_name); err == nil {
		if forjj_interpret.MatchString(v) {
			gotrace.Trace("Interpreting '%s' from '%s'", v, parameter_name)
			// Initialized defaults value from templates
			var doc bytes.Buffer

			if t, err := template.New("forj-data").Funcs(template.FuncMap{
				"ToLower": strings.ToLower,
			}).Parse(v); err != nil {
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
		gotrace.Error("%s", err)
		return
	}
}

func (a *Forj) GetInternalForjData(flag_name string) (v string, found bool) {
	forjj_regexp, _ := regexp.Compile("forjj-(.*)")
	forjj_vars := forjj_regexp.FindStringSubmatch(flag_name)
	if forjj_vars != nil {
		v, found = a.InternalForjData[forjj_vars[1]]
	}
	return
}
