package main

import (
    "fmt"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "log"
)


// Call docker to create the Solution source code from scratch with validated parameters.
// This container do the real stuff (git/call drivers)
// I would expect to have this go tool to have a do_create to replace the shell script.
// But this would be a next version and needs to be validated before this decision is made.
func (a *Forj) Create() error {
    // Initialize forjj-options structure
    a.o.Init()

    // Read Repos list from infra-repo/forjj-repos.yaml
    if err := a.RepoCodeLoad() ; err != nil {
        return err
    }
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
    } else {
        if err := a.do_driver_maintain(a.w.Instance) ; err != nil { // This will create/configure the upstream service
            return err
        }
    }

    // Now, we are in the infra repo root directory and at least, the 1st commit exist and connected to an upstream.

    // TODO: flow_start to execute instructions before creating source code for new apps in appropriate branch. Possible if a flow is already implemented otherwise git must stay in master branch
    // flow_start()

    // Loop on drivers requested like jenkins classified as ci type.
    for instance, d := range a.drivers {
        if instance == a.w.Instance || ! d.app_request {
            continue // Do not try to create infra-upstream twice or create from a non requested app (--apps)
        }

        defer a.driver_cleanup(instance) // Ensure all instances will be shutted down when done.

        if err, aborted := a.do_driver_task("create", instance) ; err != nil {
            if !aborted {
                return fmt.Errorf("Failed to create '%s' source files. %s", instance, err)
            }
            log.Printf("Warning. %s", err)
            continue
        }

        // TODO: Except if --no-maintain is set, we could just create files and do maintain later.
        // Do push before maintaining it. So the code updated gets published before we instantiate it.
        if a.w.Infra.Remotes["origin"] != "" {
            if err := gitPush() ; err != nil {
                return err
            }
        }
        // TODO: Move maintain to a second create phase. But before flow_create.
        if err := a.do_driver_maintain(instance) ; err != nil { // This will create/configure the upstream service
            return err
        }
    }

    // TODO: Implement the flow requested
    // flow_create() # Implement the flow on running tools for the infra-repo

    a.o.SaveForjjOptions(fmt.Sprintf("Organization %s updated.", a.w.Organization))

    // Start working on repositories, writing repos source code.
    a.RepoCodeBuild("create")

    // TODO: If maintain is not sequenced, we could avoid pushing to upstream as well.
    if a.w.Infra.Remotes["origin"] != "" {
        if err := gitPush() ; err != nil {
            return err
        }
    }
    // TODO: Implement flow_close() to close the create task
    // flow_close()

    // TODO: Implement all repositories (maintain)
    // a.RepoMaintain() # This will implement the flow for the infra-repo as well.

    log.Print("FORJJ - create ", a.w.Organization, " DONE")
    return nil
}

// This function will ensure minimal git repo exists to store resources plugins data files.
// It will take care of several GIT scenarios. See ensure_local_repo_synced for details
// Used by create/update actions only.
// In case of create a commit must be created but the push will be possible only when the upstream will be created through maintain step.
func (a *Forj) ensure_infra_exists(action string) (err error, aborted bool) {

    if err := a.ensure_local_repo_initialized(a.w.Infra.Name) ; err != nil {
        return fmt.Errorf("Unable to ensure infra repository gets initialized. %s.", err), false
    }

    // Now, we are in the infra repo root directory. But at least is completely empty.

    // Set the Initial README.md content for the infra repository.
    a.infra_readme = fmt.Sprintf("Infrastructure Repository for the organization %s", *a.Orga_name)

    var remote_exist, remote_connected bool

    if a.InfraPluginDriver == nil { // NO infra upstream driver loaded and defined.
        // But should be ok if the git remote is already set.

        var hint string
        if *a.Actions[action].flagsv["infra-upstream"] == "" {
            hint = "\nIf you are ok with this configuration, use '--infra-upstream none' to confirm. Otherwise, please define the --apps with the upstream driver and needed flags."
        }

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
                if err := a.ensure_local_repo_synced(a.w.Infra.Name, "master", "", "", a.infra_readme) ; err != nil {
                    return err, false
                }
                return
            case a.w.Infra.Remotes["origin"] == "" : // The infra upstream string is not defined
                return fmt.Errorf("You provided the infra upstream instance name to connect to your local repository, without defining the upstream instance. please retry and use --apps to define it."), false
        }
        return
    }

    // Upstream defined

    err, aborted = a.do_driver_task(action, a.w.Instance)

    if err != nil && ! aborted { // An error occured, so, we need to exit.
        return
    }

    // Save the instance supporting the infra.
    if v, found := a.r.Repos[a.w.Infra.Name]; found {
        v.Instance = a.w.Instance
    } else {
        a.r.Repos[a.w.Infra.Name] = &RepoStruct {
            Instance: a.w.Instance,
        }
    }

    // Ok Do we have an upstream on the server side?
    if ! a.w.Infra.Exist { // No. So, nothing else now to do. We need maintain to do the needed work to create the expected resources.
        return nil, false
    }

    if action == "create" {
        log.Printf("Plugin instance %s(%s) informed service already exists. Nothing created.", a.w.Instance, a.w.Driver)
    }

    remote_exist, remote_connected, err = git_remote_exist("master", "origin", a.w.Infra.Remotes["origin"])
    if err != nil {
        return err, false
    }

    if ! remote_connected {
        // the upstream driver has detected that resources already exists.
        // As the remote one seems different, we  must be restore the local repo content from this resource.

        if e := a.restore_infra_repo() ; e != nil {
            err = fmt.Errorf("%s\n%s", err, e)
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
    if ! a.w.Infra.Exist {
        return fmt.Errorf("Unable to rebuild your workspace from the upstream '%s'. Inexistent.", a.w.Infra)
    }

    // Restoring the workspace.
    log.Printf("Updating your workspace from '%s(%s)'.", a.w.Infra.Name, a.w.Infra.Remotes["origin"])
    if err := a.ensure_local_repo_synced(a.w.Infra.Name, "master", "origin", a.w.Infra.Remotes["origin"], a.infra_readme) ; err != nil {
        return fmt.Errorf("infra repository '%s' issue. %s", a.w.Infra)
    }
    log.Printf("As the upstream service already exists (not created as requested), forjj has only rebuilt or updated your workspace infra repository from '%s'.", a.w.Infra.Remotes["origin"])
    log.Print("HINT: Use create to create new application sources, or update to update existing application sources")
    a.LoadForjjOptions()
    return nil
}
