package main

import "fmt"

// Used to map between flags (cli) and Forjfile

type AppMapEntry struct {
	cli_obj string
	cli_instance string
	cli_field string

	forj_section string
	forj_instance string
	forj_field string
}

func (a *Forj)AddMap(key, cli_obj, cli_instance, cli_field, forj_section, forj_instance, forj_field string) {
	if a.appMapEntries == nil {
		a.appMapEntries = make(map[string]AppMapEntry)
	}
	a.appMapEntries[key] = AppMapEntry{
		cli_obj: cli_obj,
		cli_instance: cli_instance,
		cli_field: cli_field,
		forj_section: forj_section,
		forj_instance: forj_instance,
		forj_field: forj_field,
	}
}


// GetPrefs return a value found (or not) from different source of data
// 1: from cli, then exit if found
// 2: from Forjfile if found
// 3: Then cli default
func (a *Forj)GetPrefs(field string) (string, bool, error) {
	var entry AppMapEntry

	if e, found := a.appMapEntries[field] ; !found {
		return "", false, fmt.Errorf("Unable to get '%s' from Forjfile/cli mapping. Missing", field)
	} else {
		entry = e
	}

	v, found, isdefault, err := a.cli.GetStringValue(entry.cli_obj, entry.cli_instance, entry.cli_field)
    if err != nil {
		return v, found, err
	}
	if found && !isdefault {
		return v, found, err
	}
	if v, found := a.f.Get(entry.forj_section, entry.forj_instance, entry.forj_field); found {
		return v, found, nil
	}
	return v, found, err
}

// GetForgePrefs return a value found (or not) from Forjfile
func (a *Forj)GetForgePrefs(field string) (v string, found bool, _ error) {
	var entry AppMapEntry

	if e, found := a.appMapEntries[field]; !found {
		return "", false, fmt.Errorf("Unable to get '%s' from Forjfile/cli mapping. Missing", field)
	} else {
		entry = e
	}

	v, found = a.f.Get(entry.forj_section, entry.forj_instance, entry.forj_field)
	return
}

func (a *Forj)SetPrefs(field, value string) error {
	var entry AppMapEntry

	if e, found := a.appMapEntries[field]; !found {
		return  fmt.Errorf("Unable to get '%s' from Forjfile/cli mapping. Missing", field)
	} else {
		entry = e
	}

	a.f.Set(entry.forj_section, entry.forj_instance, entry.forj_field, value)
	return nil
}
