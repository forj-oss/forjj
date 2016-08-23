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
    err := a.w.Load(a)
    kingpin.FatalIfError(err, "Unable to maintain requested resources.")

    // Read forjj infra file and the options file given
    a.LoadForjjPluginsOptions()

    // Loop on instances to maintain them
    for instance, _ := range a.drivers {
        err = a.do_driver_maintain(instance)
        kingpin.FatalIfError(err, "Unable to maintain requested resources of %s.", instance)
    }

    // Loop on each drivers to ask them to maintain the infra, it controls.

    println("FORJJ - maintain", a.w.Organization, "DONE") // , cmd.ProcessState.Sys().WaitStatus)
}

func (a *Forj) do_driver_maintain(instance string) error {
    // Ensure remote upstream exists - calling upstream driver - maintain
    // This will create/update the upstream service
    if err := a.driver_do(instance, "maintain"); err != nil {
        return fmt.Errorf("Driver issue. %s.", err)
    }

    fmt.Printf("%d Repositories to manage.\n", len(a.drivers[instance].plugin.Result.Data.Repos))
    // Loop on upstream repositories to ensure it exists with at least a README.md file.
    for name, repo := range a.drivers[instance].plugin.Result.Data.Repos {
        if err := a.ensure_local_repo_initialized(name) ; err != nil {
            return err
        }

        // TODO: Generate README.md text from template.
        if err := a.ensure_local_repo_synced(name, repo.Upstream, fmt.Sprintf("Repository %s created by Forjj.", name)) ; err != nil {
            return err
        }
    }

    // loop on other plugins to maintain themselves
    return nil
}
