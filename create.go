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
    // Initialize Repos with at least the infra-repo.
    // TODO: Add list of repos requested from the cli in this list.
    a.o.Init(a.w.Instance, a.w.Infra.Name)

    // Parent function has loaded context. ie Forj.w and Forj.o are loaded.

    if err := a.define_infra_upstream("create") ; err != nil {
        return fmt.Errorf("Unable to identify a valid infra repository upstream. %s", err)
    }

    gotrace.Trace("Infra upstream selected: '%s'", a.w.Instance)

    // save infra repository location in the workspace.
    defer a.w.Save(a)

    defer a.driver_cleanup(a.w.Instance) // Ensure upstream instances will be shutted down when done.

    if err, aborted := a.ensure_infra_exists("create") ; err != nil {
        if !aborted {
            return fmt.Errorf("Failed to ensure infra exists. %s", err)
        }
        log.Printf("Warning. %s", err)
    }

    if err := a.do_driver_maintain(a.w.Instance) ; err != nil { // This will create/configure the upstream service
        return err
    }

    // Now, we are in the infra repo root directory and at least, the 1st commit exist.

    // Loop on drivers requested like jenkins classified as ci type.
    for instance, d := range a.drivers {
        if instance == a.w.Instance || ! d.app_request {
            continue // Do not try to create infra-upstream twice or create from a non requested app (--apps)
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
    a.RepoCodeBuild()

    // TODO: If maintain is not sequenced, we could avoid pushing to upstream as well.
    if a.w.Infra.Remotes["origin"] != "" {
        if err := gitPush() ; err != nil {
            return err
        }
    }


    log.Print("FORJJ - create ", a.w.Organization, " DONE")
    return nil
}

// This function will ensure minimal git repo exists to store resources plugins data files.
// It will take care of several GIT scenarios. See ensure_local_repo_synced for details
// Used by create/update actions only.
// In case of create, a git push is executed.
func (a *Forj) ensure_infra_exists(action string) (err error, aborted bool) {

    if err := a.ensure_local_repo_initialized(a.w.Infra.Name) ; err != nil {
        return fmt.Errorf("Unable to ensure infra repository gets initialized. %s.", err), false
    }

    // Now, we are in the infra repo root directory. But at least is completely empty.

    // Set the Initial README.md content for the infra repository.
    a.infra_readme = fmt.Sprintf("Infrastructure Repository for the organization %s", a.Orga_name)

    if a.InfraPluginDriver == nil { // NO infra upstream driver loaded and defined.
        // But should be ok if the git remote is already set.

        var hint string
        if *a.Actions[action].flagsv["infra-upstream"] == "" {
            hint = "\nIf you are ok with this configuration, use '--infra-upstream none' to confirm. Otherwise, please define the --apps with the upstream driver and needed flags."
        }

        var remote_exist, remote_connected bool

        remote_exist, remote_connected, err = git_remote_exist("master", "origin", a.w.Infra.Remotes["origin"])
        if err != nil {
            return err, false
        }

        switch {
            case a.w.Instance == "" : // The infra repo upstream instance has not been defined.
                msg := fmt.Sprintf("Your workspace contains your infra repository called '%s' but not connected to", a.w.Infra)
                if ! remote_exist {
                    return fmt.Errorf("%s an upstream.%s", msg, hint), false
                }
                if ! remote_connected {
                    return fmt.Errorf("%s a valid upstream '%s'.%s", msg, a.w.Infra.Remotes["origin"], hint), false
                }
                return
            case a.w.Instance == "none" : // The infra is set with no upstream instance
                // Will create the 1st commit and nothing more.
                if err := a.ensure_local_repo_synced(a.w.Infra.Name, "", a.infra_readme) ; err != nil {
                    return err, false
                }
                return
            case a.w.Infra.Remotes["origin"] == "" : // The infra upstream string is not defined
                return fmt.Errorf("You provided the infra upstream instance name to connect to your local repository, without defining the upstream instance. please retry and use --apps to define it."), false
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
    a.o.Repos[a.w.Infra.Name] = a.w.Instance // Save the instance supporting the infra.

    if a.w.Infra.Remotes["origin"] != "" && action == "create" {
        if err := gitPush() ; err != nil {
            return err, false
        }
    }

    return
}

// Search for upstreams drivers and with or without --infra-upstream setting, the appropriate upstream will define the infra-repo upstream instance to use.
// It sets/Initialize
// - Forj.w.Instance     : Instance name
// - Forj.w.Plugin       : Driver name
// - Forj.o.Repos        : List Repositories and templates.
func (a *Forj) define_infra_upstream(action string) (err error) {
    // Identify list of upstream instances
    gotrace.Trace("Identifying the infra Plugin driver...")

    defer func() {
        if a.w.Instance == "none" {
            gotrace.Trace("No infra instance driver to load.")
            return
        }
        if d, found := a.drivers[a.w.Instance] ; found {
            d.InfraRepo = true
            a.InfraPluginDriver = d
            a.w.Driver = d.Name
            gotrace.Trace("Infra Plugin driver identified and referenced.")
        } else {
            log.Printf("Infra '%s' Plugin driver not found. Use --apps upstream:<DriverName>:%s to create/configure the infra repo upstream.", a.w.Instance, a.w.Instance)
        }
    }()

    // Set upstream instance requested to the infra-repo
    instance_requested := *a.Actions[action].flagsv["infra-upstream"]
    if instance_requested != "" && a.w.Instance != instance_requested { // Requested to change the infra instance name
        switch {
        case a.w.Instance == "none" || a.w.Instance == "":
            a.w.Instance = instance_requested
        default:
            log.Printf("Unable to migrate your infra-repository to a different instance. Not yet implemented.")
        }
    }

    if a.w.Instance != "" { // No need to define infra upstream as loaded from the workspace context.
        gotrace.Trace("Infra repository instance used: %s", a.w.Instance)
        return
    }
    infra := a.w.Infra.Name
    a.w.Instance = ""
    a.w.Driver = "none"
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
    v, found := a.InfraPluginDriver.plugin.Result.Data.Repos[a.w.Infra.Name]

    if  !found {
        return fmt.Errorf("Unable to rebuild your workspace from the upstream '%s'. Not found.", a.w.Infra)
    }

    if ! v.Exist {
        return fmt.Errorf("Unable to rebuild your workspace from the upstream '%s'. Inexistent.", a.w.Infra)
    }

    // Restoring the workspace.
    log.Printf("Updating your workspace from '%s(%s)'.", a.w.Infra.Name, a.w.Infra.Remotes["origin"])
    if err := a.ensure_local_repo_synced(a.w.Infra.Name, a.w.Infra.Remotes["origin"], a.infra_readme) ; err != nil {
        return fmt.Errorf("infra repository '%s' issue. %s", a.w.Infra)
    }
    log.Printf("As the upstream service already exists (not created as requested), forjj has only rebuilt or updated your workspace infra repository from '%s'.", a.w.Infra.Remotes["origin"])
    log.Print("HINT: Use create to create new application sources, or update to update existing application sources")
    a.LoadForjjOptions()
    return nil
}

// Execute the driver task, with commit and will execute the maintain step.
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
    // When the plugin inform that resource already exist, it returns an error with aborted = true
    if err, aborted = d.driver_do(a, instance, "create") ; err != nil && ! aborted  {
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
        for repo, _ := range a.o.Repos {
            if _, found := a.InfraPluginDriver.plugin.Result.Data.Repos[repo] ; found  {
                // Saving infra repository information to the workspace
                a.o.Repos[repo] = instance
            }
        }
    }

    if aborted {
        // Do not do any normal GIT tasks as everything already exists
        // Do not test the flag file as nothing done by the driver. If aborted, we assume the flag file already exists in the existing upstream repo
        return
    }

    // Check the flag file
    if _, err = os.Stat(flag_file) ; err != nil {
        log.Printf("Warning! Driver '%s' has not created the expected flag file (%s). Probably a driver bug. Contact the plugin maintainer to fix it.", d.Name, flag_file)
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


    gotrace.Trace("----- Do GIT tasks in the INFRA repository.")

    // Ensure initial commit exists and upstream are set for the infra repository
    if err := a.ensure_local_repo_synced(a.w.Infra.Name, a.w.Infra.Remotes["origin"], a.infra_readme) ; err != nil {
        return fmt.Errorf("infra repository '%s' issue. %s", err), false
    }

    // Add source files
    if err := d.gitAddPluginFiles() ; err != nil {
        return fmt.Errorf("Issue to add driver '%s' generated files. %s", a.CurrentPluginDriver.Name, err), false
    }

    // Check about uncontrolled files. Existing if one uncontrolled file is found
    if status := git_status() ; status.err != nil {
        return fmt.Errorf("Issue to check git status. %s", err), false
    } else {
        if len(status.Untracked) > 0 {
            log.Print("Following files created by the plugin are not controlled by the plugin. You must fix it manually and contact the plugin maintainer to fix this issue.")
            log.Printf("files: %s", strings.Join(status.Untracked, ", "))
            return fmt.Errorf("Unable to commit. Uncontrolled files found."), false
        }
    }

    // Commit files and drivers options
    if err := d.gitCommit() ; err != nil {
        return fmt.Errorf("git commit issue. %s", err), false
    }

    return
}
