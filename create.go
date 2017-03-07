package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/trace"
	"log"
	"forjj/drivers"
	"forjj/git"
)

// Create the Solution source code with validated parameters.
// The first time, an empty repo gets at least created with one README.md or a repotemplate code.
//
// Workspace data has been initialized or loaded.
// forjj-options has been initialized or loaded
func (a *Forj) Create() error {
	if !*a.no_maintain {
		log.Printf("CREATE: Automatic git push and forjj maintain enabled.")
	}

	if err := a.define_infra_upstream("create"); err != nil {
		return fmt.Errorf("Unable to identify a valid infra repository upstream. %s", err)
	}

	if a.w.Instance == "" {
		return fmt.Errorf("Unable to determine the upstream for your infra repository.\nIf you do not want any upstream, use --infra-upstream none.\nIf you want to connect to an upstream, at least use --apps upstream:<upstream_driver>[:<instance_name>]. If you have several upstreams, you will need to add the --infra-upstream <instance_name>.")
	}
	gotrace.Trace("Infra upstream selected: '%s'", a.w.Instance)

	// save infra repository location in the workspace.
	defer a.w.Save()

	if err, aborted, new_infra := a.ensure_infra_exists("create"); err != nil {
		if !aborted {
			return fmt.Errorf("Failed to ensure infra exists. %s", err)
		}
		log.Printf("Warning. %s", err)
	} else {
		if d, found := a.drivers[a.w.Instance]; new_infra && found {
			gotrace.Trace("New infra '%s' created. Need to connect it to the upstream.", a.w.Infra.Name)
			// New infra = new commits. Must maintain. Maintain will push because the upstream connection did not exist.

			// TODO: Repotemplates to help creating a the first commit (README.md at least)
			if e := a.ensure_local_repo_synced(a.w.Infra.Name, "master", "", "", a.infra_readme); e != nil {
				return fmt.Errorf("%s\n%s", err, e)
			}

			if d.HasNoFiles() {
				return fmt.Errorf("Plugin issue: No files to add/commit returned. Creating '%s' upstream requires to commit at least one file.", a.w.Instance)
			}
			// Commiting source code.
			if err := a.do_driver_commit(d); err != nil {
				return fmt.Errorf("Failed to commit '%s' source files. %s", a.w.Instance, err)
			}
			if err := a.do_driver_maintain(a.w.Instance); err != nil {
				// This will create/configure the upstream service
				// The commit will be pushed if the local repo upstream is inexistent. Which is the case of a new infra.
				return err
			}
			gotrace.Trace("The new infra is NOW connected to the upstream.")
		}
	}

	// Now, we are in the infra repo root directory and at least, the 1st commit exist and connected to an upstream.

	// TODO: flow_start to execute instructions before creating source code for new apps in appropriate branch. Possible if a flow is already implemented otherwise git must stay in master branch
	// flow_start()

	defer func() {
		// Save forjj-repos.yml
		if err := a.RepoCodeSave(); err != nil {
			log.Printf("%s", err)
		}

		if err := a.SaveForjjPluginsOptions(); err != nil {
			log.Printf("%s", err)
		}

		// Save forjj-options.yml
		a.o.SaveForjjOptions(fmt.Sprintf("Organization %s updated.", a.w.Organization))

		// Push if exist and automatic task is still enabled.
		if a.w.Infra.Exist && !*a.no_maintain {
			git.Do("push")
		} else {
			gotrace.Trace("No final push: infra is marked as inexistent.")
		}
	}()

	// Loop on drivers requested like jenkins classified as ci type.
	for instance, d := range a.drivers {
		if instance == a.w.Instance || !d.AppRequest() {
			continue // Do not try to create infra-upstream twice or create from a non requested app (--apps)
		}

		if err, aborted := a.do_driver_task("create", instance); err != nil {
			if !aborted {
				return fmt.Errorf("Failed to create '%s' source files. %s", instance, err)
			}
			log.Printf("Warning. %s", err)
			a.o.Drivers[instance] = d // Keep driver info in the forjj options
			continue
		}

		if d.HasNoFiles() {
			return fmt.Errorf("Plugin issue: No files to add/commit returned. Creating '%s' %s requires to commit at least one file.", a.w.Instance, d.DriverType)
		}

		// Commiting source code.
		if err := a.do_driver_commit(d); err != nil {
			return fmt.Errorf("Failed to commit '%s' source files. %s", instance, err)
		}

		a.o.Drivers[instance] = d // Keep driver info in the forjj options
	}

	// TODO: Implement the flow requested
	// flow_create() # Implement the flow on running tools for the infra-repo

	// TODO: Implement flow_close() to close the create task
	// flow_close()

	return nil
}

