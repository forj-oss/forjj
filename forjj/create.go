package main

import (
    "gopkg.in/alecthomas/kingpin.v2"
    "fmt"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
)


// Call docker to create the Solution source code from scratch with validated parameters.
// This container do the real stuff (git/call drivers)
// I would expect to have this go tool to have a do_create to replace the shell script.
// But this would be a next version and needs to be validated before this decision is made.
func (a *Forj) Create() {
    err := a.define_infra_upstream("create")
    kingpin.FatalIfError(err, "Driver create issue. Unable to identify a valid infra repository upstream.")

    gotrace.Trace("Infra upstream selected: '%s'", a.w.Instance)

    err = a.ensure_infra_exists()
    kingpin.FatalIfError(err, "Driver create issue. failed to ensure infra exists.")

    // Now, we are in the infra repo root directory and at least, the 1st commit exist.

    // Loop on drivers requested like jenkins classified as ci type.
    for instance, _ := range a.drivers {
        defer a.driver_cleanup(instance) // Ensure all instances will be shutted down when done.

        if instance == a.w.Instance {
            continue // Do not try to create infra-upstream twice.
        }
        err = a.do_driver_create(instance)
        kingpin.FatalIfError(err, "Driver create issue. failed to create '%s' source files.", instance)
    }

    println("FORJJ - create", a.w.Organization, "DONE") // , cmd.ProcessState.Sys().WaitStatus)
    // save infra repository location in the workspace.
    a.w.Save(a)
}

// This function will ensure minimal git repo exists to store resources plugins data files.
// It will take care of several GIT scenarios. See ensure_local_repo_synced for details
func (a *Forj) ensure_infra_exists() error {

    if err := a.ensure_local_repo_initialized(a.w.Infra) ; err != nil {
        return fmt.Errorf("Unable to ensure infra repository gets initialized. %s.", err)
    }

    // Now, we are in the infra repo root directory. But at least is completely empty.

    // Set the Initial README.md content for the infra repository.
    a.infra_readme = fmt.Sprintf("Infrastructure Repository for the organization %s", a.Orga_name)

    if i_d := a.InfraPluginDriver ; i_d == nil { // upstream UNdefined.
        // Will create the 1st commit and nothing more.
        if err := a.ensure_local_repo_synced(a.w.Infra, "", a.infra_readme) ; err != nil {
            return err
        }
        return nil
    }

    // Upstream defined

    if err := a.do_driver_create(a.w.Instance) ; err != nil {
        return err
    }

    return nil
}

// Search for upstreams drivers and with or without --infra-upstream setting, the appropriate upstream will define the infra-repo upstream instance to use.
// It sets
// - Forj.w.Instance     : Instance name
func (a *Forj) define_infra_upstream(action string) (err error) {
    // Identify list of upstream instances

    infra := a.w.Infra
    a.w.Instance = "none"
    a.infra_upstream = "none"
    upstreams := []Driver{}
    upstream_requested := *a.Actions[action].flagsv["infra-upstream"]

    if upstream_requested == "none" {
        gotrace.Trace("No upstream instance configured as requested by --infra-upstream none")
        return
    }

    defer func() {
        if d, found := a.drivers[a.w.Instance] ; found {
            a.InfraPluginDriver = &d
        } else {
            if a.w.Instance != "none" {
                err = fmt.Errorf("Unable to find driver instance '%s' in loaded drivers list.", a.w.Instance)
            }
        }
    }()

    for _, dv := range a.drivers {
        if dv.driver_type == "upstream" {
            upstreams = append(upstreams, dv)
        }
        if dv.name == upstream_requested {
            a.w.Instance = upstream_requested
            return
        }
    }

    if len(upstreams) >1 {
        err = fmt.Errorf("--infra-upstream missing with multiple upstreams defined. please select the appropriate upstream for your Infra repository or 'none'.")
        return
    }

    a.w.Instance = upstreams[0].name
    gotrace.Trace("Selected by default '%s' as upstream instance to connect '%s' repo", a.w.Instance, infra)
    return
}

// Execute the driver task, with commit/push and will execute the maintain step.
func (a *Forj) do_driver_create(instance string) error {
    // Calling upstream driver - To create plugin source files for the current upstream infra repository
    if err := a.driver_do(a.w.Instance, "create") ; err != nil {
        return err
    }

    if a.InfraPluginDriver != nil && a.infra_upstream == "none" {
        if v, found := a.InfraPluginDriver.plugin.Result.Data.Repos[a.w.Infra] ; found {
            a.infra_upstream = v.Upstream
        } else {
            return fmt.Errorf("Unable to find '%s' from driver '%s'", a.w.Infra, a.w.Instance)
        }
    }

    // Ensure initial commit exists and upstream are set for the infra repository
    if err := a.ensure_local_repo_synced(a.w.Infra, a.infra_upstream, a.infra_readme) ; err != nil {
        return fmt.Errorf("infra repository '%s' issue. %s", err)
    }

    // Add source files
    if err := a.CurrentPluginDriver.gitAddPluginFiles() ; err != nil {
        return fmt.Errorf("Issue to add driver '%s' generated files. %s", a.CurrentPluginDriver.name, err)
    }

    // Commit files and drivers options
    if err := a.CurrentPluginDriver.gitCommit() ; err != nil {
        return fmt.Errorf("git commit issue. %s", err)
    }

    if a.infra_upstream != "none" {
        if err := gitPush() ; err != nil {
            return err
        }
    }

    // TODO: Except if --no-maintain is set, we could just create files and do maintain later.
    if err := a.do_driver_maintain(instance) ; err != nil { // This will create/configure the upstream service
        return err
    }
    return nil
}
