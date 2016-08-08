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
    // Ensure upstream driver is given
    if _, ok := a.drivers["upstream"] ; ! ok {
        kingpin.Fatalf("Missing upstream driver. Please use --git-us\n")
    }

    // Ensure local repo exists
    a.ensure_local_repo(a.w.Infra, "", fmt.Sprintf("Infrastructure Repository for the organization %s", a.Orga_name))
    // Now, we are in the infra repo root directory.

    // Create source for the infra repository - Calling upstream driver - create
    defer a.driver_cleanup("upstream")
    err := a.driver_do("upstream", "create")
    kingpin.FatalIfError(err, "Driver create issue. Unable to create plugin source code.")

    // Commit drivers files and drivers options
    err = a.DoCommitAll(fmt.Sprintf("Create %s DevOps infrastructure", a.w.Organization))
    kingpin.FatalIfError(err, "git commit issue")

    // Ensure remote upstream exists - calling upstream driver - maintain
    err = a.DoMaintainUpstream() // This will create/update the upstream service
    kingpin.FatalIfError(err, "Driver create issue. Unable to instantiate requested resources.")

    // Ensure local repo upstream properly configured
    //err = a.ensure_remote_repo(a.w.Infra)
    kingpin.FatalIfError(err, "Driver create issue. Unable to set git upstream.")

    // git add/commit and push
    //err = git("push")
    kingpin.FatalIfError(err, "Driver create issue. Unable to push.")

    println("FORJJ - create", a.w.Organization, "DONE") // , cmd.ProcessState.Sys().WaitStatus)
    // save infra repository location in the workspace.
    a.w.Save(a)
}
