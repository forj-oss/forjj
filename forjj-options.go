package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/trace"
	"forjj/git"
)

const (
	forjj_options_file = "Forjfile.yml"
)

// This data structure is going to be saved in the infra repository anytime a global update is done.
type ForjjOptions struct {
	Defaults map[string]string
	//Drivers  map[string]*drivers.Driver
}

func (a *Forj) GetUniqDriverName(driverType string) (od string) {
	var found_one_driver *string

	for instance, d := range a.drivers {
		if d.DriverType == driverType {
			switch {
			case found_one_driver == nil:
				found_one_driver = &od
				od = instance
			case *found_one_driver != "":
				od = ""
			}
		}
	}
	switch {
	case found_one_driver == nil:
		gotrace.Trace("No %s instance found.", driverType)
	case *found_one_driver == "":
		gotrace.Trace("Too many %s instances found.", driverType)
	default:
		gotrace.Trace("One upstream instance found: %s", *found_one_driver)
	}
	return
}

// Forjj-option: Set default
// - instance
// - flow
func (a *Forj) SetDefault(action string) {
	if od := a.GetUniqDriverName("upstream"); od != "" {
		a.o.Defaults["instance"] = od
	}

	// Set Defaults for repositories.
	if v, found, _, _ := a.cli.GetStringValue(infra, "", flow); found && v != "" {
		a.o.Defaults[flow] = v
	}
}

func (a *Forj) SaveForge(CommitMsg string) error {
	if err := a.f.Save(); err != nil {
		return fmt.Errorf("Unable to write '%s'. %s", a.f.Forjfile_name(), err)
	}

	git.Do("add", a.f.Forjfile_name())

	if err := git.Commit(CommitMsg, false); err != nil {
		return fmt.Errorf("Unable to commit the organization update. %s", err)
	}

	return nil
}

// LoadForge loads the forjj options definitions from the LoadContext().
func (a *Forj) LoadForge() (err error) {
	// Read definition file from repo.
	if err = a.f.SetInfraPath(a.w.InfraPath()) ; err != nil {
		return
	}

	_, err = a.f.Load()
	return
}
