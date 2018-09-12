package main

import (
	"fmt"
	"forjj/creds"
	"forjj/utils"
	"log"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/trace"
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
//
// Warning! Return the error if you want to abort the cli process.
// Otherwise, use a.w.SetError and `return nil, false`
//
// This behavior is required to let cli execute the --help if needed.
// if an error is returned, the help if selected will never been displayed...
func (a *Forj) ParseContext(c *cli.ForjCli, _ interface{}) (error, bool) {
	gotrace.Trace("Setting FORJJ Context...")

	// Load Forjfile models in case of 'create' task.
	if cmds := c.GetCurrentCommand(); cmds != nil && len(cmds) >= 1 {
		a.contextAction = cmds[0].FullCommand()
	} else {
		return nil, false
	}

	a.secrets.context.defineContext(c.GetParseContext())
	if a.contextAction == cr_act || a.contextAction == val_act {
		// Detect and load a Forjfile model given.
		if err := a.LoadForjfile(a.contextAction); err != nil {
			a.w.SetError(err)
			return nil, false
		}
	}

	// Define workspace
	if err := a.setWorkspace(); err != nil {
		// failure test exit is made after parse time.
		return err, false
	}

	// Load Workspace information if found
	a.w.Load()

	// Read definition file from repo.
	is_valid_action := (utils.InStringList(a.contextAction, val_act, cr_act, upd_act, maint_act, add_act, rem_act, ren_act, chg_act, list_act) != "")
	need_to_create := (a.contextAction == cr_act)
	need_to_update := (a.contextAction == upd_act)
	need_to_validate := (a.contextAction == val_act)
	if err := a.f.SetInfraPath(a.w.InfraPath(), is_valid_action && (need_to_create || need_to_validate)); err != nil {
		a.w.SetError(err)
		return nil, false
	}

	// Load Forjfile from infra repo, if found.
	if err := a.LoadForge(); err != nil {
		if utils.InStringList(a.contextAction, upd_act, maint_act, add_act, rem_act, ren_act, chg_act, list_act) != "" {
			a.w.SetError(fmt.Errorf("Forjfile not loaded. %s", err))
			return nil, false
		}
		gotrace.Warning("%s", err)
	}

	if !a.f.IsLoaded() {
		a.w.SetError(fmt.Errorf("No Forjfile or Forjfile model were loaded"))
		return nil, false
	}

	// By default, with `forjj update`, the deploy source code generated by forjj are generated in a repo beside the infra cloned repo.
	// This is the default use case from Developer side. And Only DEV type can manage the beside infra repository.
	// TEST/PRO always uses the internal workspace.
	//
	// From CI, we usually test and deploy. Then, the generated deploy source code requires to be published automatically.
	// So, we need to add --deploy-publish.
	// In this case, forjj uses his internal workspace to update, commit and push, in one shot.

	// TODO: Move this code in forjfile/forge.go
	// Setup each deployment internal data
	deployPath := path.Join(a.w.Path(), "deployments")
	deployPublish, _, _ := a.cli.GetBoolValue("_app", "forjj", "deploy-publish")
	for name, deploy := range a.f.GetDeployments() {
		deploy.DeploymentCoreStruct.GitSetRepo(deployPath, "")

		if deploy.Type == "DEV" && !deployPublish && need_to_update {
			devRepoWS := deploy.GetRepoPath()
			devRepoAside := path.Join(path.Dir(a.f.InfraPath()), name)
			os.Remove(devRepoAside)

			if err := os.Symlink(devRepoWS, devRepoAside); err != nil {
				gotrace.Error("Unable to create link to %s in %s. \n%s", devRepoWS, devRepoAside, err)
			}
		}
	}

	// Define the current deployment in create mode.
	if need_to_create || need_to_validate {
		// TODO: Be able to choose another deployment than the PRO one in create phase.

		// Determine the default PRO deployment
		if v, err := a.f.GetDeploymentPROType(); err != nil {
			return err, false
		} else {
			gotrace.Info("Using production deployment: '%s'.", v.Name())
			a.f.SetDeployment(v.Name())
			a.d = &v.DeploymentCoreStruct
		}
	} else {
		deployTo := a.f.GetDeployment()
		// Setup each deployment internal data
		if v, found := a.f.GetADeployment(deployTo); found {
			// Define selected deployment.
			a.d = &v.DeploymentCoreStruct
			gotrace.Info("Using %s deployment.", v.Name())
		} else {
			a.w.SetError(fmt.Errorf("Unknown deployment environment '%s'. Use one defined in your Forjfile", deployTo))
			return nil, false
		}

	}

	if a.f.GetDeployment() == "global" && (utils.InStringList(a.contextAction, val_act, cr_act, upd_act, maint_act) != "") {
		return fmt.Errorf("'global' is not a valid deployment environment"), false
	}

	// Build in memory representation from source files loaded.
	if err := a.f.BuildForjfileInMem(); err != nil {
		return err, false
	}

	// Set organization name to use.
	if err := a.set_organization_name(); err != nil {
		a.w.SetError(err)
		return nil, false
	}

	// Setting infra repository name
	if err := a.setInfraName(a.contextAction); err != nil {
		a.w.SetError(err)
		return nil, false
	}

	gotrace.Trace("Infrastructure repository defined : %s (organization: %s)", a.w.Infra().Name, a.w.GetString("organization"))

	// Identifying appropriate Contribution Repository.
	// The value is not set in flagsv. But is in the parser context.

	a.ContribRepoURIs = make([]*url.URL, 0, 1)

	if v, err := a.setFromURLFlag("contribs-repo", func(f, v string) (updated bool) {
		return a.w.Set(f, v, need_to_create)
	}); err == nil {
		a.ContribRepoURIs = append(a.ContribRepoURIs, v)
		gotrace.Trace("Using '%s' for '%s'", v, "contribs-repo")
	} else {
		return fmt.Errorf("Contribs repository url issue: %s", err), false
	}
	if v, err := a.setFromURLFlag("flows-repo", func(f, v string) (updated bool) {
		return a.w.Set(f, v, need_to_create)
	}); err == nil {
		a.flows.AddRepoPath(v)
		vpath, _ := url.PathUnescape(v.String())
		gotrace.Trace("Using '%s' for '%s'", vpath, "flows-repo")
	} else {
		gotrace.Error("Flow repository url issue: %s", err)
	}
	if v, err := a.setFromURLFlag("repotemplates-repo", func(f, v string) (updated bool) {
		return a.w.Set(f, v, need_to_create)
	}); err == nil {
		a.RepotemplateRepo_uri = v
		gotrace.Trace("Using '%s' for '%s'", v, "repotemplates-repo")
	} else {
		gotrace.Error("RepoTemplates repository url issue: %s", err)
	}

	// Credential management
	a.s.InitEnvDefaults(a.w.Path(), a.f.GetDeployment())
	if fileDesc, err := a.cli.GetAppStringValue(cred_f); err == nil && fileDesc != "" {
		a.s.SetFile(a.f.GetDeployment(), fileDesc)
	}

	if err := a.s.EncryptAll(true); err != nil {
		return fmt.Errorf("Unable to encrypt files. %s", err), false
	}
	if err := a.s.Load(); err != nil {
		return fmt.Errorf("Some credential files were not loaded. %s", err), false
	}

	if err := a.s.Upgrade(func(d *creds.Secure, version string) (err error) {
		// Function to identify creds V0 and do upgrade
		if deployObj, err := a.f.GetDeploymentPROType(); err != nil {
			return fmt.Errorf("Unable to upgrade. %s", err)
		} else if credPath := d.DirName(creds.Global); credPath == a.w.Path() { // In Workspace
			oldFile := path.Join(credPath, creds.DefaultCredsFile)
			if d.Version(creds.Global) == "V0" && d.Version(deployObj.Name()) == "" { // global loaded, but missing deploy one ie version="" => V0 identified
				// Considered current creds is for PRO type environment
				newfile := d.DefineDefaultCredFileName(credPath, deployObj.Name())
				if err := os.Rename(oldFile, newfile); err != nil {
					return fmt.Errorf("Unable to upgrade. %s", err)
				}
				gotrace.Info("Credential 'V0' file upgraded to '%s'.", version)
			}
		}
		return nil
	}); err != nil {
		gotrace.Info("Unable to upgrade your credentials. %s", err)
	}

	if v := a.cli.GetAction(cr_act).GetBoolAddr(no_maintain_f); v != nil {
		a.no_maintain = v
	}

	// Load drivers from repository Forjfile
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
	}

	if i, err := a.cli.GetAppStringValue(debug_instance_f); err == nil && i != "" {
		a.debug_instances = strings.Split(i, ",")
	}
	a.contextDisplayed()

	if err := a.DefineDefaultUpstream(); err != nil {
		if ok, err2 := a.f.DeployForjfile().Repos.AllHasAppWith("appRelName:upstream"); err != nil {
			a.w.SetError(err2)
			return nil, false
		} else if ok {
			gotrace.Warning("%s", err)
		} else {
			a.w.SetError(err)
			return nil, false
		}
	}

	return nil, true
}

