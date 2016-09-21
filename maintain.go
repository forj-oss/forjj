package main

import (
    "fmt"
    "log"
    "os"
    "path"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
)

// Call docker to create the Solution source code from scratch with validated parameters.
// This container do the real stuff (git/call drivers)
// I would expect to have this go tool to have a do_create to replace the shell script.
// But this would be a next version and needs to be validated before this decision is made.
// Workspace information already loaded by the cli context.
func (a *Forj) Maintain() error {
    // Read forjj infra file and the options --file given, defined by create/update driver flags settings saved or not
    // This load Maintain context required by plugins. Maintain has limited flags to provide at runtime. Everything, except credentials should be stored in the infra-repo and workspace. Credentials is given with the --file option in yaml format.
    if err := a.LoadForjjPluginsOptions() ; err != nil {
        return fmt.Errorf("Unable to load plugins options. %s", err)
    }

    // Identify where is the infra-repo and move to it.
    infra_repo := path.Join(a.w.Path(), a.w.Infra.Name)
    if s, err := os.Stat(infra_repo) ; err != nil || !s.IsDir() {
        return fmt.Errorf("Invalid Infra repo. Inexistent or not a directory.")
    }
    if _, err := os.Stat(path.Join(infra_repo, ".git", "config")) ; err != nil {
        return fmt.Errorf("Invalid Infra repo. Seems not to be a git repository. Please check.")
    }

    gotrace.Trace("Moving to infra-repo '%s'", infra_repo)

    if err := os.Chdir(infra_repo) ; err != nil {
        return fmt.Errorf("Unable to move to the infra-repo at '%s'. %s", infra_repo, err)
    }

    // Load drivers from forjj-options.yml
    // loop from options/Repos and keep them in a.drivers

    return a.do_maintain()
}

func (a *Forj) do_maintain() error {
    // Read Repos list from infra-repo/forjj-repos.yaml
    if err := a.RepoCodeLoad() ; err != nil {
        return err
    }

    // Loop on instances to maintain them
    for instance, _ := range a.o.Drivers {
        if err := a.do_driver_maintain(instance) ; err != nil {
            return fmt.Errorf("Unable to maintain requested resources of %s. %s", instance, err)
        }
    }
    return nil
}

func (a *Forj) do_driver_maintain(instance string) error {
    if instance == "none" {
        return nil
    }

    gotrace.Trace("Start maintaining instance '%s'", instance)
    if err := a.driver_start(instance) ; err != nil {
        return err
    }
    d := a.CurrentPluginDriver

    // Ensure remote upstream exists - calling upstream driver - maintain
    // This will create/update the upstream service
    if err, _ := d.driver_do(a, instance, "maintain"); err != nil {
        return fmt.Errorf("Driver issue. %s.", err)
    }

    if d.DriverType != "upstream" {
        return nil
    }
    log.Printf("%s maintained by %s.\n", NumReposDisplay(len(a.drivers[instance].plugin.Result.Data.Repos)), instance)
    // Loop on upstream repositories to ensure it exists with at least a README.md file.

    // Ensure we are back to the infra repository.
    defer os.Chdir(a.RepoPath(a.w.Infra.Name))

    for name, repo := range a.drivers[instance].plugin.Result.Data.Repos {
        log.Printf("Maintaining local repo '%s'", name)
        if err := a.ensure_local_repo_initialized(name) ; err != nil {
            return err
        }

        // TODO: Generate README.md text from template.
        if err := a.ensure_local_repo_synced(name, "master", "origin", repo.Remotes["origin"], fmt.Sprintf("Repository %s created by Forjj.", name)) ; err != nil {
            return err
        }

        if a.InfraPluginDriver == d { // Infra upstream instance case
            if v, found := d.plugin.Result.Data.Repos[a.w.Infra.Name] ; found {
                // Saving infra repository information returned to the workspace
                a.w.Infra = v
            } else {
                return fmt.Errorf("Unable to find '%s' from driver '%s'", a.w.Infra.Name, a.w.Instance)
            }
        }
    }

    return nil
}
