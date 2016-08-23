package main

import (
    "gopkg.in/alecthomas/kingpin.v2"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
)


func (a *Forj)Update() {
    err := a.define_infra_upstream("update")
    kingpin.FatalIfError(err, "Driver update issue. Unable to identify a valid infra repository upstream.")

    gotrace.Trace("Infra upstream selected: '%s'", a.w.Instance)

    // Load Workspace information
    err = a.w.Load(a)
    kingpin.FatalIfError(err, "Driver create issue. failed to create '%s' source files.", a.w.Instance)

    // TODO: Add update code

    println("FORJJ - update", a.w.Organization, "DONE")
    // save infra repository location in the workspace.
    a.w.Save(a)

}

