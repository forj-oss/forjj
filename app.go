package main

import (
    "fmt"
    "os"
    "os/exec"
    "github.com/alecthomas/kingpin"
    "github.hpe.com/christophe-larsonneur/goforjj"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "net/url"
    "regexp"
)

// TODO: Support multiple contrib sources.
// TODO: Add flag for branch name to ensure local git branch is correct.

// Struct for args/flags for an action
type ActionOpts struct {
    name   string                         // Action name
    flags  map[string]*kingpin.FlagClause // list of additional flags loaded.
    flagsv map[string]*string             // list of additional flags value loaded.
    args   map[string]*kingpin.ArgClause  // List of Arguments by name
    argsv  map[string]*string             // List of Arguments value by name
    repoList *ReposList                   // List of values for --(add-)?repos flag.
    Cmd    *kingpin.CmdClause             // Command object
}

type DriverCmdOptions struct {
    flags map[string]DriverCmdOptionFlag // list of flags values
//    args  map[string]string // list of args values
}

type DriverCmdOptionFlag struct {
    driver_flag_name string
    value string
}

type Driver struct {
    DriverType         string                         // driver type name
    InstanceName       string                         // Instance name.
    Name               string                         // Name of driver to load Yaml.Name is the real internal driver name.
    cmds               map[string]DriverCmdOptions    // List of flags per commands
    flags              map[string]*kingpin.FlagClause // list of additional flags loaded at app level.
    flagsv             map[string]*string             // list of additional flags value loaded at app level.
    plugin goforjj.PluginDef                          // Plugin Data
    InfraRepo          bool                           // True if this driver instance is the one hosting the infra repository.
    FlagFile           string                         // Path to the predefined plugin or generic forjj plugin flag file.
    ForjjFlagFile      bool                           // true if the flag_file is set by forjj.
    app_request        bool                           // true if the driver is loaded by a apps create/update/maintain task (otherwise requested by Repos or flows request.)
}

// Structure used as template context. The way to get it: Driver.Model()
type DriverModel struct {
    InstanceName string
    Name string
}

// List of maintain drivers options required by each plugin.
type DriverOptions struct {
    driver_type string
    Options map[string]goforjj.PluginOption  // List of options with helps given by the plugin through create/update phase.
}

type DriversOptions struct {
    Drivers map[string]DriverOptions         // List of options for each drivers
}


type Forj struct {
    // Collections of fields regarding flags given
    infra_rep_f      *kingpin.FlagClause  // Infra flag kingpin struct.
    Orga_name_f      *kingpin.FlagClause  // Organization flag kingpin struct.
    app              *kingpin.Application // Kingpin Application object
    c_drivers_list_f *kingpin.FlagClause  // Cumulative drivers flag for create
    u_drivers_list_f *kingpin.FlagClause  // Cumulative drivers flag for update
    c_repos_list_f   *kingpin.FlagClause  // Cumulative repos flag for create
    u_repos_list_f   *kingpin.FlagClause  // Cumulative Repos flag for update
    drivers_list     DriversList          // List of drivers passed to the command line argument from --app.
    Actions        map[string]*ActionOpts  // map of Commands with their arguments/flags

    flags_loaded map[string]string        // key/values for flags laoded. Used when doing a create AND maintain at the same time (create case)

    drivers map[string]*Driver            // List of drivers data/flags/... per instance name (key)
    drivers_options DriversOptions        // forjj-maintain.yml See infra-maintain.go


    // Flags values
    CurrentCommand *ActionOpts            // Loaded CurrentCommand reference.
    debug_f        *kingpin.FlagClause    // Flag debug

    CurrentPluginDriver *Driver           // Driver executing
    InfraPluginDriver *Driver             // Driver used by upstream

    // Forjj Core values, saved at create time, updated at update time. maintain should save also.
    Infra_repo *string // Infra repository name flag value
    Orga_name  *string // Infra repository name flag value

    Branch                 string   // Update feature branch name
    Workspace              string   // Workspace name
    Workspace_path         string   // Workspace directory path.
    ContribRepo_uri        *url.URL // URL to github raw files for plugin files.
    RepotemplateRepo_uri   *url.URL // URL to github raw files for RepoTemplates.
    FlowRepo_uri           *url.URL // URL to github raw files for Flows.
    contrib_repo_path      string   // Contribution repository Path
    flow_repo_path         string   // Contribution repository Path
    repotemplate_repo_path string // Contribution repository Path
    // TODO: enhance infra README.md with a template.

    infra_readme      string   // Initial infra repo README.md text.

    w Workspace    // Data structure to stored in the workspace. See workspace.go
    o ForjjOptions // Data structured stored in the root of the infra repo. See forjj-options.go
    r ReposList    // Collection of Repositories managed. Data structured stored in the root of the infra repo. See repos.go
}

