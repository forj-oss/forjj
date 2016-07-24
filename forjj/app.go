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
    flags map[string]string // list of flags values
    args  map[string]string // list of args values
}

type Driver struct {
    driver_type        string                      // driver type name
    name               string                      // Name of driver to load Yaml.Name is the real internal driver name.
    cmds               map[string]DriverCmdOptions // List of flags per commands
    goforjj.PluginData                             // Plugin Data
    Yaml               goforjj.YamlPlugin          // Plugin yaml definition
}

type Forj struct {
    // Collections of fields regarding flags given
    infra_rep_f *kingpin.FlagClause   // Infra flag kingpin struct.
    Orga_name_f *kingpin.FlagClause   // Organization flag kingpin struct.
    app         *kingpin.Application  // Kingpin Application object
    Actions     map[string]ActionOpts // map of Commands with their arguments/flags

    drivers map[string]Driver // List of drivers data/flags/...

    // Flags values
    CurrentCommand *ActionOpts         // Loaded CurrentCommand reference.
    debug_f        *kingpin.FlagClause // Flag debug

    // Forjj Core values, saved at create time, updated at update time. maintain should save also.
    Infra_repo *string // Infra repository name flag value
    Orga_name  *string // Infra repository name flag value

    Branch            string   // branch name
    Workspace         string   // Workspace name
    Workspace_path    string   // Workspace directory path.
    ContribRepo_uri   *url.URL // URL to github raw files
    contrib_repo_path string   // Contribution repository Path
    flow              string   // Name of the flow implemented. defined at create time.

    w Workspace // Data structure to stored in the workspace. See workspace.go
}

