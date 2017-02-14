package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/trace"
	"log"
	"net/url"
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
func (a *Forj) ParseContext(c *cli.ForjCli, _ interface{}) (error, bool) {
	gotrace.Trace("Setting FORJJ Context...")
	// load FORJJ workspace information
	a.setWorkspace() // failure test exit is made after parse time.

	// Load Workspace information if found
	a.w.Load()

	// Load Global Forjj options from infra repo, if found.
	a.LoadForjjOptions()

	w_o := c.GetObject(workspace)
	// Set organization name to use.
	// Can be set only the first time
	if f, found, _, _ := c.GetStringValue(workspace, "", orga_f); !found {
		if a.w.Organization == "" && a.w.workspace != "" {
			a.w.Organization = a.w.workspace
		}
		if a.w.Organization != "" {
			w_o.SetParamOptions(orga_f, cli.Opts().Default(a.w.Organization))
		}
	} else {
		if f == "" {
			f = a.w.workspace
		}
		if a.w.Organization == "" {
			a.w.Organization = f
		} else {
			if f != a.w.Organization {
				fmt.Print("Warning!!! You cannot update the organization name in an existing workspace.\n")
			}
		}
	}

	if a.w.Organization != "" {
		log.Printf("Organization : '%s'", a.w.Organization)
	} else {
		if a.w.error == nil {
			a.w.error = fmt.Errorf("No organization defined.")
		}
	}

	i_o := c.GetObject(infra)
	if f, found, isDefault, _ := c.GetStringValue(infra, "", infra_f); found {
		if isDefault {
			if a.w.Organization != "" {
				// Set the 'infra' default flag value
				i_o.SetParamOptions(infra_f, cli.Opts().Default(fmt.Sprintf("%s-infra", a.w.Organization)))
				f, _, _, _ = c.GetStringValue(infra, "", infra_f)
			}
		}

		// Set the infra repo name to use
		// Can be set only the first time
		if a.w.Infra.Name == "" {
			// Get infra name from the flag
			a.w.Infra.Name = f
		} else {
			if f != a.w.Infra.Name && !isDefault {
				fmt.Print("Warning!!! You cannot update the Infra repository name in an existing workspace.\n")
			}
		}
	} else {
		if a.w.Organization != "" {
			// Use the default setting.
			a.w.Infra.Name = fmt.Sprintf("%s-infra", a.w.Organization)
		}
	}

	gotrace.Trace("Infrastructure repository defined : %s (organization: %s)", a.w.Infra.Name, a.w.Organization)

	// Identifying appropriate Contribution Repository.
	// The value is not set in flagsv. But is in the parser context.
	if v, err := a.set_from_urlflag("contribs-repo", &a.w.Contrib_repo_path); err == nil {
		a.ContribRepo_uri = v
	}
	if v, err := a.set_from_urlflag("flows-repo", &a.w.Flow_repo_path); err == nil {
		a.FlowRepo_uri = v
	}
	if v, err := a.set_from_urlflag("repotemplates-repo", &a.w.Repotemplate_repo_path); err == nil {
		a.RepotemplateRepo_uri = v
	}

	// Read forjj infra file and the options --file given, defined by create/update driver flags settings saved or not
	// This load Maintain context required by plugins. Maintain has limited flags to provide at runtime. Everything, except credentials should be stored in the infra-repo and workspace. Credentials is given with the --file option in yaml format.
	if file_desc, err := a.cli.GetAppStringValue(cred_f); err == nil {
		if err := a.LoadForjjPluginsOptions(file_desc); err != nil {
			gotrace.Trace("Warning! Options files were not loaded. %s", err)
		}
	}

	if v := a.cli.GetAction(cr_act).GetBoolAddr("no-maintain"); v != nil {
		a.no_maintain = v
	}
	return nil, true
}

// Initialize the workspace environment required by Forjj to work.
func (a *Forj) setWorkspace() {
	// The value is not set in argsv. But is in the parser context.
	var orga_path string
	var found bool
	var err error

	a.w.Init()

	orga_path, found, _, err = a.cli.GetStringValue(workspace, "", workspace)

	if err != nil {
		gotrace.Trace("Unable to find '%s' value. %s Trying to detect it.", workspace, err)
	}
	if !found {
		orga_path, err = a.w.DetectIt()
		a.w.error = fmt.Errorf("Unable to find the workspace from current directory, FORJJ_WORKSPACE or --workspace. "+
			"please define one to create it. %s", err)
		return
	} else {
		if p, err := Abs(orga_path); err == nil {
			orga_path = p
		}
		gotrace.Trace("Found workspace setting : '%s'", orga_path)
	}

	a.w.SetPath(orga_path)
}

// set_from_urlflag initialize a URL structure from a flag given.
// If the flag is set and valid, the URL will be stored in the given string address (store).
//
// flag : Application flag value (from cli module)
//
// store : string address where this flag will stored
func (a *Forj) set_from_urlflag(flag string, store *string) (*url.URL, error) {
	value, _, _, err := a.cli.GetStringValue(workspace, "", flag)
	if err != nil {
		gotrace.Trace("%s", err)
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
