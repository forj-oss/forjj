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

    // save infra repository location in the workspace.
    defer a.w.Save(a)

    defer a.driver_cleanup(a.w.Instance) // Ensure upstream instances will be shutted down when done.

    if err, aborted := a.ensure_infra_exists() ; err != nil {
        if !aborted {
            return fmt.Errorf("Failed to ensure infra exists. %s", err)
        }
        log.Printf("Warning. %s", err)
    }

    // Now, we are in the infra repo root directory and at least, the 1st commit exist.

    // Loop on drivers requested like jenkins classified as ci type.
    for instance, _ := range a.drivers {
        if instance == a.w.Instance {
            continue // Do not try to create infra-upstream twice.
        }

        defer a.driver_cleanup(instance) // Ensure all instances will be shutted down when done.

        if err, aborted := a.do_driver_create(instance) ; err != nil {
            if !aborted {
                return fmt.Errorf("Failed to create '%s' source files. %s", instance, err)
            }
            log.Printf("Warning. %s", err)
        }
    }

    log.Print("FORJJ - update ", a.w.Organization, " DONE")
    return nil
}

