package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/trace"
	"log"
	"net/url"
	"forjj/utils"
	"os"
	"path"
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

	var action string

	// Load Forjfile templates in case of 'create' task.
	if cmds := c.GetCurrentCommand() ; cmds != nil && len(cmds) >= 1 {
		action = cmds[0].FullCommand()
	}
	switch action {
	case "create":
		// Detect and load a Forjfile template given.
		if err := a.LoadForjfile() ; err != nil {
			a.w.SetError(err)
			return nil, false
		}
	}

	// Define workspace
	a.setWorkspace() // failure test exit is made after parse time.

	// Load Workspace information if found
	a.w.Load()

	// Load Forjfile from infra repo, if found.
	a.LoadForge()

	// Set organization name to use.
	if err := a.set_organization_name() ; err != nil {
		return err, false
	}

	// Setting infra repository name
	if err := a.set_infra_name() ; err != nil {
		return err, false
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

	if v := a.cli.GetAction(cr_act).GetBoolAddr(no_maintain_f); v != nil {
		a.no_maintain = v
	}

	// Load drivers from repository
	a.prepare_registered_drivers()

	// TODO: Provide a caching feature if we keep loading from internet.
	gotrace.Trace("Loading drivers...")
	// Add drivers listed by the cli.
	for instance, d := range a.drivers {
		gotrace.Trace("Loading '%s'", instance)
		if err := a.load_driver_options(instance); err != nil {
			log.Printf("Unable to load plugin information for instance '%s'. %s", instance, err)
			continue
		}

		// Complete the driver information in cli records
		// The instance record has been created automatically with  cli.ForjObject.AddInstanceField()
		a.cli.SetValue(app, d.Name, cli.String, "type", d.DriverType)
		a.cli.SetValue(app, d.Name, cli.String, "driver", d.Name)

		d.Plugin.PluginSetWorkspace(a.w.Path())
		d.Plugin.PluginSetSource(path.Join(a.w.Path(), a.w.Infra.Name, "apps", d.DriverType))
		d.Plugin.PluginSocketPath(path.Join(a.w.Path(), "lib"))
	}
	return nil, true
}

func (a *Forj) set_infra_name() error {
	// Setting default if the organization is defined.
	if a.w.Organization != "" {
		// Set the 'infra' default flag value in cli
		a.cli.GetObject(infra).
			SetParamOptions(infra_name_f, cli.Opts().Default(fmt.Sprintf("%s-infra", a.w.Organization)))
	}

	infra_name, found, err := a.GetPrefs(infra_name_f)
	if err != nil {
		return err
	}

	if found {
		// Set the infra repo name to use
		// Can be set only the first time
		if a.w.Infra.Name == "" {
			// Get infra name from the flag
			a.w.Infra.Name = infra_name
			return a.SetPrefs(infra_name_f, a.w.Infra.Name) // Forjfile update
		}
		if infra_name != a.w.Infra.Name && a.w.Organization != "" {
			gotrace.Warning("You cannot update the Infra repository name from '%s' to '%s'.", a.w.Infra.Name, infra_name)
		}
		return nil
	}
	// Default infra-name
	if a.w.Organization != "" {
		// Use the default setting.
		a.w.Infra.Name = fmt.Sprintf("%s-infra", a.w.Organization)
		err = a.SetPrefs(infra_name_f, a.w.Infra.Name) // Forjfile update
	}
	return err
}

// set_organization_name Define the Organization name to use
// The organisation name can be defined from Forjfile or cli and will be stored in the workspace and the Forjfile in infra repo
// As soon as a workspace is defined (from a repo clone) the organization name could not be changed.
func (a *Forj) set_organization_name() error {
	orga, found, err := a.GetPrefs(orga_f)
	if err != nil {
		return err
	}
	if a.w.Organization == "" {
		if found && orga != "" {
			a.w.Organization = orga
		}
	} else {
		if found && a.w.Organization != orga {
			gotrace.Warning("Sorry, but you cannot update the organization name. The Forjfile will be updated back to '%s'", a.w.Organization)
		}
	}
	if a.w.Organization != "" {
		if err := a.SetPrefs(orga_f, a.w.Organization) ; err != nil {
			return err
		}
		log.Printf("Organization : '%s'", a.w.Organization)
		a.cli.GetObject(workspace).SetParamOptions(orga_f, cli.Opts().Default(a.w.Organization))
	} else {
		if a.w.Error() == nil {
			a.w.SetError(fmt.Errorf("No organization defined. Use --organization or add 'organization' to your Forjfile under 'forj-settings')"))
		}
	}
	return nil
}

// Initialize the workspace environment required by Forjj to work.
func (a *Forj) setWorkspace() {
	a.w.Init()

	infra_path, found, _, err := a.cli.GetStringValue(infra, "", infra_path_f)

	var workspace_path string
	if err != nil {
		gotrace.Trace("Unable to find '%s' value. %s Trying to detect it.", infra_path_f, err)
	}
	if !found {
		if pwd, e := os.Getwd() ; err != nil {
			gotrace.Error("%s", e)
			return
		} else {
			workspace_path = path.Join(pwd, ".forj-workspace")
		}
	} else {
		if p, err := utils.Abs(path.Join(infra_path, ".forj-workspace")); err == nil {
			workspace_path = p
		}
		gotrace.Trace("Using workspace '%s'", workspace_path)
	}

	a.w.SetPath(workspace_path)
}

// set_from_urlflag initialize a URL structure from a flag given.
// If the flag is set and valid, the URL will be stored in the given string address (store).
// if the flag has no value, store data is used as default.
// flag : Application flag value (from cli module)
//
// store : string address where this flag will be stored
func (a *Forj) set_from_urlflag(flag string, store *string) (*url.URL, error) {
	value, found, _, err := a.cli.GetStringValue(workspace, "", flag)
	if err != nil {
		gotrace.Trace("%s", err)
		return nil, err
	}

	if ! found {
		value = *store
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