const ssh_dir_flag_name = "ssh-dir"

//
// Define application cli options
//

func (a *Forj) init() {
    a.app = kingpin.New(os.Args[0], app_help).UsageTemplate(DefaultUsageTemplate)
    a.debug_f = a.app.Flag("debug", app_debug_help)
    a.debug_f.Bool()
    a.infra_rep_f = a.app.Flag("infra", app_infra_name_help).Short('I').Default("<organization>-infra")
    a.Orga_name_f = a.app.Flag("organization", app_orga_name_help).Short('O')
    a.Orga_name = a.Orga_name_f.String()
    a.Infra_repo = a.infra_rep_f.String()
    a.app.Version("forjj V0.0.1 (POC)").Author("Christophe Larsonneur <christophe.larsonneur@hpe.com>")

    u, _ := url.Parse("https://github.hpe.com/forj/forjj-contribs/raw/master")
    a.ContribRepo_uri = u
    u, _ = url.Parse("https://github.hpe.com/forj/forjj-repotemplates/raw/master")
    a.RepotemplateRepo_uri = u
    u, _ = url.Parse("https://github.hpe.com/forj/forjj-flows/raw/master")
    a.FlowRepo_uri = u

    a.drivers = make(map[string]*Driver)
    a.Actions = make(map[string]*ActionOpts)
    a.o.Drivers = make(map[string]*Driver)

    no_opts := map[string]interface{}{}
    contribs_repo := map[string]interface{}{"envar": "CONTRIBS_REPO"}
    repotemplates_repo := map[string]interface{}{"envar": "REPOTEMPLATES_REPO"}
    flows_repo := map[string]interface{}{"envar": "FLOWS_REPO"}
    required := map[string]interface{}{"required": true}
    ssh_dir_opts := map[string]interface{}{"default": fmt.Sprintf("%s/.ssh", os.Getenv("HOME"))}
    no_set_value_opts := map[string]interface{}{"set_value": false}

    a.SetCommand("create", create_action_help)
    a.SetCmdArg("create", "workspace", create_orga_help, required)
    a.SetCmdFlag("create", ssh_dir_flag_name, create_ssh_dir_help, ssh_dir_opts)
    a.SetCmdFlag("create", "contribs-repo", contribs_repo_help, contribs_repo)
    a.SetCmdFlag("create", "flows-repo", flows_repo_help, flows_repo)
    a.SetCmdFlag("create", "repotemplates-repo", repotemplates_repo_help, repotemplates_repo)
    a.SetCmdFlag("create", "infra-upstream", create_infra_upstream, no_opts)
    a.SetCmdFlag("create", "docker-exe-path", docker_exe_path_help, no_opts)

    // DriversList is loaded from the cli context to adapt the list of supported flags.
    a.c_drivers_list_f = SetDriversListFlag(a.SetCmdFlag("create", "apps", driver_help, no_set_value_opts))
    a.c_repos_list_f = SetReposListFlag(a.Actions["create"], a.SetCmdFlag("create", "repos", repos_help, no_set_value_opts))

    /********** UPDATE Action ************
      Update is not as clear as Create except that update should update the infra repository.

      We could probably use this update to add repositories or migrate the solution to a different place.  */
    a.SetCommand("update", update_action_help)
    a.SetCmdArg("update", "workspace", update_orga_help, required)
    a.SetCmdArg("update", "branch", update_branch_help, required)
    a.SetCmdFlag("update", ssh_dir_flag_name, update_ssh_dir_help, ssh_dir_opts)
    a.SetCmdFlag("update", "contribs-repo", contribs_repo_help, contribs_repo)
    a.SetCmdFlag("update", "flows-repo", flows_repo_help, flows_repo)
    a.SetCmdFlag("update", "repotemplates-repo", repotemplates_repo_help, repotemplates_repo)
    a.SetCmdFlag("update", "infra_upstream", update_infra_upstream, no_opts)
    a.SetCmdFlag("update", "docker-exe-path", docker_exe_path_help, no_opts)
    // Additional options will be loaded from the selected driver itself.

    // DriversList is loaded from the cli context to adapt the list of supported flags.
    a.u_drivers_list_f = SetDriversListFlag(a.SetCmdFlag("update", "apps", driver_help, no_set_value_opts))
    a.u_repos_list_f = SetReposListFlag(a.Actions["update"], a.SetCmdFlag("update", "add-repos", repos_help, no_set_value_opts))

    /****** MAINTAIN Action ************
      Maintain is the real infra creation/update
      It has to ensure that the infra is updated as defined by the source code in $INFRA.

      The technology to ensure (orchestration), could be puppet/ansible combined with docker.
      This is not fully defined.  */
    a.SetCommand("maintain", maintain_action_help)
    a.SetCmdArg("maintain", "workspace", maintain_orga_help, required)
    a.SetCmdFlag("maintain", "infra_url", maintain_infra_url_help, no_opts)
    // FUTURE: Difficulty: 2. Provide a way to expose options defined in the maintain option file as CLI maintain flags.
    // FUTURE: Difficulty: 0. Set default file name from workspace path.
    a.SetCmdFlag("maintain", "file", maintain_option_file, required)

    a.GetDriversFlags(os.Args[1:])

    _, err := exec.LookPath("git")
    kingpin.FatalIfError(err, "Unable to find 'git' command. Ensure it available in your PATH and retry.\n")
}

