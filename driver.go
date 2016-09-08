package main

import (
    "path"
    "time"
    "log"
    "fmt"
    "os"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "strings"
    "github.hpe.com/christophe-larsonneur/goforjj"
)

const (
    defaultTimeout = 32 * time.Second
    default_socket_baseurl = "http:///anyhost"
    default_mount_path = "/src"
)

// Execute the driver task, with commit
func (a *Forj) do_driver_task(action, instance string) (err error, aborted bool) {
    if err = a.driver_start(instance) ; err != nil {
        return
    }

    d := a.CurrentPluginDriver

    // Add ref to this driver in the forjj infra repo
    a.o.Drivers[instance] = d

    // check flag for create
    if action == "create" {
        if err := d.check_flag_before(instance) ; err != nil {
            return err, true
        }
    }

    // Provide Repo data for upstream repos creation/update.



    // Calling upstream driver - To create plugin source files for the current upstream infra repository
    // When the plugin inform that resource already exist, it returns an error with aborted = true
    if err, aborted = d.driver_do(a, instance, action) ; err != nil && ! aborted  {
        return
    }

    // The driver has created or aborted his task.

    if a.InfraPluginDriver == d { // Infra upstream instance case
        if v, found := a.InfraPluginDriver.plugin.Result.Data.Repos[a.w.Infra.Name] ; found {
            // Saving infra repository information to the workspace
            a.w.Infra = v
        } else {
            if a.w.Infra.Name != "none" {
                return fmt.Errorf("Unable to find '%s' from driver '%s'", a.w.Infra.Name, a.w.Instance), false
            }
        }
    }

    // Save Managed repository to forjj options
    if d.DriverType == "upstream" {
        a.SaveManagedRepos(d, instance)
    }

    if aborted {
        // Do not do any normal GIT tasks as everything already exists
        // Do not test the flag file as nothing done by the driver. If aborted, we assume the flag file already exists in the existing upstream repo
        return
    }

    // Check the flag file
    if err = d.check_flag_after() ; err != nil {
        return
    }

    // Commiting source code.
    err = a.do_driver_commit(d)

    return
}

// Check if the flag exist to avoid creating the resource a second time. It must use update instead.
func (d *Driver)check_flag_before(instance string) error {
    flag_file := path.Join("apps", d.DriverType , d.FlagFile)

    if d.ForjjFlagFile {
        if _, err := os.Stat(flag_file) ; err == nil {
            return fmt.Errorf("The driver instance '%s' has already created the resources. Use 'Update' to update it, and maintain to instanciate it as soon as your infra repo flow is completed.", instance)
        }
    }
    return nil
}

func (d *Driver)check_flag_after() error {
    flag_file := path.Join("apps", d.DriverType , d.FlagFile)

    // Check the flag file
    if _, err := os.Stat(flag_file) ; err == nil {
        return err
    }

    log.Printf("Warning! Driver '%s' has not created the expected flag file (%s). Probably a driver bug. Contact the plugin maintainer to fix it.", d.Name, flag_file)

    // Create a forjj flag file instead.
    if err := touch(flag_file) ; err != nil {
        return err
    }

    var found bool
    for _, f := range d.plugin.Result.Data.Files {
        if f == d.FlagFile {
            found = true
        }
    }
    if ! found && ! d.ForjjFlagFile {
        if ! d.ForjjFlagFile {
            log.Printf("Warning! Driver '%s' has identified '%s' as controlled by itself. Probably a driver bug. Contact the plugin maintainer to fix it.", d.Name, d.FlagFile)
        }
        d.plugin.Result.Data.Files = append(d.plugin.Result.Data.Files, d.FlagFile)
    }
    return nil
}