// setInfraName configure the infra repository name
// It can take data from the cli, or Forjfile.
// If none are defined, forjj will create one based on the organization name
//
// When this is done, forjj will ensure this repo is available in list of repositories
// For next actions (REST API against plugins)
//
// Note that in maintains use case, cli is not usable. Data must be in the Forfile if default forjj calculates needs to be changed.
//
// forjj do not read any infra fields from Deployment Forjfiles. The only place where infra data must be stored is in the global one.
//
func (a *Forj) setInfraName(action string) (err error) {
	defer a.f.SetInfraAsRepo()
	org := a.w.GetString("organization")

	// Setting default if the organization is defined.
	if org != "" {
		// Set the 'infra' default flag value in cli
		a.cli.GetObject(infra).
			SetParamOptions(infra_name_f, cli.Opts().Default(fmt.Sprintf("%s-infra", org)))
	}

	var infraName string
	var found bool

	if action == maint_act {
		infraName, found, err = a.GetForgePrefs(infra_name_f)
	} else {
		infraName, found, err = a.GetPrefs(infra_name_f)
	}
	if err != nil {
		return err
	}

	infra := a.w.Infra()
	if found {
		// Set the infra repo name to use
		// Can be set only the first time
		if infra == nil {
			return fmt.Errorf("Internal issue: Infra object not found.")
		}
		if infra.Name == "" {
			// Get infra name from the flag
			infra.Name = infraName
			return a.SetPrefsTo("global", "forjj", infra_name_f, infra.Name) // Global Forjfile update
		}
		if infraName != infra.Name && org != "" {
			gotrace.Warning("You cannot update the Infra repository name from '%s' to '%s'.", infra.Name, infraName)
		}
		return a.SetPrefs("forjj", infra_name_f, infra.Name)
	}
	// Default infra-name
	if org != "" {
		// Use the default setting.
		infra.Name = fmt.Sprintf("%s-infra", org)
		err = a.SetPrefsTo("global", "forjj", infra_name_f, infra.Name) // Global Forjfile update
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
	curOrg := a.w.GetString("organization")
	if curOrg == "" {
		if found && orga != "" {
			a.w.Set("organization", orga, true)
		}
	} else {
		if found && curOrg != orga {
			gotrace.Warning("Sorry, but you cannot update the organization name. The Forjfile will be updated back to '%s'", curOrg)
		}
	}
	curOrg = a.w.GetString("organization")
	if curOrg != "" {
		if err := a.SetPrefs("forjj", orga_f, curOrg); err != nil {
			return err
		}
		log.Printf("Organization : '%s'", curOrg)
		a.cli.GetObject(workspace).SetParamOptions(orga_f, cli.Opts().Default(curOrg))
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
		if pwd, e := os.Getwd(); err != nil {
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

// setFromURLFlag initialize a URL structure from a flag given.
// If the flag is set from cli and valid, the URL will be stored in the given string address (store).
// if the flag has no value, store data is used as default.
// flag : Application flag value (from cli module)
//
// store : string address where this flag will be stored
//
func (a *Forj) setFromURLFlag(flag string, Set func(string, string) bool) (u *url.URL, e error) {
	value, found, err := a.GetLocalPrefs(flag)
	if err != nil {
		gotrace.Trace("%s", err)
		return nil, err
	}

	if found {
		if u, e = url.Parse(value); e == nil {
			Set(flag, u.String())
		}
		return
	}

	return nil, fmt.Errorf("Unable to define '%s'. Value not found in cli or workspace data.")
}
