package main

import (
    "fmt"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "log"
    "os"
    "path"
    "strings"
)


// Call docker to create the Solution source code from scratch with validated parameters.
// This container do the real stuff (git/call drivers)
// I would expect to have this go tool to have a do_create to replace the shell script.
// But this would be a next version and needs to be validated before this decision is made.
func (a *Forj) Create() error {
    if err := a.define_infra_upstream("create") ; err != nil {
        return fmt.Errorf("Unable to identify a valid infra repository upstream. %s", err)
    }

    gotrace.Trace("Infra upstream selected: '%s'", a.w.Instance)

    // save infra repository location in the workspace.
    defer a.w.Save(a)

    defer a.driver_cleanup(a.w.Instance) // Ensure upstream instances will be shutted down when done.

    if err, aborted := a.ensure_infra_exists() ; err != nil {
        if !aborted {
            return fmt.Errorf("Failed to ensure infra exists. %s", err)
        }
        log.Printf("Warning. %s", err)
    }

    if err := a.do_driver_maintain(a.w.Instance) ; err != nil { // This will create/configure the upstream service
        return err
    }

    // Now, we are in the infra repo root directory and at least, the 1st commit exist.

    a.o.update_options()

    // Loop on drivers requested like jenkins classified as ci type.
    for instance, _ := range a.drivers {
        if instance == a.w.Instance {
            continue // Do not try to create infra-upstream twice.
        }

        defer a.driver_cleanup(instance) // Ensure all instances will be shutted down when done.

        if err, aborted := a.do_driver_create(instance) ; err != nil {
            if !aborted {
                return fmt.Errorf("Failed to create '%s' source files. %s", instance, err)
            }
            log.Printf("Warning. %s", err)
        }

        // TODO: Except if --no-maintain is set, we could just create files and do maintain later.
        if err := a.do_driver_maintain(instance) ; err != nil { // This will create/configure the upstream service
            return err
        }
    }

    a.o.SaveForjjOptions(fmt.Sprintf("Organization %s updated.", a.w.Organization))

    // Start working on repositories
    a.RepoSave()

    log.Print("FORJJ - create ", a.w.Organization, " DONE")
    return nil
}

// This function will ensure minimal git repo exists to store resources plugins data files.
// It will take care of several GIT scenarios. See ensure_local_repo_synced for details
// Used by create action only.
func (a *Forj) ensure_infra_exists() (err error, aborted bool) {

    if err := a.ensure_local_repo_initialized(a.w.Infra) ; err != nil {
        return fmt.Errorf("Unable to ensure infra repository gets initialized. %s.", err), false
    }

    // Now, we are in the infra repo root directory. But at least is completely empty.

    // Set the Initial README.md content for the infra repository.
    a.infra_readme = fmt.Sprintf("Infrastructure Repository for the organization %s", a.Orga_name)

    if a.InfraPluginDriver == nil { // upstream UNdefined.
        // But should be ok if the git remote is already set.

        if *a.Actions["create"].flagsv["infra-upstream"] != "none" && a.w.Instance != "none"{
            var remote_exist, remote_connected bool
            remote_exist, remote_connected, err = git_remote_exist("master", "origin", a.w.Upstream)
            if err != nil {
                return err, false
            }

            msg := fmt.Sprintf("Your workspace contains your infra repository called '%s' but not connected to", a.w.Infra)
            hint := "\nIf you are ok with this configuration, use '--infra-upstream none' to confirm. Otherwise, please define the --apps with the upstream driver and needed flags."
            if ! remote_exist {
                return fmt.Errorf("%s an upstream. %s", msg, hint), false
            }
            if ! remote_connected {
                return fmt.Errorf("%s a valid upstream '%s'. %s", msg, a.w.Upstream, hint), false
            }
            return
        } else {
            a.w.Instance = "none"
        }

        // Will create the 1st commit and nothing more.
        if err := a.ensure_local_repo_synced(a.w.Infra, "", a.infra_readme) ; err != nil {
            return err, false
        }
        return
    }

    // Upstream defined

    err, aborted = a.do_driver_create(a.w.Instance)
    if aborted {
        // the upstream driver was not able to create the resources because already exist.
        // So the upstream resource may already exist and must be used to restore the local repo content from this resource.
        log.Printf("Plugin instance %s(%s) informed service already exists. Nothing created.", a.w.Instance, a.w.Driver)
        if e := a.restore_infra_repo() ; e != nil {
            err = fmt.Errorf("%s\n%s", err, e)
        }
    }

    return
}