// do driver commit
func (a *Forj) do_driver_commit(d *Driver) error {
    gotrace.Trace("----- Do GIT tasks in the INFRA repository.")

    // Ensure initial commit exists and upstream are set for the infra repository
    if a.w.Infra.Exist { // Upstream that we can ensure it to be connected.
        if err := a.ensure_local_repo_synced(a.w.Infra.Name, "master", "origin", a.w.Infra.Remotes["origin"], a.infra_readme) ; err != nil {
            return fmt.Errorf("infra repository '%s' issue. %s", a.w.Infra.Name, err)
        }
    } else { // No upstream to ensure connected.
        if err := a.ensure_local_repo_synced(a.w.Infra.Name, "master", "", "", a.infra_readme) ; err != nil {
            return fmt.Errorf("infra repository '%s' issue. %s", a.w.Infra.Name, err)
        }
    }

    // Add source files
    if err := d.gitAddPluginFiles() ; err != nil {
        return fmt.Errorf("Issue to add driver '%s' generated files. %s", a.CurrentPluginDriver.Name, err)
    }

    // Check about uncontrolled files. Existing if one uncontrolled file is found
    if status := git_status() ; status.err != nil {
        return fmt.Errorf("Issue to check git status. %s", status.err)
    } else {
        if len(status.Untracked) > 0 {
            log.Print("Following files created by the plugin are not controlled by the plugin. You must fix it manually and contact the plugin maintainer to fix this issue.")
            log.Printf("files: %s", strings.Join(status.Untracked, ", "))
            return fmt.Errorf("Unable to commit. Uncontrolled files found.")
        }
    }

    // Commit files and drivers options
    if err := d.gitCommit() ; err != nil {
        return fmt.Errorf("git commit issue. %s", err)
    }
    return nil
}

// Define starting on this driver
// Forj.CurrentPluginDriver set
func (a *Forj) driver_start(instance string) (error) {

    d, found := a.drivers[instance]
    if !found {
        return fmt.Errorf("Internal error: Unable to find %s from drivers.", instance)
    }
    a.CurrentPluginDriver = d
    return nil
}

// Start driver task.
// Forj.CurrentPluginDriver is set to the current driver
func (d *Driver) driver_do(a *Forj, instance_name, action string, args ...string) (err error, aborted bool) {
    log.Print("-------------------------------------------")
    log.Printf("Running %s on %s...", action, instance_name)

    if err := d.plugin.PluginInit(a.Workspace) ; err != nil {
        return err, false
    }

    d.plugin.PluginSetSource(path.Join(a.Workspace_path, a.Workspace, a.w.Infra.Name, "apps", d.DriverType))
    d.plugin.PluginSetWorkspace(path.Join(a.Workspace_path, a.Workspace))
    d.plugin.PluginSocketPath(path.Join(a.Workspace_path, a.Workspace, "lib"))
    if v, found := a.Actions[action].flagsv["docker-exe-path"] ; found {
        if err := d.plugin.PluginDockerBin(*v) ; err != nil {
            return err, false
        }
    }

    if err := d.plugin.PluginStartService(a.w.Organization + "_" + instance_name) ; err != nil {
        return err, false
    }

    plugin_payload := goforjj.PluginReqData {
        Args: make(map[string]string),
    }
    a.drivers_options.GetDriversMaintainParameters(plugin_payload.Args, action)
    a.GetDriversActionsParameters(plugin_payload.Args, "common")
    a.GetDriversActionsParameters(plugin_payload.Args, action)

    // For upstream plugins at create/update, provide ReposData map structure from forjj internals.
    if d.DriverType == "upstream" && action != "maintain" {
        // TODO: Get ReposData map structure from forjj internals...
        plugin_payload.ReposData = a.GetReposData(instance_name)
    }

    d.plugin.Result, err = d.plugin.PluginRunAction(action, plugin_payload)
    if d.plugin.Result == nil {
        return fmt.Errorf("An error occured in '%s' plugin. No data has been returned. Please check plugin logs.", instance_name), false
    }

    termBrown, termReset := DefColor(33)
    for _, line := range strings.Split(d.plugin.Result.Data.Status, "\n") {
        log.Printf("%s%s%s", termBrown, line, termReset)
    }

    if err != nil {
        if d.plugin.Result.State_code == 419 { // The plugin won't do the task because of requirement not met. This is not an error which requires Forjj to exit.
            aborted = true // So, when a plugin return 419, the plugin task is considered as aborted. So forjj can continue if it is possible. (create/update action case)
        }
        return err, aborted
    }

    // store plugins options required at maintain phase from what the plugin returned.
    if action != "maintain" {
        a.drivers_options.AddForjjPluginOptions(d.Name, d.plugin.Result.Data.Options, d.DriverType)
    }
    return
}

func (a *Forj) driver_cleanup(instance_name string) {
    if d, ok := a.drivers[instance_name] ; ok {
        d.plugin.PluginStopService()
    }
}
