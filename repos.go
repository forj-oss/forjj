package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"forjj/drivers"
	"forjj/git"
)

const forjj_repo_file = "forjj-repos.yml"

// BuildReposList Load defaults, then load repos from source then add from cli.
func (a *Forj) BuildReposList(action string) error {
	// Set forjj-options defaults for new repositories.
	a.SetDefault(action)

	// Add cli repos list.
	a.AddReposFromCli()

	return nil
}

// GetReposRequestedFor Identify number of repository requested for an instance.
func (a *Forj) GetReposRequestedFor(instance, action string) (num int) {
	if instance == "" || action == "" {
		gotrace.Trace("Internal error: instance and action cannot be empty.")
		return
	}
	for _, data := range a.cli.GetObjectValues("repo") {
		if v, _, _ := data.Get("instance"); v == instance || (v == "" && instance == a.o.Defaults["instance"]) {
			num++
		}
	}
	return
}

// GetReposData Function providing a PluginRepoData content for the instance given.
func (a *Forj) GetReposData(instance string) (ret map[string]goforjj.PluginRepoData) {
	gotrace.Trace("Forjj managed %d repositories (forjj-repos.yml)", len(a.r.Repos))
	ret = make(map[string]goforjj.PluginRepoData)
	for n, d := range a.r.Repos {
		if d.Instance != instance {
			continue
		}
		ret[n] = *d
	}
	gotrace.Trace("%d repositories identified for instance %s", len(ret), instance)
	return
}

// SaveManagedRepos Stored Repositories managed by the plugin in the list of repos (forjj-repos.yaml)
func (a *Forj) SaveManagedRepos(d *drivers.Driver, instance string) {
	for name, repo := range a.r.Repos {
		if _, found := d.Plugin.Result.Data.Repos[name]; found {
			// Saving infra repository information to the workspace
			repo.Instance = instance
		}
	}
}

// Update a Repolist from another list.
// If new, added. If both exist, update from source.
func (r *ReposList) UpdateFromList(source map[string]*cli.ForjData, defaults map[string]string) {
	for name, repodata := range source {
		repo := r.create_repo_data(repodata)
		repo.SetDefaults(defaults)
		if d, found := r.Repos[name]; found {
			d.UpdateFrom(repo)
		} else {
			r.Repos[name] = repo
		}
	}
}

// Get cli repo object data and create the PluginRepoData.
func (r *ReposList) create_repo_data(repodata *cli.ForjData) (repo *goforjj.PluginRepoData) {
	repo = goforjj.NewRepoData()
	for kname, value := range repodata.Attrs() {
		switch kname {
		case "instance":
			repo.Instance = value.(string)
		case "title":
			repo.Title = value.(string)
		case "flow":
			repo.Flow = value.(string)
		case "repo-template":
			repo.Templates = append(repo.Templates, value.(string))
		}
	}
	return
}

func NumDisplay(num int, format, elements, element string) string {

	if num > 1 {
		return fmt.Sprintf(format, num, elements)
	}
	return fmt.Sprintf(format, num, element)
}

func NumReposDisplay(num int) string {
	return NumDisplay(num, "%d repositor%s", "ies", "y")
}

// Function to update Forjj Repos list. Use RepodSave to Save it as code ie in forjj-repos.yml.
// NOTE: a repo can be only created. Never updated or deleted. A repo has his own lifecycle not managed by forjj.
func (a *Forj) AddReposFromCli() {
	gotrace.Trace("Forjj managed %s.", NumReposDisplay(a.f.ObjectLen("repo")))

	cli_repos := a.cli.GetObjectValues("repo")
	for name, repo := range cli_repos {
		//
		a.f.SetHandler("repo", name,
			func (key string) (string, bool) {
				v := repo.GetString(key)
				if v == "" {
					v, _ = a.f.Get("repo", name, key)
				}
				return v, (v != "")
			},
			repo.Keys() ...,
		)
	}

	gotrace.Trace("Now, Forjj manages %s. including cli added %s.",
		NumReposDisplay(a.f.ObjectLen("repo")), NumReposDisplay(len(cli_repos)))
}

// RepoCodeSave Function to save forjj list of Repositories.
func (a *Forj) RepoCodeSave() (err error) {
	if yd, err := yaml.Marshal(a.r); err == nil {
		if err := ioutil.WriteFile(forjj_repo_file, yd, 0644); err != nil {
			return fmt.Errorf("Unable to write '%s'. %s", forjj_repo_file, err)
		}
	} else {
		return fmt.Errorf("Unable to encode to yaml '%s'. %s", forjj_repo_file, err)
	}

	gotrace.Trace("%s written with %s.", forjj_repo_file, NumReposDisplay(len(a.r.Repos)))

	git.Do("add", forjj_repo_file)
	return nil
}

// RepoCodeLoad Read the collection of repositories managed by forjj.
func (a *Forj) RepoCodeLoad() error {
	a.r.Repos = make(map[string]*goforjj.PluginRepoData)

	if _, err := os.Stat(forjj_repo_file); err != nil {
		gotrace.Trace("%s not found. %s.", forjj_repo_file, err)
		return nil
	}
	if d, err := ioutil.ReadFile(forjj_repo_file); err == nil {
		if err := yaml.Unmarshal(d, a.r); err != nil {
			return fmt.Errorf("Unable to decode '%s'. %s", forjj_repo_file, err)
		}
	} else {
		return fmt.Errorf("Unable to read '%s'. %s", forjj_repo_file, err)
	}

	gotrace.Trace("%s loaded from forjj-repos.yml", NumReposDisplay(len(a.r.Repos)))

	return nil
}

// Function to create missing repositories in the upstream defined.
// It should do:
// - create missing repos
// - set appropriate local repo config (upstream) depending on flows definition.
func (a *Forj) RepoMaintain() {
	for name, repo := range a.r.Repos {
		// Build the tree and save then (git).
		// mkdir in repos/{repo}/repo.yaml
		gotrace.Trace("Create Repo %s on instance %s", name, repo.Instance)
		d := a.DriverGet(repo.Instance)
		if d == nil {
			log.Printf("Unable to create code for Repo '%s'. Instance '%s' not found. Ignored.", name, repo.Instance)
			continue
		}
		// Ask upstream driver to create the repo. Except if the driver is none
		// Expect flow to be used

		// Create local repo

		// Sync with upstream if not "none"
	}

}
