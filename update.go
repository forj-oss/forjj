package main

import (
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "fmt"
    "log"
)


func (a *Forj)Update() error {
    if err := a.define_infra_upstream("update") ; err != nil {
        return fmt.Errorf("Unable to identify a valid infra repository upstream. %s", err)
    }

    gotrace.Trace("Infra upstream selected: '%s'", a.w.Instance)

    // Load Workspace information
    if err := a.w.Load(a) ; err != nil {
        return fmt.Errorf("Failed to create '%s' source files. %s", a.w.Instance, err)
    }

    // TODO: Add update code

    // save infra repository location in the workspace.
    defer a.w.Save(a)

    log.Print("FORJJ - update", a.w.Organization, "DONE")
    return nil
}

