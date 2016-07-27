package main

import "gopkg.in/alecthomas/kingpin.v2"

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
    a.ensure_local_repo(a.w.Infra)

    // Create source for the infra repository - Calling upstream driver - create
    _, err := a.driver_do("upstream", "create")
    kingpin.FatalIfError(err, "Driver create issue")

    // Ensure remote upstream exists - calling upstream driver - maintain
    //a.driver_do("upstream", "maintain") // This will create/update the upstream service

    // Ensure local repo upstream properly configured
    //a.ensure_remote_repo(a.w.Infra)

    // git add/commit and push
    // ??

    // To stop/remove all pending plugin services started.
    // plugins_close_all()

    println("FORJJ - create", a.w.Organization, "DONE") // , cmd.ProcessState.Sys().WaitStatus)
    // save infra repository location in the workspace.
    a.w.Save(a)
}
