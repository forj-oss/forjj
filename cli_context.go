package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/trace"
	"log"
	"net/url"
	"path"
	"os"
	"forjj/utils"
	"strings"
)

const Workspace_Name = ".forj-workspace"

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
	} else {
		return nil, false
	}
	switch action {
	case "create":
		// Detect and load a Forjfile template given.
		if err := a.LoadForjfile(action) ; err != nil {
			a.w.SetError(err)
			return nil, false
		}
	}

	// Define workspace
	if err := a.setWorkspace() ; err != nil {
		// failure test exit is made after parse time.
		return err, false
	}

	// Load Workspace information if found
	a.w.Load()

	// Load Forjfile from infra repo, if found.
	if err := a.LoadForge() ; err != nil {
		if action != "create" {
			return fmt.Errorf("Forjfile not loaded. %s", err), false
		}
		gotrace.Warning("%s", err)
	}

	// Set organization name to use.
	if err := a.set_organization_name() ; err != nil {
		return err, false
	}

	// Setting infra repository name
	if err := a.set_infra_name(action) ; err != nil {
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

	if file_desc, err := a.cli.GetAppStringValue(cred_f); err == nil && file_desc != "" {
		a.s.SetFile(file_desc)
	} else {
		a.s.SetPath(a.w.Path())
	}
	if err := a.s.Load(); err != nil {
		gotrace.Info("Credential file were not loaded. %s", err)
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

	if i, err := a.cli.GetAppStringValue(debug_instance_f); err == nil && i != "" {
		a.debug_instances = strings.Split(i, ",")
	}
	a.contextDisplayed()
	return nil, true
}

func (a *Forj) set_infra_name(action string) (err error) {
	defer a.f.SetInfraAsRepo()
	// Setting default if the organization is defined.
	if a.w.Organization != "" {
		// Set the 'infra' default flag value in cli
		a.cli.GetObject(infra).
			SetParamOptions(infra_name_f, cli.Opts().Default(fmt.Sprintf("%s-infra", a.w.Organization)))
	}

	var infra_name string
	var found bool

	if action == maint_act {
		infra_name, found, err = a.GetForgePrefs(infra_name_f)
	} else {
		infra_name, found, err = a.GetPrefs(infra_name_f)
	}
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
		return a.SetPrefs(infra_name_f, a.w.Infra.Name)
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
func (a *Forj) setWorkspace() error {
	// Ask to not save some entries, like 'infra-path' in the workspace file.
	a.w.Init(infra_path_f)

	infra_path, found, err := a.GetLocalPrefs(infra_path_f)

	var workspace_path string
	if err != nil {
		gotrace.Trace("Unable to find '%s' value. %s Trying to detect it.", infra_path_f, err)
	}
	if !found {
		if pwd, e := os.Getwd() ; err != nil {
			return e
		} else {
			workspace_path = path.Join(pwd, Workspace_Name)
		}
	} else {
		if p, err := utils.Abs(path.Join(infra_path, Workspace_Name)); err == nil {
			workspace_path = p
		}
		gotrace.Trace("Using workspace '%s'", workspace_path)
	}

	return a.w.SetPath(workspace_path)
}

// set_from_urlflag initialize a URL structure from a flag given.
// If the flag is set from cli and valid, the URL will be stored in the given string address (store).
// if the flag has no value, store data is used as default.
// flag : Application flag value (from cli module)
//
// store : string address where this flag will be stored
func (a *Forj) set_from_urlflag(flag string, store *string) (u *url.URL, e error) {
	value, found, def, err := a.cli.GetStringValue(workspace, "", flag)
	if err != nil {
		gotrace.Trace("%s", err)
		return nil, err
	}

	// cli define an url
	if found && !def {
		if u, e = url.Parse(value); e == nil {
			*store = u.String()
			return
		}
	}

	// no cli definition. Use the stored url.
	if *store != "" {
		if u, e = url.Parse(*store); e == nil {
			return
		}
	}

	// no cli, neither stored one. Use cli default value if exist.
	if found && def{
		if u, e = url.Parse(value); e == nil {
			*store = u.String()
			return
		}
	}

	// Not found return nil with last error detected
	return
}
