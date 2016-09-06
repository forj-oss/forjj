package main

import (
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "fmt"
    "log"
    "regexp"
)


func (a *Forj)Update() error {
    if err := a.define_infra_upstream("update") ; err != nil {
        return fmt.Errorf("Unable to identify a valid infra repository upstream. %s", err)
    }

    gotrace.Trace("Infra upstream selected: '%s'", a.w.Instance)

    // save infra repository location in the workspace.
    defer a.w.Save(a)

    defer a.driver_cleanup(a.w.Instance) // Ensure upstream instances will be shutted down when done.

    if err, aborted := a.ensure_infra_exists("update") ; err != nil {
        if !aborted {
            return fmt.Errorf("Failed to ensure infra exists. %s", err)
        }
        log.Printf("Warning. %s", err)
    }

    // Now, we are in the infra repo root directory and at least, the 1st commit exist.

    // TODO: flow_start to execute instructions before updating source code for existing apps in appropriate branch. Possible if a flow is already implemented otherwise git must stay in master branch
    // flow_start()

    if err := a.MoveToFixBranch(*a.Actions["update"].argsv["branch"]) ; err != nil {
        return fmt.Errorf("Unable to move to your feature branch. %s", err)
    }

    a.o.update_options()

    // Loop on drivers requested like jenkins classified as ci type.
    for instance, _ := range a.drivers {
        if instance == a.w.Instance {
            continue // Do not try to create infra-upstream twice.
        }

        defer a.driver_cleanup(instance) // Ensure all instances will be shutted down when done.

        if err, aborted := a.do_driver_task("update", instance) ; err != nil {
            if !aborted {
                return fmt.Errorf("Failed to update '%s' source files. %s", instance, err)
            }
            log.Printf("Warning. %s", err)
        }
    }

    a.o.SaveForjjOptions(fmt.Sprintf("Organization %s updated.", a.w.Organization))

    // Start working on repositories, writing repos source code.
    a.RepoCodeBuild("update")

    // TODO: Implement flow_close() to close the create task
    // flow_close()

    log.Printf("As soon as you are happy with your fixes, do a git push to submit your collection of fixes related to '%s' to your team.", a.Branch)
    return nil
}

func (a *Forj)MoveToFixBranch(branch string) error{
    a.Branch = branch

    if ok, _ := regexp.MatchString(`^[\w_-]+$`, branch) ; ! ok {
        return fmt.Errorf("Invalid git branch name '%s'. alphanumeric, '_' and '-' are supported.", branch)
    }
    return nil
}
