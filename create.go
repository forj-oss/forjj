package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/trace"
	"log"
	"forjj/drivers"
	"forjj/git"
	"strings"
	"path/filepath"
	"io/ioutil"
	"path"
)

//  initial_commit is called by infra.Create to create the initial commit with any needed files.
func (a *Forj) initial_commit() (files []string, err error) {
	files = []string{}

	if files, err = a.create_basic_README(files) ; err != nil {
		return files, err
	}

	if files, err = a.save_Forfile(files) ; err != nil {
		return files, err
	}

	if files, err = a.create_gitignore_files(files) ; err != nil {
		return files, err
	}

	return
}

// TODO: Ensure .forj-workspace is already set and set it is needed.


func (a *Forj) create_gitignore_files(files []string) (new_files []string, err error) {
	file_name := ".gitignore"
	gotrace.Trace("Generating %s", file_name)
	data := []byte(fmt.Sprintf("# Forjj workspace\n%s\n", Workspace_Name))
	err = a.create_source_text_file(file_name, data)
	if err != nil {
		return
	}
	new_files = append(files, file_name)
	return

}

func (a *Forj) save_Forfile(files []string) (new_files []string, err error) {
	if a.f.IsDirty() {
		err = a.f.Save()
	} else {
		return files, nil
	}
	if err != nil {
		return
	}
	new_files = append(files, a.f.Forjfile_name())
	return
}

func (a *Forj) create_basic_README(files []string) (new_files []string, err error) {
	file_name := "README.md"
	gotrace.Trace("Generating %s", file_name)
	data := []byte(fmt.Sprint("FYI: This Repository has been created by forjj\n"))
	err = a.create_source_text_file(file_name, data)
	if err != nil {
		return
	}
	new_files = append(files, file_name)
	return
}

func (a *Forj) create_source_text_file(file string, data []byte) error {
	var file_abs string
	if f, err := filepath.Abs(path.Join(a.i.Path(), file)) ; err != nil {
		return fmt.Errorf("Unable to create '%s'. %s\n", file_abs, err)
	} else {
		file_abs = f
	}
	if err := ioutil.WriteFile(file_abs, data, 0644); err != nil {
		return fmt.Errorf("Unable to create '%s'. %s\n", file_abs, err)
	}
	return nil
}

// Create the Solution source code with validated parameters.
// The first time, an empty repo gets at least created with one README.md or a repotemplate code.
//
// Workspace data has been initialized or loaded.
// forjj-options has been initialized or loaded
func (a *Forj) Create() error {
	a.ScanAndSetObjectData()
	if !*a.no_maintain {
		log.Printf("CREATE: Automatic git push and forjj maintain enabled.")
	}

	if err := a.define_infra_upstream("create"); err != nil {
		return fmt.Errorf("Unable to identify a valid infra repository upstream. %s", err)
	}

	gotrace.Trace("Infra upstream selected: '%s'", a.w.Instance)

	a.DefineDefaultUpstream()

	// TODO: Set/clone infra git remote when git-remote is set.

	// In create use case, a repository should not exist. If it exists one, we need an extra option to force using
	// it.

	// Then it commit initial files to the Infra repo.
	// TODO: Add force option. Currently, forced to false.
	if err := a.i.Create(a.f.InfraPath(), a.initial_commit, false) ; err != nil {
		return fmt.Errorf("Failed to create your infra repository. %s", err)
	}

	// As soon as the InfraPath gets created (or re-used) we can use the workspace in it.
	if err := a.w.RequireWorkspacePath() ; err != nil {
		return err
	}

	// Now, we are in the infra repo root directory and at least, the 1st commit exist

	// TODO: flow_start to execute instructions before creating source code for new apps in appropriate branch.
	// Possible if a flow is already implemented otherwise git must stay in master branch
	// flow_start()

	defer func() {
		// save infra repository location in the workspace.
		a.w.Save()

		if err := a.s.Save(); err != nil {
			log.Printf("%s", err)
		}
	}()

	instances := a.define_drivers_execution_order()

	// Loop on drivers requested like github or jenkins
	for _, instance := range instances {
		d := a.drivers[instance]
		if err, aborted := a.do_driver_task("create", instance); err != nil {
			if !aborted {
				return fmt.Errorf("Failed to create '%s' source files. %s", instance, err)
			}
			log.Printf("Warning. %s", err)
			continue
		}

		if d.HasNoFiles() {
			return fmt.Errorf("Plugin issue: No files to add/commit returned. Creating '%s' %s requires to commit at least one file.", a.w.Instance, d.DriverType)
		}

		// Committing source code.
		if err := a.do_driver_add(d); err != nil {
			return fmt.Errorf("Failed to Add '%s' source files. %s", instance, err)
		}
	}

	if err := git.Commit(fmt.Sprintf("Forge '%s' created.", a.w.Organization), true) ; err != nil {
		return fmt.Errorf("Failed to commit source files. %s", err)
	}

	// TODO: Implement the flow requested
	// flow_create() # Implement the flow on running tools for the infra-repo

	// TODO: Implement flow_close() to close the create task
	// flow_close()

	return nil
}

