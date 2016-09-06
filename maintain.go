package main

import (
    "fmt"
    "log"
)

// Call docker to create the Solution source code from scratch with validated parameters.
// This container do the real stuff (git/call drivers)
// I would expect to have this go tool to have a do_create to replace the shell script.
// But this would be a next version and needs to be validated before this decision is made.
// Workspace information already loaded by the cli context.
func (a *Forj) Maintain() error {
    // Read forjj infra file and the options file given, defined by create/update driver flags settings saved or not
    a.LoadForjjPluginsOptions()

    // Read Repos list from infra-repo/forjj-repos.yaml
    if err := a.RepoCodeLoad() ; err != nil {
        return err
    }

    // Loop on instances to maintain them
    for instance, _ := range a.drivers {
        if err := a.do_driver_maintain(instance) ; err != nil {
            return fmt.Errorf("Unable to maintain requested resources of %s. %s", instance, err)
        }
    }

    // Loop on each drivers to ask them to maintain the infra, it controls.

    println("FORJJ - maintain ", a.w.Organization, " DONE") // , cmd.ProcessState.Sys().WaitStatus)
        return nil
}

func (a *Forj) do_driver_maintain(instance string) error {
    if instance == "none" {
        return nil
    }

    if err := a.driver_start(instance) ; err != nil {
        return err
    }
    d := a.CurrentPluginDriver

    // Ensure remote upstream exists - calling upstream driver - maintain
    // This will create/update the upstream service
    if err, _ := d.driver_do(a, instance, "maintain"); err != nil {
        return fmt.Errorf("Driver issue. %s.", err)
    }

    log.Printf("%d Repositories to manage.\n", len(a.drivers[instance].plugin.Result.Data.Repos))
    // Loop on upstream repositories to ensure it exists with at least a README.md file.
    for name, repo := range a.drivers[instance].plugin.Result.Data.Repos {
        if err := a.ensure_local_repo_initialized(name) ; err != nil {
            return err
        }

        // TODO: Generate README.md text from template.
        if err := a.ensure_local_repo_synced(name, "master", "origin", repo.Remotes["origin"], fmt.Sprintf("Repository %s created by Forjj.", name)) ; err != nil {
            return err
        }

        if a.InfraPluginDriver == d { // Infra upstream instance case
            if v, found := d.plugin.Result.Data.Repos[a.w.Infra.Name] ; found {
                // Saving infra repository information to the workspace
                a.w.Infra = v
            } else {
                if a.w.Infra.Name != "none" {
                    return fmt.Errorf("Unable to find '%s' from driver '%s'", a.w.Infra.Name, a.w.Instance)
                }
            }
        }

        // Save Managed repository to forjj options
        if d.DriverType == "upstream" {
            a.SaveManagedRepos(d, instance)
        }

    }

    // loop on other plugins to maintain themselves
    return nil
}
