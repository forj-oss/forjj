package main

import (
    "fmt"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "github.hpe.com/christophe-larsonneur/goforjj"
    "log"
)


// Create the Solution source code with validated parameters.
// The first time, an empty repo gets at least created with one README.md or a repotemplate code.
//
// Workspace data has been initialized or loaded.
// forjj-options has been initialized or loaded
func (a *Forj) Create() error {
    if ! *a.no_maintain {
        log.Printf("CREATE: Automatic git push and forjj maintain enabled.")
    }

    if err := a.define_infra_upstream("create") ; err != nil {
        return fmt.Errorf("Unable to identify a valid infra repository upstream. %s", err)
    }

    gotrace.Trace("Infra upstream selected: '%s'", a.w.Instance)

    // save infra repository location in the workspace.
    defer a.w.Save(a)

    defer a.driver_cleanup(a.w.Instance) // Ensure upstream instances will be shutted down when done.

    if err, aborted, new_infra := a.ensure_infra_exists("create") ; err != nil {
        if !aborted {
            return fmt.Errorf("Failed to ensure infra exists. %s", err)
        }
        log.Printf("Warning. %s", err)
    } else {
        if d, found := a.drivers[a.w.Instance] ; new_infra && found {
            gotrace.Trace("New infra '%s' created. Need to connect it to the upstream.", a.w.Infra.Name)
            // New infra = new commits. Must maintain. Maintain will push because the upstream connection did not exist.

            if d.HasNoFiles() {
                return fmt.Errorf("Plugin issue: No files to add/commit returned. Creating '%s' upstream requires to commit at least one file.", a.w.Instance)
            }
            // Commiting source code.
            if err := a.do_driver_commit(d) ; err != nil {
                return fmt.Errorf("Failed to commit '%s' source files. %s", a.w.Instance, err)
            }
            if err := a.do_driver_maintain(a.w.Instance) ; err != nil {
                // This will create/configure the upstream service
                // The commit will be pushed if the local repo upstream is inexistent. Which is the case of a new infra.
                return err
            }
            gotrace.Trace("The new infra is NOW connected to the upstream.")
        }
    }

    // Now, we are in the infra repo root directory and at least, the 1st commit exist and connected to an upstream.

    // TODO: flow_start to execute instructions before creating source code for new apps in appropriate branch. Possible if a flow is already implemented otherwise git must stay in master branch
    // flow_start()

    defer func() {
        // Save forjj-repos.yml
        if err := a.RepoCodeSave() ; err != nil {
            log.Printf("%s", err)
        }

        if err := a.SaveForjjPluginsOptions() ; err != nil {
            log.Printf("%s", err)
        }

        // Save forjj-options.yml
        a.o.SaveForjjOptions(fmt.Sprintf("Organization %s updated.", a.w.Organization))

        // Push if exist and automatic task is still enabled.
        if a.w.Infra.Exist && ! *a.no_maintain {
            git("push")
        } else {
            gotrace.Trace("No final push: infra is marked as inexistent.")
        }
    }()

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
            a.o.Drivers[instance] = d // Keep driver info in the forjj options
            continue
        }

        if d.HasNoFiles() {
            return fmt.Errorf("Plugin issue: No files to add/commit returned. Creating '%s' %s requires to commit at least one file.", a.w.Instance, d.DriverType)
        }

        // Commiting source code.
        if err := a.do_driver_commit(d) ; err != nil {
            return fmt.Errorf("Failed to commit '%s' source files. %s", instance, err)
        }

        a.o.Drivers[instance] = d // Keep driver info in the forjj options
    }


    // TODO: Implement the flow requested
    // flow_create() # Implement the flow on running tools for the infra-repo

    // TODO: Implement flow_close() to close the create task
    // flow_close()

    return nil
}