func (a *Forj) define_drivers_execution_order() (instances []string) {
	instances = make([]string, len(a.drivers))
	drivers := make(map[string]*drivers.Driver)
	index:=0
	for name, driver := range a.drivers { drivers[name] = driver }
	// first: execute upstream infra
	if instance := a.f.GetInfraInstance() ; instance != "" {
		instances[index] = instance
		index++
		delete(drivers, instance)
		gotrace.Trace("execution order will start with '%s'", instance)
	}

	// Get all others.
	for name := range drivers {
		instances[index] = name
		index++
	}
	gotrace.Trace("Execution order selected: '%s'", strings.Join(instances, "', '"))
	return
}

// Search for upstreams drivers and with or without --infra-upstream setting, the appropriate upstream will define the infra-repo upstream instance to use.
// It sets/Initialize
// - Forj.w.Instance        : Instance name
// - Forj.InfraPluginDriver : Driver details
// - Forj.w.Driver          : Driver name
//
// If something is wrong an error is returned. So, at least `a.w.Instance` and the Forjfile gets updated and non-empty.
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
			gotrace.Warning("Infra '%s' Plugin driver not found.", a.w.Instance)
		}
	}()

	// Set upstream instance requested to the infra-repo. ie Forj.w.Instance

	// Requested to change the infra instance name from the cli
	var instance_requested string

	instance_requested, _, err = a.GetPrefs(infra_upstream_f)
	if err != nil {
		return fmt.Errorf("Unable to define your INFRA upstream. %s", err)
	}

	if instance_requested != "" && a.w.Instance != instance_requested {
		switch {
		case a.w.Instance == "none" || a.w.Instance == "":
			a.w.Instance = instance_requested
		default:
			log.Printf("Unable to migrate your infra-repository to a different instance. " +
				"Not yet implemented. Migrating to '%s' is ignored.", instance_requested)
		}
	}

	// Inform user selected NO UPSTREAM and exit.
	if instance_requested == "none" && a.w.Instance == "none" {
		gotrace.Trace("No upstream instance configured as requested by '--infra-upstream none' " +
			"or Forjfile (infra/upstream-app:none)")
		err = a.SetPrefs(infra_upstream_f, a.w.Instance) // Forjfile updated
		return
	}

	// Instance name is identified. Exiting.
	if a.w.Instance != "" {
		gotrace.Trace("Infra repository instance used: %s", a.w.Instance)
		err = a.SetPrefs(infra_upstream_f, a.w.Instance) // Forjfile updated
		return
	}

	// No upstream instance selected. Trying to get one from the list of drivers requested (--apps)
	gotrace.Info("You did not specified any upstream application for your forge infra repository. " +
		"Forjj is trying to get it from the list of drivers you defined.")
	upstreams := []*drivers.Driver{}
	instances := []string{}
	for _, dv := range a.drivers {
		if dv.DriverType == "upstream" {
			upstreams = append(upstreams, dv)
			instances = append(instances, dv.InstanceName)
		}
	}

	if len(upstreams) > 1 {
		err = fmt.Errorf("Multiple usptream drivers has been defined in your environment (%s). " +
			"You must use --infra-upstream (or Forjfile:/infra/upstreams-app) to select the appropriate upstream " +
			"for your Infra repository or 'none' " +
			"if you do not want any upstream for your Infra repository.", strings.Join(instances, ", "))
		return
	}

	if len(upstreams) == 1 {
		a.w.Instance = upstreams[0].InstanceName
		gotrace.Trace("Selected by default '%s' as upstream instance to connect '%s' repo", a.w.Instance, a.w.Infra.Name)
		return a.SetPrefs(infra_upstream_f, a.w.Instance) // Forjfile updated
	}

	return fmt.Errorf("No 'upstream' application defined. At least one upstream application is required, " +
		"or set 'none' to --infra-upstream (or Forjfile:/infra/upstreams-app)")
}