//
// Generic Application function settings
//

// Get the ActionsOpts of the selected Command clause in kingpin (ie create/update or maintain)
func (a *Forj) GetActionOptsFromCli(cmd *kingpin.CmdClause) *ActionOpts {
    return a.GetActionOptsFromString(cmd.FullCommand())
}

// Get the ActionsOpts of a command string (ie create/update or maintain)
func (a *Forj) GetActionOptsFromString(cmd string) *ActionOpts {
    if v, found := a.Actions[cmd]; found {
        return v
    }
    kingpin.Fatalf("FORJJ Internal error. No matching '%s' in declared commands", cmd)
    return nil
}

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
            for flag_name, _ := range driverOpts.cmds[command].flags {
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

// Provide value for some forjj internal parameters. Used by InitializeDriversFlag to provide values to plugins as they requested it.
func (a *Forj) GetInternalData(param string) (result string) {
    switch param {
    case "organization":
        result = a.w.Organization
    case "infra":
        result = a.w.Infra.Name
    case "instance-name" :
        if a.CurrentPluginDriver != nil {
            result = a.CurrentPluginDriver.InstanceName
        } else {
            gotrace.Trace("Warning. instance_name requested outside plugin context.")
        }
    case "source-mount" : // where the plugin has source mounted
        if a.CurrentPluginDriver != nil {
            result = a.CurrentPluginDriver.plugin.SourceMount
        } else {
            gotrace.Trace("Warning. source-mount requested outside plugin context.")
        }
    case "workspace-mount" : // where the plugin has source mounted
        if a.CurrentPluginDriver != nil {
            result = a.CurrentPluginDriver.plugin.WorkspaceMount
        } else {
            gotrace.Trace("Warning. workspace-mount requested outside plugin context.")
        }
    }
    gotrace.Trace("'%s' requested. Value returned '%s'", param, result)
    return
}

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
            gotrace.Trace("'%s' candidate as parameters.", k)
            if forjj_vars == nil {
                if v_saved, ok := a.flags_loaded[k] ; ok {
                    v.value = v_saved
                }
                if v.value != "" {
                    cmd_args[v.driver_flag_name] = v.value
                    a.flags_loaded[k] = v.value
                }
            } else {
                cmd_args[k] = a.GetInternalData(forjj_vars[1])
            }
        }
    }
}

