package drivers

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/trace"
	"path"
	"forjj/git"
)

// Commit a Plugin generated files.
func (d *Driver) GitCommit() error {
	if git.Do("commit", "-m", d.Plugin.Result.Data.CommitMessage) > 0 {
		return fmt.Errorf("Unable to commit.")
	}
	return nil
}

// Add Plugins generated files to ready to be commit git space.
func (d *Driver) GitAddPluginFiles() error {
	if d.Plugin.Result == nil {
		return fmt.Errorf("Strange... The plugin as no result (plugin.Result is nil). Did the plugin '%s' executed?", d.Name)
	}

	gotrace.Trace("Adding %d files related to '%s'", len(d.Plugin.Result.Data.Files), d.Plugin.Result.Data.CommitMessage)
	if len(d.Plugin.Result.Data.Files) == 0 {
		return fmt.Errorf("Nothing to commit")
	}

	if d.Plugin.Result.Data.CommitMessage == "" {
		return fmt.Errorf("Unable to commit without a commit message.")
	}

	for _, file := range d.Plugin.Result.Data.Files {
		if i := git.Do("add", path.Join("apps", d.DriverType, file)); i > 0 {
			return fmt.Errorf("Issue while adding code to git. RC=%d", i)
		}
	}
	return nil
}
