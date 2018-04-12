package drivers

import (
	"fmt"
	"forjj/git"
	"os"
	"path"

	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

// GitAddPluginFiles Add Plugins generated files to ready to be commit git space.
//
// It requires a function to change the current path to the appropriate repo to add files
// This function must accepts only goforjj.FilesSource and goforjj.FilesDeploy
//
func (d *Driver) GitAddPluginFiles(moveTo func(string) (string, error)) error {
	if d.Plugin.Result == nil {
		return fmt.Errorf("Strange... The plugin as no result (plugin.Result is nil). Did the plugin '%s' executed?", d.Name)
	}

	if len(d.Plugin.Result.Data.Files) == 0 {
		return fmt.Errorf("Nothing to commit")
	}

	if d.Plugin.Result.Data.CommitMessage == "" {
		return fmt.Errorf("Unable to commit without a commit message")
	}

	for where, files := range d.Plugin.Result.Data.Files {
		if err := RunInPath(where, moveTo, func() error {
			gotrace.Trace("GIT: Adding %s %d files related to '%s'", where, len(files), d.Plugin.Result.Data.CommitMessage)

			return d.gitAddPluginFiles(where, files)
		}); err != nil {
			return err
		}
	}

	return nil
}

// RunInPath run a function in a specificDirectory and restore the current Path.
func RunInPath(where string, moveTo func(string) (string, error), runIn func() error) error {
	if where != goforjj.FilesDeploy && where != goforjj.FilesSource { // Supports only 2 kind of repository from the plugin.
		return fmt.Errorf("Plugin error: Invalid repository type '%s'. Valid one are: %s and %s. Check with the plugin maintainer", where, goforjj.FilesDeploy, goforjj.FilesSource)
	}

	if restore, err := moveTo(where); err != nil {
		return err
	} else {
		if err = runIn(); err != nil {
			return err
		}
		if err = os.Chdir(restore); err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) gitAddPluginFiles(where string, files []string) error {
	if files == nil {
		return nil
	}
	fileToAdd := make([]string, len(files))
	for iCount, file := range files {
		if where == goforjj.FilesSource {
			fileToAdd[iCount] = path.Join("apps", d.DriverType, file)
		} else {
			fileToAdd[iCount] = file
		}
	}
	if i := git.Add(fileToAdd); i > 0 {
		return fmt.Errorf("Issue while adding code to git. RC=%d", i)
	}
	return nil
}
