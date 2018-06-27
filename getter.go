package main

import (
	"fmt"

	"github.com/forj-oss/forjj-modules/trace"
)

// Used to map between flags (cli) and Forjfile

type AppMapEntry struct {
	cliFuncs map[string]func(name string) (string, bool, bool, error)

	cli_obj      string
	cli_instance string
	cli_field    string

	forj_section  string
	forj_instance string
	forj_field    string
}

func (a *Forj) AddMap(key, cli_obj, cli_instance, cli_field, forj_section, forj_instance, forj_field string) {
	if a.appMapEntries == nil {
		a.appMapEntries = make(map[string]AppMapEntry)
	}
	a.appMapEntries[key] = AppMapEntry{
		cliFuncs:      make(map[string]func(name string) (string, bool, bool, error)),
		cli_obj:       cli_obj,
		cli_instance:  cli_instance,
		cli_field:     cli_field,
		forj_section:  forj_section,
		forj_instance: forj_instance,
		forj_field:    forj_field,
	}
}

func (a *Forj) AddMapFunc(cliCmd, key string, getter func(name string) (string, bool, bool, error)) {
	if v, found := a.appMapEntries[key]; found {
		v.cliFuncs[cliCmd] = getter
		a.appMapEntries[key] = v
	}
}

// GetPrefs return a value found (or not) from different source of data
// 1: from cli, then exit if found
// 2: from Forjfile if found
// 3: Then cli default
func (a *Forj) GetPrefs(field string) (string, bool, error) {
	var entry AppMapEntry

	if e, found := a.appMapEntries[field]; !found {
		return "", false, fmt.Errorf("Unable to get '%s' from Forjfile/cli mapping. Missing", field)
	} else {
		entry = e
	}

	v, found, isdefault, err := a.cli.GetStringValue(entry.cli_obj, entry.cli_instance, entry.cli_field)
	if err != nil {
		gotrace.Trace("Unable to get data from cli. %s", err)
		err = nil
	}
	if found && !isdefault {
		gotrace.Trace("Found Forjfile setting '%s' from cli : %s", entry.cli_field, v)
		return v, found, err
	}
	if v2, found2 := a.f.GetString(entry.forj_section, entry.forj_instance, entry.forj_field); found2 {
		gotrace.Trace("Found Forjfile setting '%s' from Forjfile : %s", entry.forj_field, v2)
		return v2, found2, nil
	}
	if found {
		gotrace.Trace("Found Forjfile setting '%s' from cli default: %s", entry.cli_field, v)
	} else {
		gotrace.Trace("Local setting '%s' not found from any of cli, Forjfile or cli default", field)
	}
	return v, found, err
}

// GetLocalPrefs return a value found (or not) from different source of data
// 1: from cli, then exit if found
// 2: from Workspace local settings if found
//    Usually, LocalSetting is loaded from a Forjfile template and stored in the Workspace.
// 3: Then cli default
func (a *Forj) GetLocalPrefs(field string) (string, bool, error) {
	var entry AppMapEntry

	if e, found := a.appMapEntries[field]; !found {
		return "", false, fmt.Errorf("Unable to get '%s' from Forjfile/cli mapping. Missing", field)
	} else {
		entry = e
	}

	var (
		v         string
		found     bool
		isdefault bool
		err       error
	)

	if f, funcFound := entry.cliFuncs[a.contextAction]; funcFound {
		v, found, isdefault, err = f(field)
	} else {
		v, found, isdefault, err = a.cli.GetStringValue(entry.cli_obj, entry.cli_instance, entry.cli_field)
	}

	if err != nil {
		gotrace.Trace("Unable to get data from cli. %s", err)
		err = nil
	}
	if found && !isdefault {
		gotrace.Trace("Found Local setting '%s' from cli: %s", entry.forj_field, v)
		return v, found, err
	}
	if v2, found2 := a.w.GetString(entry.forj_field); found2 {
		gotrace.Trace("Found Local setting '%s' from workspace : %s", entry.forj_field, v2)
		return v2, found2, nil
	}
	if found {
		gotrace.Trace("Found Local setting '%s' from cli default: %s", entry.cli_field, v)
	} else {
		gotrace.Trace("Local setting '%s' not found from any of cli, Forjfile template or cli default", entry.cli_field)
	}
	return v, found, err
}

// GetForgePrefs return a value found (or not) from Forjfile
func (a *Forj) GetForgePrefs(field string) (v string, found bool, _ error) {
	var entry AppMapEntry

	if e, found := a.appMapEntries[field]; !found {
		return "", false, fmt.Errorf("Unable to get '%s' from Forjfile/cli mapping. Missing", field)
	} else {
		entry = e
	}

	v, found = a.f.GetString(entry.forj_section, entry.forj_instance, entry.forj_field)
	if found {
		gotrace.Trace("Found Forjfile setting '%s' from Forjfile : %s", entry.forj_field, v)
	} else {
		gotrace.Trace("Forfile setting '%s' not found from Forjfile.", entry.forj_field)
	}
	return
}

func (a *Forj) SetPrefs(field, value string) error {
	var entry AppMapEntry

	if e, found := a.appMapEntries[field]; !found {
		return fmt.Errorf("Unable to get '%s' from Forjfile/cli mapping. Missing", field)
	} else {
		entry = e
	}

	a.f.Set(entry.forj_section, entry.forj_instance, entry.forj_field, value)
	return nil
}
