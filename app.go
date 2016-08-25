package main

import (
    "fmt"
    "gopkg.in/alecthomas/kingpin.v2"
    "os"
    "os/exec"
    //        "github.com/alecthomas/kingpin"
    "github.hpe.com/christophe-larsonneur/goforjj"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "net/url"
    "path"
    "regexp"
    "log"
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
    driver_type        string                         // driver type name
    instance_name      string                         // Instance name.
    name               string                         // Name of driver to load Yaml.Name is the real internal driver name.
    cmds               map[string]DriverCmdOptions    // List of flags per commands
    flags              map[string]*kingpin.FlagClause // list of additional flags loaded at app level.
    flagsv             map[string]*string             // list of additional flags value loaded at app level.
    plugin goforjj.PluginDef                          // Plugin Data
    infraRepo          bool                           // True if this driver instance is the one hosting the infra repository.
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
    drivers_list     DriversList          // List of drivers passed to the command line argument from --app.
    Actions        map[string]ActionOpts  // map of Commands with their arguments/flags

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

    Branch            string   // branch name
    Workspace         string   // Workspace name
    Workspace_path    string   // Workspace directory path.
    ContribRepo_uri   *url.URL // URL to github raw files
    contrib_repo_path string   // Contribution repository Path
    flow              string   // Name of the flow implemented. defined at create time.
    // TODO: enhance infra README.md with a template.

    infra_readme      string   // Initial infra repo README.md text.

    w Workspace // Data structure to stored in the workspace. See workspace.go
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

    u, _ := url.Parse("https://github.hpe.com/forj/forjj-contribs/raw")

    a.ContribRepo_uri = u
    a.Branch = "master"
    a.drivers = make(map[string]*Driver)
    a.Actions = make(map[string]ActionOpts)

    no_opts := map[string]interface{}{}
    required := map[string]interface{}{"required": true}
    ssh_dir_opts := map[string]interface{}{"default": fmt.Sprintf("%s/.ssh", os.Getenv("HOME"))}
    no_set_value_opts := map[string]interface{}{"set_value": false}

    a.SetCommand("create", create_action_help)
    a.SetCmdArg("create", "workspace", create_orga_help, required)
    a.SetCmdFlag("create", ssh_dir_flag_name, create_ssh_dir_help, ssh_dir_opts)
    a.SetCmdFlag("create", "contrib-repo", create_contrib_help, no_opts)
    a.SetCmdFlag("create", "infra-upstream", create_infra_upstream, no_opts)

    a.c_drivers_list_f = SetDriversListFlag(a.SetCmdFlag("create", "apps", driver_help, no_set_value_opts))

    /********** UPDATE Action ************
      Update is not as clear as Create except that update should update the infra repository.

      We could probably use this update to add repositories or migrate the solution to a different place.  */
    a.SetCommand("update", update_action_help)
    a.SetCmdArg("update", "workspace", update_orga_help, required)
    a.SetCmdFlag("update", ssh_dir_flag_name, update_ssh_dir_help, ssh_dir_opts)
    a.SetCmdFlag("update", "contrib-repo", update_contrib_help, no_opts)
    a.SetCmdFlag("create", "infra_upstream", update_infra_upstream, no_opts)
    // Additional options will be loaded from the selected driver itself.

    a.u_drivers_list_f = SetDriversListFlag(a.SetCmdFlag("update", "apps", driver_help, no_set_value_opts))

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
        return &v
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

        gotrace.Trace("driver: '%s(%s)', command: '%s'", driverOpts.driver_type, instance_name, a.CurrentCommand.name)
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

func (a *Forj) GetInternalData(param string) (result string) {
    switch param {
    case "organization":
        result = a.w.Organization
    case "branch":
        result = a.Branch
    case "infra":
        result = a.w.Infra
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

// Load cli context to adapt the list of options/flags from the driver definition.
//
// It will
// - detect the debug mode
// - detect the organization name/path (to stored in app)
//   It will set the default Infra name.
// - detect the driver list source.
// - detect ci/us drivers name (to stored in app)
//
func (a *Forj) LoadContext(args []string) {
    context, err := a.app.ParseContext(args)
    if context == nil {
        kingpin.FatalIfError(err, "Application flags initialization issue. Driver flags issue?")
    }

    cmd := context.SelectedCommand
    if cmd == nil {
        return
    }

    opts := a.GetActionOptsFromCli(cmd)

    a.CurrentCommand = opts

    if debug_mode, found := a.flagValue(context, a.debug_f); found {
        // global debug defined in trace.go
        log.Printf("Debug set to '%s'.\n", debug_mode)
        if debug_mode == "true" {
            gotrace.SetDebug()
        }
    }

    // The value is not set in argsv. But is in the parser context.
    if orga, found := a.argValue(context, opts.args["workspace"]); found {
        // TODO: Test the path given.
        a.Workspace = path.Base(orga)
        a.Workspace_path = path.Dir(orga)
    }

    // Load Workspace information
    a.w.Load(a)

    // Set organization name to use.
    // Can be set only the first time
    if a.w.Organization == "" {
        if orga, found := a.flagValue(context, a.Orga_name_f); !found {
            a.Orga_name_f = a.Orga_name_f.Default(a.Workspace)
            a.w.Organization = a.Workspace
        } else {
            a.w.Organization = orga
        }
    } else {
        if orga, found := a.flagValue(context, a.Orga_name_f); found && orga != a.w.Organization {
            fmt.Printf("Warning!!! You cannot update the organization name in an existing workspace.\n")
        }
    }

    if a.w.Organization != "" {
        log.Printf("Organization : '%s'", a.w.Organization)
        // Set the 'infra' default flag value
        a.infra_rep_f = a.infra_rep_f.Default(fmt.Sprintf("%s-infra", a.w.Organization))
    }

    // Set the infra repo name to use
    // Can be set only the first time
    if a.w.Infra == "" {
        if infra, found := a.flagValue(context, a.infra_rep_f); found {
            // Get infra name from the flag
            a.w.Infra = infra
        } else { // Or use the default setting.
            a.w.Infra = fmt.Sprintf("%s-infra", a.w.Organization)
        }
    } else {
        if infra, found := a.flagValue(context, a.infra_rep_f); found && infra != a.w.Infra {
            fmt.Printf("Warning!!! You cannot update the Infra repository name in an existing workspace.\n")
        }
    }

    gotrace.Trace("Infrastructure repository defined : %s", a.w.Infra)

    // Identifying appropriate Contribution Repository.
    // The value is not set in flagsv. But is in the parser context.
    if contrib_repo, found := a.flagValue(context, opts.flags["contrib-repo"]); found {
        if u, err := url.Parse(contrib_repo); err != nil {
            println(err)
        } else {
            a.ContribRepo_uri = u
            if u.Scheme == "" {
                a.contrib_repo_path = contrib_repo
            }
        }
    }

    // Getting list of drivers (--app)
    a.drivers_list.list = make(map[string]DriverDef)
    a.drivers_list.GetDriversFromContext(context, a.c_drivers_list_f)
    a.drivers_list.GetDriversFromContext(context, a.u_drivers_list_f)

}

func (*Forj) argValue(context *kingpin.ParseContext, f *kingpin.ArgClause) (value string, found bool) {
    for _, element := range context.Elements {
        if flag, ok := element.Clause.(*kingpin.ArgClause); ok && flag == f {
            value = *element.Value
            found = true
            return
        }
    }
    return
}

func (*Forj) flagValue(context *kingpin.ParseContext, f *kingpin.FlagClause) (value string, found bool) {
    for _, element := range context.Elements {
        if flag, ok := element.Clause.(*kingpin.FlagClause); ok && flag == f {
            value = *element.Value
            found = true
            return
        }
    }
    return
}

// Set an application command
func (a *Forj) SetCommand(name, help string) {
    a.Actions[name] = ActionOpts{
        name:   name,
        Cmd:    a.app.Command(name, help),
        flags:  make(map[string]*kingpin.FlagClause),
        flagsv: make(map[string]*string),
        args:   make(map[string]*kingpin.ArgClause),
        argsv:  make(map[string]*string),
    }
}

// Set a command argument
func (a *Forj) SetCmdArg(cmd, name, help string, options map[string]interface{}) {
    arg := a.Actions[cmd].Cmd.Arg(name, help)

    if v, ok := options["required"]; ok && to_bool(v) {
        arg.Required()
    }
    if v, ok := options["default"]; ok {
        arg.Default(to_string(v))
    }

    a.Actions[cmd].argsv[name] = arg.String()
    a.Actions[cmd].args[name] = arg
}

// Set a Command flag.
func (a *Forj) SetCmdFlag(cmd, name, help string, options map[string]interface{}) (arg *kingpin.FlagClause) {
    arg = a.Actions[cmd].Cmd.Flag(name, help)

    if v, ok := options["required"]; ok && to_bool(v) {
        arg.Required()
    }
    if v, ok := options["default"]; ok {
        arg.Default(to_string(v))
    }
    if v, ok := options["hidden"]; ok && to_bool(v) {
        arg.Hidden()
    }

    if v, ok := options["set_value"]; ok && to_bool(v) {
        if to_bool(v) {
            a.Actions[cmd].flagsv[name] = arg.String()
        } else {
            a.Actions[cmd].flagsv[name] = nil
        }
    } else {
        a.Actions[cmd].flagsv[name] = arg.String()
    }

    a.Actions[cmd].flags[name] = arg

    return
}
