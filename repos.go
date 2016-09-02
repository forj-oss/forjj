package main

import (
    "github.hpe.com/christophe-larsonneur/goforjj"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
)

type RepoStruct struct {
    goforjj.PluginRepo
    Templates RepoTemplatesStruct
}

func (r *RepoStruct)UpdateFrom(plugin *goforjj.PluginData) {
    if d, found := plugin.Repos[r.Name] ; found {
        r.Remotes = d.Remotes
        r.BranchConnect = d.BranchConnect
        r.Flow = d.Flow
        r.Exist = d.Exist
    }
}

// Function to create source files in the infra repository
func (o *ForjjOptions)RepoCodeBuild(){
    // Loop on Repositories stored in infra-repo/forjj_options.yaml
    for name, instance := range o.Repos {
        // Build the tree and save then (git).
        // mkdir in repos/{repo}/repo.yaml
        gotrace.Trace("Generate Repo files : %s on instance %s", name, instance)
    }
}

// Function to create missing repositories in the upstream defined.
// It should do:
// - create missing repos
// - set appropriate local repo config (upstream) depending on flows definition.
func (o *ForjjOptions)RepoMaintain(){

}
