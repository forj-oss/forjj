package main

import (
	"fmt"
	"github.com/forj-oss/goforjj"
	"github.com/forj-oss/forjj-modules/trace"
	"log"
)

// This function will ensure minimal git repo exists to store resources plugins data files.
// It will take care of several GIT scenarios. See ensure_local_repo_synced for details
// Used by create/update actions only.
// In case of create a commit must be created but the push will be possible only when the upstream will be created through maintain step.
// If the repo already exist from the upstream, it will be simply restored.
func (a *Forj) ensure_infra_exists(action string) (err error, aborted, new_infra bool) {
	defer gotrace.Trace("Exiting ensure_infra_exists")

	if err = a.ensure_local_repo_initialized(a.w.Infra.Name); err != nil {
		err = fmt.Errorf("Unable to ensure infra repository gets initialized. %s.", err)
		return
	}

	// Now, we are in the infra repo root directory. But at least is completely empty.

	// Build Managed Forjj Repos list in memory.
	err = a.BuildReposList(action)
	if err != nil {
		return
	}

	// Set the Initial data & README.md content for the infra repository.
	if _, found := a.r.Repos[a.w.Infra.Name]; !found {
		gotrace.Trace("Defining your Infra-repository %s", a.w.Infra.Name)
		// TODO: Refer to a repotemplate to create the README.md content and file.
		r := goforjj.PluginRepoData{
			Title:     fmt.Sprintf("Infrastructure Repository for the organization %s", a.w.Organization),
			Instance:  a.w.Instance,
			Templates: make([]string, 0),
			Users:     make(map[string]string),
			Groups:    make(map[string]string),
			Options:   make(map[string]string),
		}
		if v, found := a.o.Defaults["flow"]; found {
			r.Flow = v
		}
		a.r.Repos[a.w.Infra.Name] = &r
	}

	a.infra_readme = fmt.Sprintf("Infrastructure Repository for the organization %s", a.w.Organization)

	if a.InfraPluginDriver == nil { // NO infra upstream driver loaded and defined.
		// But should be ok if the git remote is already set.

		var remote_exist, remote_connected bool

		var hint string
		if v, _, _, _ := a.cli.GetStringValue(infra, "", "infra-upstream"); v == "" {
			hint = "\nIf you are ok with this configuration, use '--infra-upstream none' to confirm. Otherwise, please define the --apps with the upstream driver and needed flags."
		}

		remote_exist, remote_connected, err = git_remote_exist("master", "origin", a.w.Infra.Remotes["origin"])
		if err != nil {
			return
		}

		switch {
		case a.w.Instance == "": // The infra repo upstream instance has not been defined.
			msg := fmt.Sprintf("Your workspace contains your infra repository called '%s' but not connected to", a.w.Infra.Name)
			switch {
			case !remote_exist:
				err = fmt.Errorf("%s an upstream.%s", msg, hint)
			case !remote_connected:
				err = fmt.Errorf("%s a valid upstream '%s'.%s", msg, a.w.Infra.Remotes["origin"], hint)
			}

		case a.w.Instance == "none": // The infra is set with no upstream instance
			// Will create the 1st commit and nothing more.
			err = a.ensure_local_repo_synced(a.w.Infra.Name, "master", "", "", a.infra_readme)

		case a.w.Infra.Remotes["origin"] == "": // The infra upstream string is not defined
			err = fmt.Errorf("You provided the infra upstream instance name to connect to your local repository, without defining the upstream instance. please retry and use --apps to define it.")

		case a.w.Infra.Remotes["origin"] != "" && !remote_connected:
			err = a.ensure_local_repo_synced(a.w.Infra.Name, "master", "origin", a.w.Infra.Remotes["origin"], a.infra_readme)

		case a.w.Infra.Remotes["origin"] != "" && remote_connected:
			if action == "create" {
				log.Printf("The infra already exist and is connected. The automatic git push/forjj maintain is then disabled.")
				*a.no_maintain = true
			}

		}
		return
	}

	// -- Upstream driver defined --

	err, aborted = a.do_driver_task(action, a.w.Instance)

	// If an error occured, then we need to exit.
	if err != nil && !aborted {
		return
	}

	// Save the instance supporting the infra.
	if v, found := a.r.Repos[a.w.Infra.Name]; found {
		v.Instance = a.w.Instance
	} else {
		a.r.Repos[a.w.Infra.Name] = &goforjj.PluginRepoData{
			Instance: a.w.Instance,
		}
	}

	// Ok Do we have an upstream on the server side?
	// No. So, nothing else now to do.
	// If driver has initial infra files (create case), we need to commit them, then maintain it, then push.
	// REMINDER: Create/Update works on source only.
	if !a.w.Infra.Exist {
		new_infra = true
		return
	}

	if _, remote_connected, giterr := git_remote_exist("master", "origin", a.w.Infra.Remotes["origin"]); giterr != nil {
		if err != nil {
			err = fmt.Errorf("%s. %s.", err, giterr)
		}
	} else {
		if !remote_connected {
			// the upstream driver has detected that resources already exists.
			// As the remote one seems different, we must be restore the local repo content from this resource.

			// The remote INFRA exist!!! We need to restore.

			if action == "create" {
				*a.no_maintain = true
				log.Printf("Plugin instance %s(%s) informed service already exists. Nothing created. And for this instance, you will need to use update instead of create. Automatic git push/maintain is disabled.", a.w.Instance, a.w.Driver)
			} else {
				log.Printf("Plugin instance %s(%s) informed service already exists. We need to restore the workspace before doing the update.", a.w.Instance, a.w.Driver)
			}
			if e := a.ensure_local_repo_synced(a.w.Infra.Name, "master", "origin", a.w.Infra.Remotes["origin"], a.infra_readme); e != nil {
				err = fmt.Errorf("%s\n%s", err, e)
			}

			log.Printf("As the upstream service already exists, forjj has only fetched your workspace infra repository from '%s'.", a.w.Infra.Remotes["origin"])

			// Then re-apply cli default options and repos back to the existing restored code.
			a.LoadForjjOptions()

			// Build Managed Forjj Repos list in memory.
			err = a.BuildReposList(action)

			// Now in case of update task, we can re-applied the fix on the workspace restored. In case of create, the user will need to use update instead.
			if action == "update" {
				err, aborted = a.do_driver_task(action, a.w.Instance)
			}
		}
	}
	return
}