// This function will ensure minimal git repo exists to store resources plugins data files.
// It will take care of several GIT scenarios. See ensure_local_repo_synced for details
// Used by create/update actions only.
// In case of create a commit must be created but the push will be possible only when the upstream will be created through maintain step.
// If the repo already exist from the upstream, it will be simply restored.
func (a *Forj) ensure_infra_exists(action string) (err error, aborted, new_infra bool) {
    defer gotrace.Trace("Exiting ensure_infra_exists")

    if err = a.ensure_local_repo_initialized(a.w.Infra.Name) ; err != nil {
        err = fmt.Errorf("Unable to ensure infra repository gets initialized. %s.", err)
        return
    }

    // Now, we are in the infra repo root directory. But at least is completely empty.

    // Build Managed Forjj Repos list in memory.
    err = a.BuildReposList(action)
    if err != nil {
        return
    }

    // Set the Initial data & README.md content for the infra repository.
    if _, found := a.r.Repos[a.w.Infra.Name] ; ! found {
        gotrace.Trace("Defining your Infra-repository %s", a.w.Infra.Name)
        // TODO: Refer to a repotemplate to create the README.md content and file.
        r := goforjj.PluginRepoData {
            Title: fmt.Sprintf("Infrastructure Repository for the organization %s", *a.Orga_name),
            Instance: a.w.Instance,
            Templates: make([]string,0),
            Users: make(map[string]string),
            Groups: make(map[string]string),
            Options: make(map[string]string),
        }
        if v, found := a.o.Defaults["flow"] ; found {
            r.Flow = v
        }
        a.r.Repos[a.w.Infra.Name] = &r
    }

    a.infra_readme = fmt.Sprintf("Infrastructure Repository for the organization %s", *a.Orga_name)

    if a.InfraPluginDriver == nil { // NO infra upstream driver loaded and defined.
        // But should be ok if the git remote is already set.

        var remote_exist, remote_connected bool

        var hint string
        if *a.Actions[action].flagsv["infra-upstream"] == "" {
            hint = "\nIf you are ok with this configuration, use '--infra-upstream none' to confirm. Otherwise, please define the --apps with the upstream driver and needed flags."
        }

        remote_exist, remote_connected, err = git_remote_exist("master", "origin", a.w.Infra.Remotes["origin"])
        if err != nil {
            return
        }

        switch {
            case a.w.Instance == "" : // The infra repo upstream instance has not been defined.
                msg := fmt.Sprintf("Your workspace contains your infra repository called '%s' but not connected to", a.w.Infra.Name)
                switch {
                case !remote_exist :
                    err = fmt.Errorf("%s an upstream.%s", msg, hint)
                case !remote_connected :
                    err = fmt.Errorf("%s a valid upstream '%s'.%s", msg, a.w.Infra.Remotes["origin"], hint)
                }

            case a.w.Instance == "none" : // The infra is set with no upstream instance
                // Will create the 1st commit and nothing more.
                err = a.ensure_local_repo_synced(a.w.Infra.Name, "master", "", "", a.infra_readme)

            case a.w.Infra.Remotes["origin"] == "" : // The infra upstream string is not defined
                err = fmt.Errorf("You provided the infra upstream instance name to connect to your local repository, without defining the upstream instance. please retry and use --apps to define it.")

            case a.w.Infra.Remotes["origin"] != "" && ! remote_connected :
                err = a.ensure_local_repo_synced(a.w.Infra.Name, "master", "origin", a.w.Infra.Remotes["origin"], a.infra_readme)

            case a.w.Infra.Remotes["origin"] != "" && remote_connected :
                if action == "create" {
                    log.Printf("The infra already exist and is connected. The automatic git push/forjj maintain is then disabled.")
                    *a.no_maintain = true
                }

        }
        return
    }

    // -- Upstream driver defined --

    err, aborted = a.do_driver_task(action, a.w.Instance)

    // If an error occured, then we need to exit.
    if err != nil && ! aborted {
        return
    }

    // Save the instance supporting the infra.
    if v, found := a.r.Repos[a.w.Infra.Name]; found {
        v.Instance = a.w.Instance
    } else {
        a.r.Repos[a.w.Infra.Name] = &goforjj.PluginRepoData {
            Instance: a.w.Instance,
        }
    }

    // Ok Do we have an upstream on the server side?
    // No. So, nothing else now to do.
    // If driver has initial infra files (create case), we need to commit them, then maintain it, then push.
    // REMINDER: Create/Update works on source only.
    if ! a.w.Infra.Exist {
        new_infra = true
        return
    }

    if _, remote_connected, giterr := git_remote_exist("master", "origin", a.w.Infra.Remotes["origin"]) ; giterr != nil {
        if err != nil {
            err = fmt.Errorf("%s. %s.", err, giterr)
        }
    } else {
        if ! remote_connected {
            // the upstream driver has detected that resources already exists.
            // As the remote one seems different, we must be restore the local repo content from this resource.

            // The remote INFRA exist!!! We need to restore.

            if action == "create" {
                *a.no_maintain = true
                log.Printf("Plugin instance %s(%s) informed service already exists. Nothing created. And for this instance, you will need to use update instead of create. Automatic git push/maintain is disabled.", a.w.Instance, a.w.Driver)
            } else {
                log.Printf("Plugin instance %s(%s) informed service already exists. We need to restore the workspace before doing the update.", a.w.Instance, a.w.Driver)
            }
            if e := a.ensure_local_repo_synced(a.w.Infra.Name, "master", "origin", a.w.Infra.Remotes["origin"], a.infra_readme) ; e != nil {
                err = fmt.Errorf("%s\n%s", err, e)
            }

            log.Printf("As the upstream service already exists, forjj has only fetched your workspace infra repository from '%s'.",  a.w.Infra.Remotes["origin"])

            // Then re-apply cli default options and repos back to the existing restored code.
            a.LoadForjjOptions()

            // Build Managed Forjj Repos list in memory.
            err = a.BuildReposList(action)

            // Now in case of update task, we can re-applied the fix on the workspace restored. In case of create, the user will need to use update instead.
            if action == "update" {
                err, aborted = a.do_driver_task(action, a.w.Instance)
            }
        }
    }
    return
}