// Search for upstreams drivers and with or without --infra-upstream setting, the appropriate upstream will define the infra-repo upstream instance to use.
// It sets/Initialize
// - Forj.w.Instance        : Instance name
// - Forj.InfraPluginDriver : Driver details
// - Forj.w.Driver          : Driver name
func (a *Forj) define_infra_upstream(action string) (err error) {
	// Identify list of upstream instances
	gotrace.Trace("Identifying the infra Plugin driver...")

	// At the end, Set Forj.InfraPluginDriver/Forj.w.Driver/Forj.w.Driver.InfraRepo=true from instance in drivers
	defer func() {
		a.w.Driver = "none"
		if a.w.Instance == "none" {
			gotrace.Trace("No infra instance driver to load.")
			return
		}
		if d, found := a.drivers[a.w.Instance]; found {
			d.InfraRepo = true
			a.InfraPluginDriver = d
			a.w.Driver = d.Name
			gotrace.Trace("Infra Plugin driver identified and referenced.")
		} else {
			gotrace.Trace("Infra '%s' Plugin driver not found.", a.w.Instance, a.w.Instance)
		}
	}()

	// Set upstream instance requested to the infra-repo. ie Forj.w.Instance

	// Requested to change the infra instance name from the cli
	var instance_requested string
	if i, found, _, err := a.cli.GetStringValue(infra, "", infra_upstream_f) ; err != nil {
		return fmt.Errorf("Unable to define your INFRA upstream. %s", err)
	} else {
		if !found {
			return fmt.Errorf("INFRA upstream not defined.")
		}
		instance_requested = i
	}

	if instance_requested != "" && a.w.Instance != instance_requested {
		switch {
		case a.w.Instance == "none" || a.w.Instance == "":
			a.w.Instance = instance_requested
		default:
			log.Printf("Unable to migrate your infra-repository to a different instance. Not yet implemented. Migrating to '%s' is ignored.", instance_requested)
		}
	}

	// Inform user selected NO UPSTREAM and exit.
	if instance_requested == "none" && a.w.Instance == "none" {
		gotrace.Trace("No upstream instance configured as requested by --infra-upstream none")
		return
	}

	// Instance name is identified. Exiting.
	if a.w.Instance != "" {
		gotrace.Trace("Infra repository instance used: %s", a.w.Instance)
		return
	}

	// No upstream instance selected. Trying to get one from the list of drivers requested (--apps)
	upstreams := []*drivers.Driver{}
	for _, dv := range a.drivers {
		if dv.DriverType == "upstream" {
			upstreams = append(upstreams, dv)
		}
	}

	if len(upstreams) > 1 {
		err = fmt.Errorf("Multiple usptream drivers has been defined in your environment (cli/workspace). You must use --infra-upstream to select the appropriate upstream for your Infra repository or 'none' if you do not want any upstream for your Infra repository.")
		return
	}

	if len(upstreams) == 1 {
		a.w.Instance = upstreams[0].InstanceName
	}
	gotrace.Trace("Selected by default '%s' as upstream instance to connect '%s' repo", a.w.Instance, a.w.Infra.Name)
	return
}
