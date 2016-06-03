package main

import (
        "fmt"
        "os"
        "gopkg.in/alecthomas/kingpin.v2"
//        "github.com/alecthomas/kingpin"
        "net/url"
        "path"
)

var debug *bool

// TODO: Support multiple contrib sources.
// TODO: Add flag for branch name to ensure local git branch is correct.

// Struct for args/flags for an action
type ActionOpts struct {
  name string                          // Action name
  flags map[string]*kingpin.FlagClause // list of additional flags loaded.
  flagsv map[string]*string            // list of additional flags value loaded.
  args map[string]*kingpin.ArgClause   // List of Arguments by name
  argsv map[string]*string             // List of Arguments value by name
  Cmd *kingpin.CmdClause               // Command object
}

type DriverOptions struct {
  name string                     // driver name
  driver_type string              // driver type name
  flags map[string]string         // list of flags values
  args map[string]string          // list of args values
}

type Forj struct {
  Organization string             // Organization name
  Orga_path string                // Organization directory path.
  ContribRepo_uri *url.URL        // URL to github raw files
  contrib_repo_path string        // Contribution repository Path
  Branch string                   // branch name
  app *kingpin.Application        // Kingpin Application object
  Actions map[string]ActionOpts   // map of Commands with their arguments/flags
  drivers map[string]DriverOptions// List of drivers type to use, defined by cli.
  CurrentCommand *ActionOpts      // Loaded CurrentCommand reference.
}

const (
  ci_flag_name = "ci"
  us_flag_name = "git-us"
  ssh_dir_flag_name = "ssh-dir"
)

//
// Define application cli options
//

func (a *Forj) init() {
 a.app = kingpin.New(os.Args[0], app_help).UsageTemplate(DefaultUsageTemplate)
 debug = a.app.Flag("debug", "Enable debug mode.").Bool()
 a.app.Version("forjj V0.0.1 (POC)").Author("Christophe Larsonneur <christophe.larsonneur@hpe.com>")

 u, _ := url.Parse("https://github.hpe.com/forj/forjj-contribs/raw")

 a.ContribRepo_uri = u
 a.Branch = "master"
 a.drivers = make(map[string]DriverOptions)
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
 no_opts := map[string]interface{}      {}
 required := map[string]interface{}     {"required": true}
 ssh_dir_opts := map[string]interface{} {"default" : fmt.Sprintf("%s/.ssh", os.Getenv("HOME"))}
 infra_opts := map[string]interface{}   {"default" : "<organization>-infra"}

 a.SetCommand("create", create_action_help)
 a.SetCmdArg( "create", "organization",    create_orga_help,       required)
 a.SetCmdFlag("create", ssh_dir_flag_name, create_ssh_dir_help,    ssh_dir_opts)
 a.SetCmdFlag("create", "contrib-repo",    create_contrib_help,    no_opts)
 a.SetCmdFlag("create", "infra",           create_infra_name_help, infra_opts)
 a.SetCmdFlag("create", "infra_url",       create_infra_url_help,  no_opts)
 a.SetCmdFlag("create", "infra_path",      create_infra_path_help, no_opts)

 // Additional options will be loaded from the selected driver itself.
 // CI Drivers type parameters
 a.SetCmdFlag("create", ci_flag_name,   create_ci_driver_help,  no_opts)
 // GIT Upstream type Drivers parameters
 a.SetCmdFlag("create", us_flag_name,   create_us_driver_help,  no_opts)

 /********** UPDATE Action ************ 
  Update is not as clear as Create except that update should update the infra repository.

  We could probably use this update to add repositories or migrate the solution to a different place.  */
 a.SetCommand("update", update_action_help)
 a.SetCmdArg( "update", "organization",    update_orga_help,        required)
 a.SetCmdFlag("update", ssh_dir_flag_name, update_ssh_dir_help,     ssh_dir_opts)
 a.SetCmdFlag("update", "contrib-repo",    update_contrib_help,     no_opts)
 // Additional options will be loaded from the selected driver itself.

 // CI Drivers type parameters
 a.SetCmdFlag("update", ci_flag_name,   update_ci_driver_help,   no_opts)

 // GIT Upstream type Drivers parameters
 a.SetCmdFlag("update", us_flag_name,   update_us_driver_help,   no_opts)

 /****** MAINTAIN Action ************ 
  Maintain is the real infra creation/update 
  It has to ensure that the infra is updated as defined by the source code in $INFRA.

  The technology to ensure (orchestration), could be puppet/ansible combined with docker.
  This is not fully defined.  */
 a.SetCommand("maintain", maintain_action_help)
 a.SetCmdArg( "maintain", "organization",  maintain_orga_help,    required)

 a.GetDriversFlags(os.Args[1:])
}