const (
    ci_flag_name      = "ci"
    us_flag_name      = "git-us"
    ssh_dir_flag_name = "ssh-dir"
)

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
    a.drivers = make(map[string]Driver)
    a.Actions = make(map[string]ActionOpts)

    /*********** CREATE Action ********** //
      'Create' is the first part developed.

      The objective:
      - Provide a way to start some additionnal script/tool to build
        - an application context (for example, make github ready, with organization/repos/users minimal configuration
        - A relation between application (relation between github/jenkins with a predefined flow like PR flow)
        Those scripts are called 'drivers'
        I'm currently limiting it to be just ci and git-us (to set upstream and his context)
        We assume that repositories are necessarily GIT.
      - Context to start those scripts are known (docker container for that.)
      - a simple binary to help starting docker (mount/env/...) with appropriate parameters. We could simply download and run it.
      - The list of drivers options from a path or url. They are passed to the docker container.

      direct FORJJ role are:
      - create any needed GIT repository. The first one is infra.
      - call the appropriate driver to configure 'ci', 'git-us' and the link. FORJJ won't do any ci/git-us configuration stuff, without those drivers.
      - identify the list of drivers options that is supported.
      - do any git task (pull/push/add/commit/...)

      drivers roles:
      * CI:
        - create a directory containing any source files required to build a CI from code.
          MUST be stored in $INFRA/ci/<driverName>/
          No git task are done. This is the role of FORJJ to get files updated by the drivers back in the repo
        - Stored the configuration in the $INFRA/ci/<driverName>/ path.
        - MUST have a complete documentation on how to configure it to build the appropriate configuration.
      * GIT Upstream:
        - create a directory containing any source files required to build a GIT upstream service.
          For github, for example, it will use the API to create organization/repositories
          For gitlab, for example, it will do the same as github, but could also install it before if asked.
          Any source files MUST be stored in $INFRA/upstream/<driverName>/
        - Stored the configuration in the $INFRA/upstream/<driverName>/ path.
        - MUST have a complete documentation on how to configure it to build the appropriate configuration.

      FORJJ 'Create' won't do the update itself except for the SCM. The SCM upstream must exist and properly configured to track and send infra source code.
      As only the SCM side is really created/configure at create time, the other code is just stored in the INFRA repository.
      The update of the infra could be done by the Update task.
      The implementation of the infra will be done by the maintain task

      By default, 'Create' will create a fully runnable solution. But this solution is made in 2 parts:
      1. Preparation step: Create 'infra' repo, ensure the upstream is properly set and push it.
      2. Creation step: the infra is created from the source code (role of Maintain)

      During the creation, we could optionnaly limit to the first step. But I'm not fully sure it could be useful for now.

      Normally, I do not expect to re-use 'Create' task, when the infra has been REALLY created. The update should be done by Update task.
      And the real update of the infra is normally done by Maintain. See later.

      As soon as the Infra exists, ie we have the identification of resources (mesos/docker/openstack/...), then we must use the infra flow to update itself.
      The update is done by Update and maintain is the apply.

      The update could be done by an infra team member, use the flow to propose the update and when accepted, a jenkins task (or not) should apply the change with forjj maintain from the infra repository source code.

      I expect the FORJJ parameters data should be stored in a yaml file that the driver should create/maintain.  */
    no_opts := map[string]interface{}{}
    required := map[string]interface{}{"required": true}
    ssh_dir_opts := map[string]interface{}{"default": fmt.Sprintf("%s/.ssh", os.Getenv("HOME"))}

    a.SetCommand("create", create_action_help)
    a.SetCmdArg("create", "workspace", create_orga_help, required)
    a.SetCmdFlag("create", ssh_dir_flag_name, create_ssh_dir_help, ssh_dir_opts)
    a.SetCmdFlag("create", "contrib-repo", create_contrib_help, no_opts)
    a.SetCmdFlag("create", "infra_url", create_infra_url_help, no_opts)
    a.SetCmdFlag("create", "infra_path", create_infra_path_help, no_opts)

    // Additional options will be loaded from the selected driver itself.
    // CI Drivers type parameters
    a.SetCmdFlag("create", ci_flag_name, create_ci_driver_help, no_opts)
    // GIT Upstream type Drivers parameters
    a.SetCmdFlag("create", us_flag_name, create_us_driver_help, no_opts)

    /********** UPDATE Action ************
      Update is not as clear as Create except that update should update the infra repository.

      We could probably use this update to add repositories or migrate the solution to a different place.  */
    a.SetCommand("update", update_action_help)
    a.SetCmdArg("update", "workspace", update_orga_help, required)
    a.SetCmdFlag("update", ssh_dir_flag_name, update_ssh_dir_help, ssh_dir_opts)
    a.SetCmdFlag("update", "contrib-repo", update_contrib_help, no_opts)
    // Additional options will be loaded from the selected driver itself.

    // CI Drivers type parameters
    a.SetCmdFlag("update", ci_flag_name, update_ci_driver_help, no_opts)

    // GIT Upstream type Drivers parameters
    a.SetCmdFlag("update", us_flag_name, update_us_driver_help, no_opts)

    /****** MAINTAIN Action ************
      Maintain is the real infra creation/update
      It has to ensure that the infra is updated as defined by the source code in $INFRA.

      The technology to ensure (orchestration), could be puppet/ansible combined with docker.
      This is not fully defined.  */
    a.SetCommand("maintain", maintain_action_help)
    a.SetCmdArg("maintain", "workspace", maintain_orga_help, required)
    a.SetCmdFlag("maintain", "infra_url", maintain_infra_url_help, no_opts)

    a.GetDriversFlags(os.Args[1:])

    if _, err := exec.LookPath("git"); err != nil {
        fmt.Printf("Unable to find 'git' command. %s Ensure it available in your PATH and retry.\n", err)
        os.Exit(1)
    }
}

//
// Generic Application function settings
//

func (a *Forj) GetActionOpts(cmd *kingpin.CmdClause) *ActionOpts {
    if v, found := a.Actions[cmd.FullCommand()]; found {
        return &v
    }
    kingpin.Fatalf("FORJJ Internal error. No matching '%s' in declared commands", cmd.FullCommand())
    return nil
}

