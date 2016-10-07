package main

import (
    "fmt"
    "github.com/alecthomas/kingpin"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "net/url"
    "log"
    "os"
)

// Load cli context to adapt the list of options/flags from the driver definition.
//
// It will
// - detect the debug mode
// - detect the organization name/path (to stored in app)
//   It will set the default Infra name.
// - detect the driver list source.
// - detect ci/us drivers name (to stored in app)
//
// - Load missing drivers information from forjj-options.yaml
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

    // load FORJJ workspace information
    a.setWorkspace(context , opts)

    // Load Workspace information
    a.w.Load()

    // Load Global Forjj options from infra repo, if found.
    a.LoadForjjOptions()

    // Set organization name to use.
    // Can be set only the first time
    if a.w.Organization == "" {
        if orga, found := a.flagValue(context, a.Orga_name_f); !found {
            a.Orga_name_f = a.Orga_name_f.Default(a.w.workspace)
            a.w.Organization = a.w.workspace
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
    if a.w.Infra.Name == "" {
        if infra, found := a.flagValue(context, a.infra_rep_f); found {
            // Get infra name from the flag
            a.w.Infra.Name = infra
        } else { // Or use the default setting.
            a.w.Infra.Name = fmt.Sprintf("%s-infra", a.w.Organization)
        }
    } else {
        if infra, found := a.flagValue(context, a.infra_rep_f); found && infra != a.w.Infra.Name {
            fmt.Printf("Warning!!! You cannot update the Infra repository name in an existing workspace.\n")
        }
    }

    gotrace.Trace("Infrastructure repository defined : %s", a.w.Infra.Name)

    // Identifying appropriate Contribution Repository.
    // The value is not set in flagsv. But is in the parser context.
    opts.set_from_urlflag(a, context, "contribs-repo", a.ContribRepo_uri, &a.w.Contrib_repo_path)
    opts.set_from_urlflag(a, context, "flows-repo", a.FlowRepo_uri, &a.w.Flow_repo_path)
    opts.set_from_urlflag(a, context, "repotemplates-repo", a.RepotemplateRepo_uri, &a.w.Repotemplate_repo_path)

    // Getting list of drivers (--apps)
    a.drivers_list.list = make(map[string]DriverDef)
    a.drivers_list.GetDriversFromContext(context, a.c_drivers_list_f)
    a.drivers_list.GetDriversFromContext(context, a.u_drivers_list_f)

    // Read forjj infra file and the options --file given, defined by create/update driver flags settings saved or not
    // This load Maintain context required by plugins. Maintain has limited flags to provide at runtime. Everything, except credentials should be stored in the infra-repo and workspace. Credentials is given with the --file option in yaml format.
    file_desc, _ := a.flagValue(context, a.creds_file_f)
    if err := a.LoadForjjPluginsOptions(file_desc) ; err != nil {
        gotrace.Trace("Warning! Options files were not loaded. %s", err)
    }

}

// Initialize the workspace environment required by Forjj to work.
func (a *Forj) setWorkspace(context *kingpin.ParseContext, opts *ActionOpts) {
    // The value is not set in argsv. But is in the parser context.
    var orga_path string
    var found bool
    var err error

    if opts.name == "create" {
        orga_path, found = a.argValue(context, opts.args["workspace"])
    } else {
        orga_path, found = a.flagValue(context, opts.flags["ws"])
    }

    if !found {
        if v := os.Getenv("FORJJ_WORKSPACE") ; v != "" {
            orga_path = v
            found = true
        }
    }

    if ! found {
        orga_path, err = a.w.DetectIt()
        kingpin.FatalIfError(err, "Unable to find the workspace from current directory, FORJJ_WORKSPACE or cli. please define one to create it.")
    }

    a.w.Init(orga_path)
}

type validateHdlr func(string) error

// Set a string variable pointer with value found in cli context.
func (o *ActionOpts)set_from_flag(a *Forj, context *kingpin.ParseContext, flag string, store *string, val_fcnt validateHdlr) error {
    if d, found := a.flagValue(context, o.flags[flag]) ; found {
        if val_fcnt != nil {
            if err := val_fcnt(d) ; err != nil {
                return err
            }
        }
        *store = d
    }
    return nil
}

func (o *ActionOpts)set_from_urlflag(a *Forj, context *kingpin.ParseContext, flag string, theurl *url.URL, store *string)  {
    value := ""

    if _, found := o.flags[flag] ; ! found {
        gotrace.Trace("No '%s' flag found in cli context. Used stored value '%s'", flag, *store)
        value = *store
    } else {
        if d, found := a.flagValue(context, o.flags[flag]) ; found {
            value = d
        } else {
            if o.flags[flag].HasEnvarValue() {
                gotrace.Trace("Getting value from env for flag '%s'", flag)
                value = o.flags[flag].GetEnvarValue()
            }
        }
    }

    if u, err := url.Parse(value); err != nil {
        log.Printf("%s", err)
    } else {
        *theurl = *u
        if u.Scheme == "" {
            *store = value
        }
    }
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
    a.Actions[name] = &ActionOpts{
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

    if v, ok := options["envar"]; ok {
        arg.Envar(to_string(v))
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

func SetBoolFlag(flag *kingpin.FlagClause) *bool {
    return flag.Bool()
}
