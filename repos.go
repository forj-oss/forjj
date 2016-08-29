package main

import (
    "github.hpe.com/christophe-larsonneur/goforjj"
)

type RepoStruct struct {
    goforjj.PluginRepo
    Templates RepoTemplatesStruct
}


// Function to create source files in the infra repository
func (a *Forj)RepoSave(){

}

// Function to create missing repositories in the upstream defined.
// It should do:
// - create missing repos
// - set appropriate local repo config (upstream) depending on flows definition.
func (a *Forj)RepoMaintain(){

}
