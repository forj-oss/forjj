package main

import (
	"fmt"
	"forjj/creds"
	"forjj/drivers"
	"forjj/forjfile"
	"forjj/git"
	"forjj/utils"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

const (
	defaultTimeout         = 32 * time.Second
	default_socket_baseurl = "http:///anyhost"
	default_mount_path     = "/src"
)

// Execute the driver task, without commit.
// Used in create/update case.
// It uses the driver (or forjj) flag to ensure proper task is executed by the driver.
// if exist in create, fails and ask to abort
// If ! exist in update, fails
func (a *Forj) do_driver_task(action, instance string) (err error, aborted bool) {
	if action != "create" && action != "update" {
		return fmt.Errorf("Internal error: Invalid action '%s'. Supports only 'create' and 'update'.", action), false
	}

	if err = a.driver_init(instance); err != nil {
		return
	}

	d := a.CurrentPluginDriver

	// Add ref to this driver in the forjj infra repo
	//a.o.Drivers[instance] = d

	// check flag for create
	if err := d.CheckFlagBefore(instance, action); err != nil {
		return err, (action == "create") // Abort-able if create, because the resource exist and we can use it. So, forjj can continue the task.
	}

	// Calling upstream driver - To create plugin source files for the current upstream infra repository
	// When the plugin inform that resource already exist, it returns an error with aborted = true
	if err, aborted = a.driver_do(d, instance, action); err != nil && !aborted {
		return
	}

	// The driver has created or aborted his task.

	if a.InfraPluginDriver == d { // Infra upstream instance case
		if v, found := a.InfraPluginDriver.Plugin.Result.Data.Repos[a.w.Infra.Name]; found {
			// Saving infra repository information to the workspace
			a.w.Infra = &v
		} else {
			return fmt.Errorf("Unable to find Infra repository '%s' from driver '%s'", a.w.Infra.Name, a.w.Instance), false
		}
	}

	if aborted {
		// Do not do any normal GIT tasks as everything already exists
		// Do not test the flag file as nothing done by the driver. If aborted, we assume the flag file already exists in the existing upstream repo
		return
	}

	// Check the flag file
	if err = d.CheckFlagAfter(); err != nil {
		return
	}

	return
}

func (a *Forj) moveTo(where string) (cur_dir string, _ error) {
	if v, err := os.Getwd(); err != nil {
		return "", fmt.Errorf("Unable to get the current directory. %s", err)
	} else {
		cur_dir = v
	}
	if where == goforjj.FilesSource {
		err := os.Chdir(a.f.InfraPath())
		if err != nil {
			return "", fmt.Errorf("Unable to move to '%s'. %s", a.f.InfraPath(), err)
		}
		gotrace.Trace("Moved to %s repo (%s)", where, a.f.InfraPath())
	}
	if where == goforjj.FilesDeploy {
		err := os.Chdir(a.d.GetRepoPath())
		if err != nil {
			return "", fmt.Errorf("Unable to move to '%s'. %s", a.d.GetRepoPath(), err)
		}
		gotrace.Trace("Moved to %s repo (%s)", where, a.d.GetRepoPath())
	}
	return
}

// do driver add files
func (a *Forj) do_driver_add(d *drivers.Driver) error {
	if len(d.Plugin.Result.Data.Files) == 0 {
		gotrace.Trace("No files to add/commit returned by the driver.")
		return nil
	}
	gotrace.Trace("----- Do GIT tasks in the INFRA repository.")

	// Add source files
	if err := d.GitAddPluginFiles(a.moveTo); err != nil {
		return fmt.Errorf("Issue to add driver '%s' generated files. %s", a.CurrentPluginDriver.Name, err)
	}

	// Check about uncontrolled files. Existing if one uncontrolled file is found
	if status := git.GetStatus(); status.Err != nil {
		return fmt.Errorf("Issue to check git status. %s", status.Err)
	} else {
		if num := status.CountUntracked(); num > 0 {
			log.Print("Following files created by the plugin are not controlled by the plugin. You must fix it manually and contact the plugin maintainer to fix this issue.")
			log.Printf("files: %s", strings.Join(status.Untracked(), ", "))
			return fmt.Errorf("Unable to complete commit process. '%d' Uncontrolled files found", num)
		}
	}
	return nil
}

// Define starting on this driver
// Forj.CurrentPluginDriver set
func (a *Forj) driver_init(instance string) error {

	d, found := a.drivers[instance]
	if !found {
		return fmt.Errorf("Internal error: Unable to find %s from drivers.", instance)
	}
	a.CurrentPluginDriver = d
	return nil
}

func (a *Forj) driver_cleanup_all() {
	gotrace.Trace("Stopping all running loaded services...")
	for instance, d := range a.drivers {
		gotrace.Trace("- %s", instance)
		d.Plugin.PluginStopService()
	}
}

// Create the flag to a kingpin Command. (create/update/maintain)
func (a *Forj) init_driver_flags_for(d *drivers.Driver, option_name, command, forjj_option_name, forjj_option_help string,
	opts *cli.ForjOpts) {
	if command == "" {
		// Add to the Application layer.
		gotrace.Trace("Set App flag '%s(%s)'", forjj_option_name, option_name)
		a.cli.AddAppFlag(cli.String, forjj_option_name, forjj_option_help, opts)
		return
	}
	// No value by default. Will be set later after complete parse.
	d.InitCmdFlag(command, forjj_option_name, option_name)

	// Create flag 'option_name' on kingpin cmd or app
	if forjj_option_name != option_name {
		gotrace.Trace("Set action '%s' flag '%s(%s)'", command, forjj_option_name, option_name)
	} else {
		gotrace.Trace("Set action '%s' flag '%s'", command, forjj_option_name)
	}
	a.cli.OnActions(command).AddFlag(cli.String, forjj_option_name, forjj_option_help, opts)
	return
}

// Start driver task.
// Forj.CurrentPluginDriver is set to the current driver
func (a *Forj) driver_do(d *drivers.Driver, instance_name, action string, args ...string) (err error, aborted bool) {
	defer log.Print("-------------------------------------------")
	log.Print("-------------------------------------------")
	log.Printf("Running %s on %s...", action, instance_name)

	if err := d.Plugin.PluginInit(a.w.Organization + "_" + instance_name); err != nil {
		return err, false
	}

	if found, _ := goforjj.InArray(instance_name, a.debug_instances); found {
		d.Plugin.RunningFromDebugger()
	}

	d.Plugin.PluginSetSource(path.Join(a.i.Path(), "apps", d.DriverType))
	d.Plugin.PluginSetDeployment(a.d.GetReposPath())
	d.Plugin.PluginSetDeploymentName(a.d.Name())
	d.Plugin.PluginSetVersion(d.DriverVersion)
	d.Plugin.PluginSetWorkspace(a.w.Path())
	d.Plugin.PluginSocketPath(path.Join(a.w.Path(), "lib"))
	if v, found, _, _ := a.cli.GetStringValue(workspace, "", "docker-exe-path"); found && v != "" {
		a.w.DockerBinPath = v
	}
	if err := d.Plugin.PluginDockerBin(a.w.DockerBinPath); err != nil {
		return err, false
	}

	// Set default envs from the forjj process environment.
	if d.Plugin.Yaml.Runtime.Docker.Env == nil {
		d.Plugin.Yaml.Runtime.Docker.Env = make(map[string]string)
	}

	d.Plugin.ServiceAddEnv("LOGNAME", "$LOGNAME", false)
	d.Plugin.Yaml.Runtime.Docker.Env["LOGNAME"] = "$LOGNAME"
	if v := os.Getenv("http_proxy"); v != "" {
		d.Plugin.Yaml.Runtime.Docker.Env["http_proxy"] = v
		d.Plugin.Yaml.Runtime.Docker.Env["https_proxy"] = v
	}
	if v := os.Getenv("no_proxy"); v != "" {
		d.Plugin.Yaml.Runtime.Docker.Env["no_proxy"] = v
	}
	if v, b, _ := d.Plugin.GetDockerDoodParameters(); v != nil {
		d.Plugin.Yaml.Runtime.Docker.Env["DOCKER_DOOD"] = strings.Join(v, " ")
		d.Plugin.Yaml.Runtime.Docker.Env["DOCKER_DOOD_BECOME"] = strings.Join(b, " ")
	}

	if err := d.Plugin.PluginStartService(); err != nil {
		return err, false
	}

	plugin_payload := goforjj.NewReqData()

	// Load all internal Forjj data, identified by 'forjj-*'
	a.LoadInternalData()
	a.GetForjjFlags(plugin_payload, d, common_acts)
	a.GetForjjFlags(plugin_payload, d, action)
	if err := a.GetObjectsData(plugin_payload, d, action); err != nil {
		return fmt.Errorf("Unable to Get Object data on '%s'. %s", instance_name, err), aborted
	}
	err = a.AddReqDeployment(plugin_payload)
	if err != nil {
		return fmt.Errorf("Unable to %s. %s. You may need to execute a forjj update to a deployment environment", action, err), false
	}

	d.Plugin.Result, err = d.Plugin.PluginRunAction(action, plugin_payload)
	if err != nil {
		return fmt.Errorf("Internal Error: %s", err), false
	}
	if d.Plugin.Result == nil {
		return fmt.Errorf("An error occured in '%s' plugin. No data has been returned. Please check plugin logs.", instance_name), false
	}

	termBrown, termReset := utils.DefColor(33)
	for _, line := range strings.Split(d.Plugin.Result.Data.Status, "\n") {
		log.Printf("%s%s%s", termBrown, line, termReset)
	}

	if d.Plugin.Result.Data.ErrorMessage != "" {
		termRed, _ := utils.DefColor(31)
		for _, line := range strings.Split(d.Plugin.Result.Data.ErrorMessage, "\n") {
			log.Printf("%s%s%s", termRed, line, termReset)
		}
	}
	if err != nil {
		if d.Plugin.Result.State_code == 419 { // The plugin won't do the task because of requirement not met. This is not an error which requires Forjj to exit.
			aborted = true // So, when a plugin return 419, the plugin task is considered as aborted. So forjj can continue if it is possible. (create/update action case)
		}
		return err, aborted
	}

	// Dispatch driver information in Forjj

	// Deliver list of Remotes in Internal Forjfile
	if d.DriverType == "upstream" {
		for Name, Repo := range d.Plugin.Result.Data.Repos {
			var repo_obj *forjfile.RepoStruct
			ffd := a.f.InMemForjfile()
			if r, ok := ffd.GetObjectInstance(repo, Name).(*forjfile.RepoStruct); !ok {
				continue
			} else {
				repo_obj = r
			}
			repo_obj.Set("forjj", "remote", Repo.Remotes["origin"].Ssh)
			repo_obj.Set("forjj", "remote-url", Repo.Remotes["origin"].Url)
			repo_obj.SetInstanceOwner(Repo.Owner)
			repo_obj.SetPluginOwner(d)
			if a.f.GetInfraName() == Name { // 'infra' *Repostruct is a copy, not a ref to the repos["<infra>"] *RepoStruct
				ffd.Set("forjj", "infra", Name, "remote", Repo.Remotes["origin"].Ssh)
				ffd.Set("forjj", "infra", Name, "remote-url", Repo.Remotes["origin"].Url)
			}

			// New repo code which could be created with repo-templates
			if !Repo.Exist && !repo_obj.IsCurrentDeploy() && !repo_obj.IsInfra() {
				codeRepoPath := path.Join(a.w.Path(), "git", Name)
				git.EnsureRepoExist(codeRepoPath)
				git.RunInPath(codeRepoPath, func() (_ error) {
					git.EnsureRemoteIs("origin", Repo.Remotes["origin"].Ssh)
					syncRemoteBranch := "origin" + "/" + "master"
					git.Do("branch", "--set-upstream-to="+syncRemoteBranch)
					return
				})
			}

			// Current deploy only
			if deployName, found, _ := repo_obj.Get(forjfile.FieldRepoDeployName); found && repo_obj.IsCurrentDeploy() {
				deployObj, _ := a.f.GetADeployment(deployName.GetString())

				deployObj.RunInContext(func() (_ error) {
					deployObj.GitDefineRemote("origin", Repo.Remotes["origin"].Ssh)
					deployObj.GitSyncFrom("origin", "master")
					return
				})
			}
		}
		if err := a.FlowApply(); err != nil {
			return err, false
		}
		if err := a.scanAndSetDefaults(a.f.InMemForjfile(), creds.Global); err != nil {
			return err, false
		}
	}
	// Collect application API if published by the driver.
	if u, found := d.Plugin.Result.Data.Services.Urls["api_url"]; found {
		d.DriverAPIUrl = u
	}

	return
}

func (a *Forj) DriverGet(instance string) (d *drivers.Driver) {
	var found bool

	if d, found = a.drivers[instance]; found {
		return
	}

	return nil
}
