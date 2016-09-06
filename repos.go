package main

import (
    //"github.hpe.com/christophe-larsonneur/goforjj"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "log"
    "os"
    "fmt"
    "gopkg.in/yaml.v2"
    "io/ioutil"
)

const forjj_repo_file = "forjj-repos.yaml"

type RepoStruct struct {
    //goforjj.PluginRepo                // Plugin Repo data
    Templates     RepoTemplatesStruct // RepoTemplates to apply
    Title         string              // Repo Description
    Users         map[string]string   // Users and rights given
    Groups        map[string]string   // Groups and rights given
    Flow          string              // Flow applied to the Repo
    Instance      string              // Instance managing the upstream repo.
}

// Stored Repositories managed by the plugin in the list of repos (forjj-repos.yaml)
func (a *Forj)SaveManagedRepos(d *Driver, instance string) {
    for name, repo := range a.r.Repos {
        if _, found := d.plugin.Result.Data.Repos[name] ; found  {
            // Saving infra repository information to the workspace
            repo.Instance = instance
        }
    }
}

// Update a Repolist from another list.
// If new, added. If both exist, update from source.
func (r *ReposList)UpdateFromList(source *ReposList, defaults *DefaultsStruct) {
    for name, repo := range source.Repos {
        if d, found := r.Repos[name] ; found {
            repo.SetDefaults(defaults)
            d.UpdateFrom(repo)
        } else {
            repo.SetDefaults(defaults)
            r.Repos[name] = repo
        }
    }
}

func (r *RepoStruct)SetDefaults(defaults *DefaultsStruct) {
    if r.Flow == "" {
        r.Flow = defaults.Flow
    }
    if r.Instance == "" {
        r.Instance = defaults.Instance
    }
}

func (r *RepoStruct)UpdateFrom(source *RepoStruct) {
    if source.Title != "" {
        r.Title = source.Title
    }
    if source.Users != nil && len(source.Users) >0 {
        r.Users = source.Users
    }
    if source.Flow != "" {
        r.Flow = source.Flow
    }
    if source.Groups != nil && len(source.Groups) >0 {
        r.Groups = source.Groups
    }
    if source.Instance != "" {
        r.Instance = source.Instance
    }
}

// Function to create source files in the infra repository
// NOTE: a repo can be only created. Never updated or deleted. A repo has his own lifecycle not managed by forjj.
func (a *Forj)RepoCodeBuild(action string) (err error) {
    a.r.UpdateFromList(a.Actions[action].repoList, &a.o.Defaults)

    if yd, err := yaml.Marshal(a.r) ; err == nil {
        if err := ioutil.WriteFile(forjj_repo_file, yd, 0644) ; err != nil {
            return fmt.Errorf("Unable to write '%s'. %s", forjj_repo_file, err)
        }
    } else {
        return fmt.Errorf("Unable to encode to yaml '%s'. %s", forjj_repo_file, err)
    }

    git("add", forjj_repo_file)
    return nil
}

// Read the collection of repositories managed by forjj.
func (a *Forj)RepoCodeLoad() (error) {
    a.r.Repos = make(map[string]*RepoStruct)

    if _, err := os.Stat(forjj_repo_file) ; err != nil {
        return nil
    }
    if d, err := ioutil.ReadFile(forjj_repo_file) ; err == nil {
        if err := yaml.Unmarshal(d, a.r) ; err != nil {
            return fmt.Errorf("Unable to decode '%s'. %s", forjj_repo_file, err)
        }
    } else {
        return fmt.Errorf("Unable to read '%s'. %s", forjj_repo_file, err)
    }
    return  nil
}

func (a *Forj)DriverGet(instance string) (d *Driver){
    var found bool

    if d, found = a.drivers[instance] ; found {
            return
        }

    if d, found = a.o.Drivers[instance] ; ! found {
        a.drivers[instance] = &Driver{
            Name:          d.Name,
            DriverType:    d.DriverType,
            InstanceName:  d.InstanceName,
        }
        return
    }
    return nil
}

// Function to create missing repositories in the upstream defined.
// It should do:
// - create missing repos
// - set appropriate local repo config (upstream) depending on flows definition.
func (a *Forj)RepoMaintain(){
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
