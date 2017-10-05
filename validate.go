package main

import (
	"fmt"
	"github.com/forj-oss/goforjj"
)


func (a *Forj)Validate() error {
	return a.ValidateForjfile()
}

func (a *Forj)FoundValidAppFlag(key, driver, object string, required bool) (err error, _ bool) {
	d, _ := a.drivers[driver]
	if d == nil {
		err = fmt.Errorf("Internal issue. Driver %s not found in memory.", driver)
		return
	}
	if o, found := d.Plugin.Yaml.Objects[object]; !found {
		if required {
			err = fmt.Errorf("Plugin issue. objects/'%s' is not found. Contact Plugin maintainer.", object)
			return
		}
		return

	} else {
		return nil, o.HasValidKey(key)
	}
}

// ValidateForjfile read all object fields and check if they are recognized by forjj or plugins.
func (a *Forj)ValidateForjfile() (_ error) {
	f := a.f.Forjfile()

	// ForjSettingsStruct.More

	// RepoStruct.More (infra : Repos)

	// AppYamlStruct.More
	for _, app := range f.Apps {
		for key := range app.More {
			if err, found := a.FoundValidAppFlag(key, app.Driver, goforjj.ObjectApp, true); err != nil {
				return err
			} else if ! found {
				return fmt.Errorf("'%s' has no effect. No drivers use it.", key)
			}
		}
	}

	// UserStruct.More

	// GroupStruct.More

	// ForgeYaml.More
	return
}