// Function initializing driver flags with values.
// From values found in the commandline, extract them
// From forjj-* values, get it from Forjj internal data.
func (a *Forj) InitializeDriversFlag() {

    forjj_regexp, _ := regexp.Compile("forjj-(.*)")

    for service_type, driverOpts := range a.drivers {
        if driverOpts.Yaml.Name == "" {
            continue
        }

        gotrace.Trace("driver: '%s', command: '%s'\n", service_type, a.CurrentCommand.name)
        for _, command := range []string{"common", a.CurrentCommand.name} {
            gotrace.Trace(" From '%s' flags list\n", command)
            for flag_name, _ := range driverOpts.cmds[command].flags {
                gotrace.Trace("  Flag_name => '%s'\n", flag_name)
                forjj_vars := forjj_regexp.FindStringSubmatch(flag_name)
                if forjj_vars == nil {
                    if flag_value, ok := a.CurrentCommand.flagsv[flag_name]; ok {
                        a.drivers[service_type].cmds[command].flags[flag_name] = *flag_value
                        gotrace.Trace("   %s := %s\n", flag_name, *flag_value)
                    }
                } else {
                    flag_value := a.GetInternalData(forjj_vars[1])
                    a.drivers[service_type].cmds[command].flags[flag_name] = flag_value
                    gotrace.Trace("   forjj(%s) => %s := %s\n", forjj_vars[1], flag_name, flag_value)
                }
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
    }
    return
}

func (a *Forj) GetDriversParameters(cmd_args []string, cmd string) []string {

    for _, pluginOpts := range a.drivers {
        if cmd != "common" {
            cmd_args = append(cmd_args, fmt.Sprintf("--driver-%s %s", pluginOpts.driver_type, pluginOpts.Yaml.Name))
        }
        for k, v := range pluginOpts.cmds[cmd].flags {
            if v != "" {
                cmd_args = append(cmd_args, fmt.Sprintf("--%s", k), v)
            }
        }
    }
    return cmd_args
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
// TODO: In the context of a maintain/update, the context is first loaded from the workspace/infra repo.
func (a *Forj) LoadContext(args []string) (opts *ActionOpts) {
    context, err := a.app.ParseContext(args)
    if context == nil {
        kingpin.FatalIfError(err, "Application flags initialization issue. Driver flags issue?")
    }

    cmd := context.SelectedCommand
    if cmd == nil {
        return
    }

    opts = a.GetActionOpts(cmd)

    a.CurrentCommand = opts

    if debug_mode, found := a.flagValue(context, a.debug_f); found {
        // global debug defined in trace.go
        fmt.Printf("Debug set to '%s'.\n", debug_mode)
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
        fmt.Printf("Organization : '%s'\n", a.w.Organization)
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

    gotrace.Trace("Infrastructure repository defined : %s\n", a.w.Infra)

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

    // Identifying `ci` drivers options
    // The value is not set in flagsv. But is in the parser context.
    if value, found := a.flagValue(context, opts.flags[ci_flag_name]); found {
        a.drivers["ci"] = Driver{
            name:        value,
            driver_type: "ci",
            cmds: map[string]DriverCmdOptions{
                "common":   DriverCmdOptions{make(map[string]string), make(map[string]string)},
                "create":   DriverCmdOptions{make(map[string]string), make(map[string]string)},
                "update":   DriverCmdOptions{make(map[string]string), make(map[string]string)},
                "maintain": DriverCmdOptions{make(map[string]string), make(map[string]string)},
            },
        }
        fmt.Printf("Selected '%s' driver: %s\n", ci_flag_name, value)
    }

    // Identifying `git-us` drivers options
    // The value is not set in flagsv. But is in the parser context.
    if value, found := a.flagValue(context, opts.flags[us_flag_name]); found {
        a.drivers["upstream"] = Driver{
            name:        value,
            driver_type: "upstream",
            cmds: map[string]DriverCmdOptions{
                "common":   DriverCmdOptions{make(map[string]string), make(map[string]string)},
                "create":   DriverCmdOptions{make(map[string]string), make(map[string]string)},
                "update":   DriverCmdOptions{make(map[string]string), make(map[string]string)},
                "maintain": DriverCmdOptions{make(map[string]string), make(map[string]string)},
            },
        }
        fmt.Printf("Selected '%s' driver: %s\n", us_flag_name, value)
    }
    return
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
func (a *Forj) SetCmdFlag(cmd, name, help string, options map[string]interface{}) {
    arg := a.Actions[cmd].Cmd.Flag(name, help)

    if v, ok := options["required"]; ok && to_bool(v) {
        arg.Required()
    }
    if v, ok := options["default"]; ok {
        arg.Default(to_string(v))
    }
    if v, ok := options["hidden"]; ok && to_bool(v) {
        arg.Hidden()
    }

    a.Actions[cmd].flagsv[name] = arg.String()
    a.Actions[cmd].flags[name] = arg
}
