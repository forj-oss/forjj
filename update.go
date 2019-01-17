package main

import (
	"fmt"
	"forjj/creds"
	"log"
	"regexp"

	"github.com/forj-oss/forjj-modules/trace"
)

func (a *Forj) updateAction(string) {
	if err := a.Update(); err != nil {
		log.Fatalf("Forjj update issue. %s", err)
	}
	println("FORJJ - update ", a.w.GetString("organization"), " DONE") // , cmd.ProcessState.Sys().WaitStatus)

}

// Update execute an update on the workspace given.
//
// Workspace data has been initialized or loaded.
// forjj-options has been initialized or loaded
func (a *Forj) Update() error {
	if _, err := a.w.EnsureExist(); err != nil {
		return fmt.Errorf("Invalid workspace. %s. Please create it with 'forjj create'", err)
	}

	defer func() {
		// save infra repository location in the workspace.
		a.w.Save()
	}()

	if err := a.ValidateForjfile(); err != nil {
		return fmt.Errorf("Your Forjfile is having issues. %s Try to fix and retry", err)
	}

	// Set plugin defaults for objects defined by plugins loaded.
	if err := a.scanAndSetDefaults(a.f.DeployForjfile(), creds.Global); err != nil {
		return fmt.Errorf("Unable to update. Global dispatch issue. %s", err)
	}

	// Build in memory representation from source files loaded.
	if err := a.f.BuildForjfileInMem(); err != nil {
		return err
	}

	// Get it
	ffd := a.f.InMemForjfile()

	// Add missing deployment Repositories and warn
	if err := a.DefineDeployRepositories(ffd, true); err != nil {
		return fmt.Errorf("Issues to automatically add your deployment repositories. %s", err)
	}

	// Defining information about current deployment repository
	a.defineDeployContext()

	// Load flow identified by Forjfile source with missing repos.
	if err := a.FlowInit(); err != nil {
		return err
	}

	if err := a.define_infra_upstream(); err != nil {
		return fmt.Errorf("Unable to identify a valid infra repository upstream. %s", err)
	}

	gotrace.Trace("Infra upstream selected: '%s'", a.w.GetString("infra-instance-name"))

	// Apply the flow to the inMemForjfile
	if err := a.FlowApply(); err != nil {
		return fmt.Errorf("Unable to apply flows. %s", err)
	}

	// Set plugin defaults for objects added dynamically in the in memory Forjfile.
	if err := a.scanAndSetDefaults(ffd, creds.Global); err != nil {
		return fmt.Errorf("Unable to update. Global dispatch issue. %s", err)
	}

	// Checking infra repository: A valid infra repo is a git repository with at least one commit and
	// a Forjfile from repo root.
	if err := a.i.Use(a.f.InfraPath()); err != nil {
		return fmt.Errorf("Failed to update your infra repository. %s", err)
	}

	// Now, we are in the infra repo root directory and at least, the 1st commit exist.

	// Disabled as not ready.
	//if err := a.MoveToFixBranch(*a.Actions["update"].argsv["branch"]) ; err != nil {
	//    return fmt.Errorf("Unable to move to your feature branch. %s", err)
	//}

	instances := a.define_drivers_execution_order()

	// Loop on drivers requested like github or jenkins
	for _, instance := range instances {
		d := a.drivers[instance]
		if err, aborted := a.do_driver_task("update", instance); err != nil {
			d.HasNoFiles()

			// Clean identified source code.
			if err := a.doDriverClean(d); err != nil {
				return fmt.Errorf("Failed to Add '%s' source files. %s", instance, err)
			}
			if !aborted {
				return fmt.Errorf("Failed to update '%s' source files. %s", instance, err)
			}
			log.Printf("Warning. %s", err)
			continue
		}

		if d.HasNoFiles() {
			gotrace.Info("No files to add/commit.")
			continue
		}

		// Adding source code to GIT.
		if err := a.do_driver_add(d); err != nil {
			return fmt.Errorf("Failed to Add '%s' source files. %s", instance, err)
		}
	}

	commitMsg := fmt.Sprintf("Forge '%s' updated.", a.w.GetString("organization"))

	if deployPublish, found, _ := a.cli.GetBoolValue("_app", "forjj", "deploy-publish"); found && deployPublish {
		if err := a.d.GitCommit(commitMsg); err != nil {
			return fmt.Errorf("Failed to commit deploy files. %s", err)
		}

		if err := a.d.GitPush(false); err != nil {
			return fmt.Errorf("Failed to push deploy commits. %s", err)
		}
	}

	return nil
}

func (a *Forj) MoveToFixBranch(branch string) error {
	a.Branch = branch

	if ok, _ := regexp.MatchString(`^[\w_-]+$`, branch); !ok {
		return fmt.Errorf("Invalid git branch name '%s'. alphanumeric, '_' and '-' are supported.", branch)
	}
	return nil
}
