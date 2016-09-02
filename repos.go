package main

import (
    //"github.hpe.com/christophe-larsonneur/goforjj"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "log"
    "os"
    "path"
    "fmt"
)

type RepoStruct struct {
    //goforjj.PluginRepo                // Plugin Repo data
    Templates     RepoTemplatesStruct // RepoTemplates to apply
    Title         string              // Repo Description
    Users         map[string]string   // Users and rights given
    Groups        map[string]string   // Groups and rights given
    Flow          string              // Flow applied to the Repo
}

/*func (r *RepoStruct)UpdateFrom(plugin *goforjj.PluginData) {
    if d, found := plugin.Repos[r.Name] ; found {
        r.Remotes = d.Remotes
        r.BranchConnect = d.BranchConnect
        r.Flow = d.Flow
        r.Exist = d.Exist
    }
}*/

// Function to create source files in the infra repository
func (a *Forj)RepoCodeBuild() error {
    // Loop on Repositories stored in infra-repo/forjj_options.yaml
    for name, instance := range a.o.Repos {
        // Build the tree and save then (git).
        // mkdir in repos/{repo}/repo.yaml
        gotrace.Trace("Generating Repo files : %s on instance %s", name, instance)
        d := a.DriverGet(instance)
        if d == nil {
            log.Printf("Unable to create code for Repo '%s'. Instance '%s' not found. Ignored.", name, instance)
            continue
        }

        // Create repos/<Repo> directory structure
        w := path.Join("repos", name)
        if err := os.MkdirAll(w, 0755) ; err != nil {
            return fmt.Errorf("Unable to create '%s'", w, err)
        }

        // Save a repo_def.yaml
        // To define
        // template:
        // title:
        // users:
        // groups:
        file := path.Join(w, "repo.yaml")

        git("add", file)
    }
    return nil
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
    for name, instance := range a.o.Repos {
        // Build the tree and save then (git).
        // mkdir in repos/{repo}/repo.yaml
        gotrace.Trace("Create Repo %s on instance %s", name, instance)
        d := a.DriverGet(instance)
        if d == nil {
            log.Printf("Unable to create code for Repo '%s'. Instance '%s' not found. Ignored.", name, instance)
            continue
        }
        // Ask upstream driver to create the repo. Except if the driver is none
        // Expect flow to be used

        // Create local repo

        // Sync with upstream if not "none"
    }

}
