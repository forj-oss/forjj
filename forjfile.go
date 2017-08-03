package main

import (
	"forjj/forjfile"
	"fmt"
	"forjj/utils"
)

// TODO: Be able to choose where to load one of more (merged) Forjfiles.
// TODO: Be able to load from url.

func (a *Forj)LoadForjfile(action string) error {
	if v, found, _, err := a.cli.GetStringValue("_app","forjj", forjfile_path_f) ; err != nil {
		return fmt.Errorf("Unable to find '%s' flag from '%s' action. %s", forjfile_path_f, action, err)
	} else {
		if found {
			a.forjfile_tmpl_path, _ = utils.Abs(v)
		}
	}
	f, found, err := forjfile.LoadTmpl(a.forjfile_tmpl_path)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	// Load Forjfile to Forj internal.
	a.w.SetFrom(f.Workspace)
	a.f.SetFromTemplate(f)

	return nil
}