//
// Generic Application function settings
//
func (a *Forj)GetActionOpts(cmd *kingpin.CmdClause) *ActionOpts {
 if v, found := a.Actions[cmd.FullCommand()] ; found { return &v }
 kingpin.Fatalf("FORJJ Internal error. No matching '%s' in declared commands", cmd.FullCommand())
 return nil
}

// Load cli context to adapt the list of options/flags from the driver definition.
//
// It will 
// - detect the organization name/path (to stored in app)
// - detect the driver list source.
// - detect ci/us drivers name (to stored in app)
func (a *Forj)LoadContext(args []string) (opts *ActionOpts) {
 context, err := a.app.ParseContext(args)
 kingpin.FatalIfError(err, "Issue in parsing '%s'", args)

 cmd := context.SelectedCommand
 if cmd == nil { return }

 opts = a.GetActionOpts(cmd)

 a.CurrentCommand = opts

 // The value is not set in argsv. But is in the parser context.
 if orga, found := a.argValue(context, opts.args["organization"]) ; found {
    // TODO: Test the path given.
    // TODO: Expose so internal data like Organization as --forj-organization to expose it to any plugins that request it.
    a.Organization = path.Base(orga)
    a.Orga_path = path.Dir(orga)
    fmt.Printf("Organization '%s' found in '%s' (given: %s)\n", a.Organization, a.Orga_path, orga)
 }

 // Identifying appropriate Contribution Repository.
 // The value is not set in flagsv. But is in the parser context.
 if contrib_repo, found := a.flagValue(context, opts.flags["contrib-repo"]) ; found {
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
 if value, found := a.flagValue(context, opts.flags[ci_flag_name]) ; found {
   a.drivers["ci"] = DriverOptions{ value, "ci", make(map[string]string), make(map[string]string) }
   fmt.Printf("Selected '%s' driver: %#v\n", ci_flag_name, value)
 }

 // Identifying `git-us` drivers options
 // The value is not set in flagsv. But is in the parser context.
 if value, found := a.flagValue(context, opts.flags[us_flag_name]) ; found {
   a.drivers["upstream"] = DriverOptions{ value, "upstream", make(map[string]string), make(map[string]string) }
   fmt.Printf("Selected '%s' driver: %#v\n", us_flag_name, value)
 }
 return
}

func (*Forj)argValue(context *kingpin.ParseContext, f *kingpin.ArgClause) (value string, found bool){
 for _, element := range context.Elements {
     if flag, ok := element.Clause.(*kingpin.ArgClause); ok && flag == f {
        value = *element.Value
        found = true
        return
     }
  }
 return
}

func (*Forj)flagValue(context *kingpin.ParseContext, f *kingpin.FlagClause) (value string, found bool){
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
func (a *Forj)SetCommand(name, help string) {
 a.Actions[name] = ActionOpts{
   name: name,
   Cmd: a.app.Command(name, help),
   flags: make(map[string]*kingpin.FlagClause),
   flagsv: make(map[string]*string),
   args: make(map[string]*kingpin.ArgClause),
   argsv: make(map[string]*string),
 }
}

// Set a command argument
func (a *Forj)SetCmdArg(cmd, name, help string, options map[string]interface{}) {
 arg := a.Actions[cmd].Cmd.Arg(name, help)

 if v, ok := options["required"] ; ok && to_bool(v) { arg.Required() }
 if v, ok := options["default"] ; ok { arg.Default(to_string(v)) }

 a.Actions[cmd].argsv[name] = arg.String()
 a.Actions[cmd].args[name] = arg
}

// Set a Command flag.
func (a *Forj)SetCmdFlag(cmd, name, help string, options map[string]interface{}) {
 arg := a.Actions[cmd].Cmd.Flag(name, help)

 if v, ok := options["required"] ; ok && to_bool(v) { arg.Required() }
 if v, ok := options["default"] ; ok { arg.Default(to_string(v)) }

 a.Actions[cmd].flagsv[name] = arg.String()
 a.Actions[cmd].flags[name] = arg
}
