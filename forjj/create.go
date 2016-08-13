package main

import (
    "gopkg.in/alecthomas/kingpin.v2"
    "fmt"
)


// Call docker to create the Solution source code from scratch with validated parameters.
// This container do the real stuff (git/call drivers)
// I would expect to have this go tool to have a do_create to replace the shell script.
// But this would be a next version and needs to be validated before this decision is made.
func (a *Forj) Create() {
    // Ensure local repo exists
    err := a.ensure_local_repo(a.w.Infra, "", fmt.Sprintf("Infrastructure Repository for the organization %s", a.Orga_name))
    kingpin.FatalIfError(err, "Driver create issue. Unable to get a valid infra repository.")
    // Now, we are in the infra repo root directory.

    infra_upstream, err := a.define_upstream(a.w.Infra, "create")
    kingpin.FatalIfError(err, "Driver create issue. Unable to identify a valid infra repository upstream.")
    // Do connecting the upstream if set.

    if _, ok := a.drivers[infra_upstream] ; ! ok {
        a.do_driver_create(infra_upstream)
    }

    // Loop on drivers requested like jenkins classified as ci type.
    for instance, _ := range a.drivers {
        if instance == infra_upstream {
            continue // Do not try to create infra-upstream twice.
        }
        a.do_driver_create(instance)
    }

    println("FORJJ - create", a.w.Organization, "DONE") // , cmd.ProcessState.Sys().WaitStatus)
    // save infra repository location in the workspace.
    a.w.Save(a)
}

// Search for upstreams drivers and with or without --infra-upstream setting, the appropriate upstream will define the infra-repo upstream instance to use.
func (a *Forj) define_upstream(infra, action string) (instance string, err error) {
    // Identify list of upstream instances

    upstreams := []Driver{}
    upstream_requested := *a.Actions[action].flagsv["infra-upstream"]

    if upstream_requested == "none" {
        return "none", nil
    }

    for _, dv := range a.drivers {
        if dv.driver_type == "upstream" {
            upstreams = append(upstreams, dv)
        }
        if dv.name == upstream_requested {
            return upstream_requested, nil
        }
    }

    if len(upstreams) >1 {
        return "", fmt.Errorf("--infra-upstream missing with multiple upstreams defined. please select the appropriate upstream for your Infra repository or 'none'.")
    }

    instance = upstreams[0].name
    return
}

// Execute the driver task, with commit.
// if maintain is kept at the end of create phase, it ensures local repos exist in the workspace
func (a *Forj) do_driver_create(instance string){

    // Create source for the infra repository - Calling upstream driver - create
    defer a.driver_cleanup(instance)
    err := a.driver_do(instance, "create")
    kingpin.FatalIfError(err, "Driver create issue. Unable to create plugin source code.")

    // Commit drivers files and drivers options
    err = a.DoCommitAll(fmt.Sprintf("Create %s DevOps infrastructure", a.w.Organization))
    kingpin.FatalIfError(err, "git commit issue")

    // Ensure remote upstream exists - calling upstream driver - maintain
    err = a.do_driver_maintain(instance) // This will create/update the upstream service
    kingpin.FatalIfError(err, "Driver create issue. Unable to instantiate requested resources.")

    // Ensure local repo upstream properly configured
    //err = a.ensure_remote_repo(a.w.Infra)
    kingpin.FatalIfError(err, "Driver create issue. Unable to set git upstream.")

    // git add/commit and push
    //err = git("push")
    kingpin.FatalIfError(err, "Driver create issue. Unable to push.")

}
