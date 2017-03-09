package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/trace"
	"log"
	"os"
)

// Call docker to create the Solution source code from scratch with validated parameters.
// This container do the real stuff (git/call drivers)
// I would expect to have this go tool to have a do_create to replace the shell script.
// But this would be a next version and needs to be validated before this decision is made.
// Workspace information already loaded by the cli context.
func (a *Forj) Maintain() error {
	if _, err := a.w.Check_exist(); err != nil {
		return fmt.Errorf("Invalid workspace. %s. Please create it with 'forjj create'", err)
	}

	gotrace.Trace("Infra upstream selected: '%s'", a.w.Instance)

	if _, err := a.local_repo_exist(a.w.Infra.Name); err != nil {
		return fmt.Errorf("Invalid workspace. %s. Please create it with 'forjj create'", err)
	}

	// Now, we are in the infra repo root directory and at least, the 1st commit exist.

	// Load drivers from forjj-options.yml
	// loop from options/Repos and keep them in a.drivers

	return a.do_maintain()
}

func (a *Forj) do_maintain() error {
	// Read Repos list from infra-repo/forjj-repos.yaml
	if err := a.RepoCodeLoad(); err != nil {
		return err
	}

	// Loop on instances to maintain them
	for instance, _ := range a.drivers {
		if err := a.do_driver_maintain(instance); err != nil {
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
	if err := a.driver_start(instance); err != nil {
		return err
	}
	d := a.CurrentPluginDriver

	// Ensure remote upstream exists - calling upstream driver - maintain
	// This will create/update the upstream service
	if err, _ := a.driver_do(d, instance, "maintain"); err != nil {
		return fmt.Errorf("Driver issue. %s.", err)
	}

	if d.DriverType != "upstream" {
		return nil
	}
	log.Printf("%s maintained by %s.\n", NumReposDisplay(len(a.drivers[instance].Plugin.Result.Data.Repos)), instance)
	// Loop on upstream repositories to ensure it exists with at least a README.md file.

	// Ensure we are back to the infra repository.
	defer os.Chdir(a.RepoPath(a.w.Infra.Name))

	for name, repo := range a.drivers[instance].Plugin.Result.Data.Repos {
		log.Printf("Maintaining local repo '%s'", name)
		if err := a.ensure_local_repo_initialized(name); err != nil {
			return err
		}

		// TODO: Generate README.md text from template.
		if err := a.ensure_local_repo_synced(name, "master", "origin", repo.GetOrigin(),
			fmt.Sprintf("Repository %s created by Forjj.", name)); err != nil {
			return err
		}

		if a.InfraPluginDriver == d { // Infra upstream instance case
			if v, found := d.Plugin.Result.Data.Repos[a.w.Infra.Name]; found {
				// Saving infra repository information returned to the workspace
				a.w.Infra = &v
			} else {
				return fmt.Errorf("Unable to find '%s' from driver '%s'", a.w.Infra.Name, a.w.Instance)
			}
		}
	}

	return nil
}
