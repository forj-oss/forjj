package main

import (
    "gopkg.in/alecthomas/kingpin.v2"
    "fmt"
)

// Call docker to create the Solution source code from scratch with validated parameters.
// This container do the real stuff (git/call drivers)
// I would expect to have this go tool to have a do_create to replace the shell script.
// But this would be a next version and needs to be validated before this decision is made.
func (a *Forj) Maintain() {
    // Load Workspace information
    a.w.Load(a)

    // Read forjj infra file and the options file given
    a.LoadForjjPluginsOptions()

    // Do the maintain first on upstream.
    err := a.DoMaintainUpstream()
    kingpin.FatalIfError(err, "Unable to maintain requested resources.")

    // Loop on each drivers to ask them to maintain the infra, it controls.

    println("FORJJ - maintain", a.w.Organization, "DONE") // , cmd.ProcessState.Sys().WaitStatus)
}

func (a *Forj) DoMaintainUpstream() error {
    // Ensure remote upstream exists - calling upstream driver - maintain
    // This will create/update the upstream service
    if err := a.driver_do("upstream", "maintain"); err != nil {
        return fmt.Errorf("Driver issue. %s.", err)
    }

    fmt.Printf("%d Repositories to manage.\n", len(a.drivers["upstream"].plugin.Result.Data.Repos))
    // Loop on upstream repositories to ensure it exists with at least a README.md file.
    for name, repo := range a.drivers["upstream"].plugin.Result.Data.Repos {
        if err := a.ensure_local_repo(name, repo.Upstream, "") ; err != nil {
            return err
        }
    }

    // loop on other plugins to maintain themselves
    return nil
}