// Search for upstreams drivers and with or without --infra-upstream setting, the appropriate upstream will define the infra-repo upstream instance to use.
// It sets
// - Forj.w.Instance     : Instance name
// - Forj.w.Plugin       : Driver name
func (a *Forj) define_infra_upstream(action string) (err error) {
    // Identify list of upstream instances
    gotrace.Trace("Identifying the infra Plugin driver...")

    defer func() {
        gotrace.Trace("Getting infra Plugin driver reference ...")
        if d, found := a.drivers[a.w.Instance] ; found {
            d.InfraRepo = true
            a.InfraPluginDriver = d
            a.w.Driver = d.Name
            gotrace.Trace("Infra Plugin driver identified and referenced.")
        } else {
            gotrace.Trace("Infra '%s' Plugin driver not found.", a.w.Instance)
        }
    }()

    if a.w.Instance != "" { // No need to define infra upstream as loaded from the workspace context.
        gotrace.Trace("Loaded from workspace: %s", a.w.Instance)
        return
    }
    infra := a.w.Infra
    a.w.Instance = ""
    a.w.Driver = "none"
    a.w.Upstream = ""
    upstreams := []*Driver{}
    upstream_requested := *a.Actions[action].flagsv["infra-upstream"]

    if upstream_requested == "none" {
        gotrace.Trace("No upstream instance configured as requested by --infra-upstream none")
        return
    }

    for _, dv := range a.drivers {
        if dv.DriverType == "upstream" {
            upstreams = append(upstreams, dv)
        }
        if dv.Name == upstream_requested {
            a.w.Instance = upstream_requested
            return
        }
    }

    if len(upstreams) >1 {
        err = fmt.Errorf("--infra-upstream missing with multiple upstreams defined. please select the appropriate upstream for your Infra repository or 'none'.")
        return
    }

    if len(upstreams) == 1 {
        a.w.Instance = upstreams[0].Name
    }
    gotrace.Trace("Selected by default '%s' as upstream instance to connect '%s' repo", a.w.Instance, infra)
    return
}

// Restore the workspace infra repo from the upstream.
func (a *Forj) restore_infra_repo() error {
    if a.InfraPluginDriver.plugin.Result == nil {
        return fmt.Errorf("Internal Error: The infra plugin did not return a valid result. Forj.InfraPluginDriver.plugin.Result = nil.")
    }
    v, found := a.InfraPluginDriver.plugin.Result.Data.Repos[a.w.Infra]

    if  !found {
        return fmt.Errorf("Unable to rebuild your workspace from the upstream '%s'. Not found.", a.w.Infra)
    }

    if ! v.Exist {
        return fmt.Errorf("Unable to rebuild your workspace from the upstream '%s'. Inexistent.", a.w.Infra)
    }

    // Restoring the workspace.
    a.w.Upstream = v.Remotes["origin"]
    log.Printf("Updating your workspace from '%s(%s)'.", a.w.Infra, a.w.Upstream)
    if err := a.ensure_local_repo_synced(a.w.Infra, a.w.Upstream, a.infra_readme) ; err != nil {
        return fmt.Errorf("infra repository '%s' issue. %s", a.w.Infra)
    }
    log.Printf("As the upstream service already exists (not created as requested), forjj has only rebuilt or updated your workspace infra repository from '%s'.", a.w.Upstream)
    log.Print("HINT: Use create to create new application sources, or update to update existing application sources")
    a.LoadForjjOptions()
    return nil
}

// Execute the driver task, with commit/push and will execute the maintain step.
func (a *Forj) do_driver_create(instance string) (err error, aborted bool) {
    if err = a.driver_start(instance) ; err != nil {
        return
    }

    d := a.CurrentPluginDriver

    // Add ref to this driver in the forjj infra repo
    a.o.Drivers[instance] = d

    flag_file := path.Join("apps", d.DriverType , d.FlagFile)

    if d.ForjjFlagFile {
        if _, err = os.Stat(flag_file) ; err == nil {
            err = fmt.Errorf("The driver instance '%s' has already created the resources. Use 'Update' to update it, and maintain to instanciate it as soon as your infra repo flow is completed.", instance)
            aborted = true
            return
        }
    }

    // Calling upstream driver - To create plugin source files for the current upstream infra repository
    if err, aborted = d.driver_do(a, instance, "create") ; err != nil {
        return
    }

    // Check the flag file
    if _, err = os.Stat(flag_file) ; err != nil {
        log.Printf("Warning! Driver '%s' has not created the flag file expected. Probably a driver bug. Contact the plugin maintainer to fix it.", d.Name)
        if err = touch(flag_file) ; err != nil {
            return
        }
        err = nil

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
    }

    if a.InfraPluginDriver != nil && a.w.Upstream == "none" {
        if v, found := a.InfraPluginDriver.plugin.Result.Data.Repos[a.w.Infra] ; found {
            a.w.Upstream = v.Upstream
        } else {
            return fmt.Errorf("Unable to find '%s' from driver '%s'", a.w.Infra, a.w.Instance), false
        }
    }

    // Ensure initial commit exists and upstream are set for the infra repository
    if err := a.ensure_local_repo_synced(a.w.Infra, a.w.Upstream, a.infra_readme) ; err != nil {
        return fmt.Errorf("infra repository '%s' issue. %s", err), false
    }

    // Add source files
    if err := a.CurrentPluginDriver.gitAddPluginFiles() ; err != nil {
        return fmt.Errorf("Issue to add driver '%s' generated files. %s", a.CurrentPluginDriver.Name, err), false
    }

    // Check about uncontrolled files. Existing if one uncontrolled file is found
    status := git_status()
    fmt.Printf("%#v\n", status)
    if  status.err != nil {
        return fmt.Errorf("Issue to check git status. %s", err), false
    } else {
        if len(status.Untracked) > 0 {
            log.Print("Following files created by the plugin are not controlled by the plugin. You must fix it manually and contact the plugin maintainer to fix this issue.")
            log.Printf("files: %s", strings.Join(status.Untracked, ", "))
            return fmt.Errorf("Unable to commit. Uncontrolled files found."), false
        }
    }

    // Commit files and drivers options
    if err := a.CurrentPluginDriver.gitCommit() ; err != nil {
        return fmt.Errorf("git commit issue. %s", err), false
    }

    if a.w.Upstream != "" {
        if err := gitPush() ; err != nil {
            return err, false
        }
    }

    return
}
