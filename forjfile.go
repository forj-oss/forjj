package main

import "forjj/forjfile"

// TODO: Be able to choose where to load one of more (merged) Forjfiles.
// TODO: Be able to load from url.

func (a *Forj)LoadForjfile() error {
	f, found, err := forjfile.LoadTmpl("")
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