// Search for upstreams drivers and with or without --infra-upstream setting, the appropriate upstream will define the infra-repo upstream instance to use.
// It sets/Initialize
// - Forj.w.Instance        : Instance name
// - Forj.InfraPluginDriver : Driver details
// - Forj.w.Driver          : Driver name
func (a *Forj) define_infra_upstream(action string) (err error) {
    // Identify list of upstream instances
    gotrace.Trace("Identifying the infra Plugin driver...")

    // At the end, Set Forj.InfraPluginDriver/Forj.w.Driver/Forj.w.Driver.InfraRepo=true from instance in drivers
    defer func() {
        a.w.Driver = "none"
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

    // Set upstream instance requested to the infra-repo. ie Forj.w.Instance

    // Requested to change the infra instance name from the cli
    instance_requested := *a.Actions[action].flagsv["infra-upstream"]
    if instance_requested != "" && a.w.Instance != instance_requested {
        switch {
        case a.w.Instance == "none" || a.w.Instance == "":
            a.w.Instance = instance_requested
        default:
            log.Printf("Unable to migrate your infra-repository to a different instance. Not yet implemented. Migrating to '%s' is ignored.", instance_requested)
        }
    }

    // Inform user selected NO UPSTREAM and exit.
    if instance_requested == "none" &&  a.w.Instance == "none" {
        gotrace.Trace("No upstream instance configured as requested by --infra-upstream none")
        return
    }

    // Instance name is identified. Exiting.
    if a.w.Instance != "" {
        gotrace.Trace("Infra repository instance used: %s", a.w.Instance)
        return
    }

    // No upstream instance selected. Trying to get one from the list of drivers requested (--apps)
    upstreams := []*Driver{}
    for _, dv := range a.drivers {
        if dv.DriverType == "upstream" {
            upstreams = append(upstreams, dv)
        }
    }

    if len(upstreams) >1 {
        err = fmt.Errorf("Multiple usptream drivers has been defined in your environment (cli/workspace). You must use --infra-upstream to select the appropriate upstream for your Infra repository or 'none' if you do not want any upstream for your Infra repository.")
        return
    }

    if len(upstreams) == 1 {
        a.w.Instance = upstreams[0].Name
    }
    gotrace.Trace("Selected by default '%s' as upstream instance to connect '%s' repo", a.w.Instance, a.w.Infra.Name)
    return
}

