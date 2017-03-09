package main

import (
	"forjj/forjfile"
	"fmt"
)

// TODO: Be able to choose where to load one of more (merged) Forjfiles.
// TODO: Be able to load from url.

func (a *Forj)LoadForjfile() error {
	aPath := ""
	if v, found, _, err := a.cli.GetStringValue("_app","forjj", forjfile_f) ; err != nil {
		return fmt.Errorf("Unable to find '%s' flag from '%s' action. %s", forjfile_f, cr_act, err)
	} else {
		if found {
			aPath = v
		}
	}
	f, found, err := forjfile.LoadTmpl(aPath)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	// Load Forjfile to Forj internal.
	a.w.SetFrom(f.Workspace)
	a.f.SetFromTemplate(f)

	/*for instance, app := range a.f.Apps {

	}*/

	return nil
}
