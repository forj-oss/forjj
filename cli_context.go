package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/trace"
	"log"
	"net/url"
	"os"
)

// ParseContext : Load cli context to adapt the list of options/flags from the driver definition.
//
// It will
// - detect the organization name/path (to stored in app)
//   It will set the default Infra name.
// - detect the driver list source.
// - detect ci/us drivers name (to stored in app)
//
// - Load missing drivers information from forjj-options.yaml
func (a *Forj) ParseContext(c *cli.ForjCli, _ interface{}) error {
	// load FORJJ workspace information
	if err := a.setWorkspace(); err != nil {
		return fmt.Errorf("Unable to define workspace from context. %s", err)
	}

	// Load Workspace information if found
	a.w.Load()

	// Load Global Forjj options from infra repo, if found.
	a.LoadForjjOptions()

	w_o := c.GetObject(workspace)
	// Set organization name to use.
	// Can be set only the first time
	if f, found, _ := c.GetStringValue(workspace, "", orga_f); found {
		if a.w.Organization == "" {
			w_o.SetParamOptions(orga_f, cli.Opts().Default(a.w.workspace))
			a.w.Organization = a.w.workspace
		} else {
			if f != a.w.Organization {
				fmt.Print("Warning!!! You cannot update the organization name in an existing workspace.\n")
			}
		}
	} else {
		a.w.Organization = f
	}

	if f, found, _ := c.GetStringValue(workspace, "", infra_f); found {
		if a.w.Organization != "" {
			log.Printf("Organization : '%s'", a.w.Organization)
			// Set the 'infra' default flag value
			w_o.SetParamOptions(infra_f, cli.Opts().Default(fmt.Sprintf("%s-infra", a.w.Organization)))
		}
		// Set the infra repo name to use
		// Can be set only the first time
		if a.w.Infra.Name == "" {
			// Get infra name from the flag
			a.w.Infra.Name = f
		} else {
			if f != a.w.Infra.Name {
				fmt.Print("Warning!!! You cannot update the Infra repository name in an existing workspace.\n")
			}
		}
	} else {
		// Use the default setting.
		a.w.Infra.Name = fmt.Sprintf("%s-infra", a.w.Organization)
	}

	gotrace.Trace("Infrastructure repository defined : %s", a.w.Infra.Name)

	// Identifying appropriate Contribution Repository.
	// The value is not set in flagsv. But is in the parser context.
	if v, err := a.set_from_urlflag("contribs-repo", &a.w.Contrib_repo_path); err != nil {
		gotrace.Trace("%s", err)
	} else {
		a.ContribRepo_uri = v
	}
	if v, err := a.set_from_urlflag("flows-repo", &a.w.Flow_repo_path); err != nil {
		gotrace.Trace("%s", err)
	} else {
		a.FlowRepo_uri = v
	}
	if v, err := a.set_from_urlflag("repotemplates-repo", &a.w.Repotemplate_repo_path); err != nil {
		gotrace.Trace("%s", err)
	} else {
		a.RepotemplateRepo_uri = v
	}

	// Getting list of drivers (--apps) - Obsolete
	/*    a.drivers_list.list = make(map[string]DriverDef)
	      a.drivers_list.GetDriversFromContext(context, a.cli.C_drivers_list_f)
	      a.drivers_list.GetDriversFromContext(context, a.cli.U_drivers_list_f)*/

	// Read forjj infra file and the options --file given, defined by create/update driver flags settings saved or not
	// This load Maintain context required by plugins. Maintain has limited flags to provide at runtime. Everything, except credentials should be stored in the infra-repo and workspace. Credentials is given with the --file option in yaml format.
	if file_desc, err := a.cli.GetAppStringValue(cred_f); err == nil {
		if err := a.LoadForjjPluginsOptions(file_desc); err != nil {
			gotrace.Trace("Warning! Options files were not loaded. %s", err)
		}
	}
	return nil
}

// Initialize the workspace environment required by Forjj to work.
func (a *Forj) setWorkspace() error {
	// The value is not set in argsv. But is in the parser context.
	var orga_path string
	var found bool
	var err error

	/* orga_path, found = a.cli.GetValue(workspace) */
	orga_path, found, _ = a.cli.GetStringValue(workspace, "", orga_f)

	if !found {
		if v := os.Getenv("FORJJ_WORKSPACE"); v != "" {
			orga_path = v
			found = true
		}
	}

	if !found {
		orga_path, err = a.w.DetectIt()
		a.w.error = fmt.Errorf("Unable to find the workspace from current directory, FORJJ_WORKSPACE or cli. "+
			"please define one to create it. %s", err)
		return nil
	}

	a.w.Init(orga_path)
	return nil
}

// type validateHdlr func(string) error

// Set a string variable pointer with value found in cli context.
/*func (o *ActionOpts) set_from_flag(a *Forj, context *kingpin.ParseContext, flag string, store *string, val_fcnt validateHdlr) error {
	if d, found := a.flagValue(context, o.flags[flag]); found {
		if val_fcnt != nil {
			if err := val_fcnt(d); err != nil {
				return err
			}
		}
		*store = d
	}
	return nil
}*/

// set_from_urlflag initialize a URL structure from a flag given.
// If the flag is set and valid, the URL will be stored in the given string address (store).
//
// flag : Application flag value (from cli module)
//
// store : string address where this flag will stored
func (a *Forj) set_from_urlflag(flag string, store *string) (*url.URL, error) {
	value, err := a.cli.GetAppStringValue(flag)
	if err != nil {
		return nil, err
	}

	if u, err := url.Parse(value); err != nil {
		return nil, err
	} else {
		if u.Scheme == "" {
			*store = value
		}
		return u, nil
	}
}

/*func (*Forj) argValue(context *kingpin.ParseContext, f *kingpin.ArgClause) (value string, found bool) {
	for _, element := range context.Elements {
		if flag, ok := element.Clause.(*kingpin.ArgClause); ok && flag == f {
			value = *element.Value
			found = true
			return
		}
	}
	return
}*/

/*func (*Forj) flagValue(context *kingpin.ParseContext, f *kingpin.FlagClause) (value string, found bool) {
	for _, element := range context.Elements {
		if flag, ok := element.Clause.(*kingpin.FlagClause); ok && flag == f {
			value = *element.Value
			found = true
			return
		}
	}
	return
}*/

// Set an application command
/*func (a *Forj) SetCommand(name, help string) {
	a.Actions[name] = &ActionOpts{
		name:   name,
		Cmd:    a.cli.App.Command(name, help),
		flags:  make(map[string]*kingpin.FlagClause),
		flagsv: make(map[string]*string),
		args:   make(map[string]*kingpin.ArgClause),
		argsv:  make(map[string]*string),
	}
}*/

// Set a command argument
/*func (a *Forj) SetCmdArg(cmd, name, help string, options map[string]interface{}) {
	arg := a.Actions[cmd].Cmd.Arg(name, help)

	if v, ok := options["required"]; ok && to_bool(v) {
		arg.Required()
	}
	if v, ok := options["default"]; ok {
		arg.Default(to_string(v))
	}

	a.Actions[cmd].argsv[name] = arg.String()
	a.Actions[cmd].args[name] = arg
}*/

// Set a Command flag.
/*func (a *Forj) SetCmdFlag(cmd, name, help string, options map[string]interface{}) (arg *kingpin.FlagClause) {
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
}*/

/*func SetBoolFlag(flag *kingpin.FlagClause) *bool {
	return flag.Bool()
}*/
