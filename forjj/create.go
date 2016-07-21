package main

import (
    "fmt"
    "os"
)

// Call docker to create the Solution source code from scratch with validated parameters.
// This container do the real stuff (git/call drivers)
// I would expect to have this go tool to have a do_create to replace the shell script.
// But this would be a next version and needs to be validated before this decision is made.
func (a *Forj) Create() {
    // Ensure upstream driver is given
    if _, ok := a.drivers["upstream"] ; ! ok {
        fmt.Printf("Missing upstream driver. Please use --git-us\n")
        os.Exit(1)
    }

    // Ensure local repo exists
    a.ensure_local_repo(a.w.Infra)

    // Create source for the infra repository - Calling upstream driver - create
    a.driver_do("upstream", "create")

    // Ensure remote upstream exists - calling upstream driver - maintain
    //a.driver_do("upstream", "maintain") // This will create/update the upstream service

    // Ensure local repo upstream properly configured
    //a.ensure_remote_repo(a.w.Infra)

    // git add/commit and push
    // ??

    println("FORJJ - create", a.w.Organization, "DONE") // , cmd.ProcessState.Sys().WaitStatus)
    // save infra repository location in the workspace.
    a.w.Save(a)
}

/*
 cmd_args = append(cmd_args, "-v", fmt.Sprintf("%s:/home/devops/.ssh", *a.CurrentCommand.flagsv[ssh_dir_flag_name]))
 cmd_args = append(cmd_args, "-v", fmt.Sprintf("%s/%s:/devops", a.Workspace_path, a.Workspace))

 if a.contrib_repo_path == "" {
   cmd_args = append(cmd_args, "-v", fmt.Sprintf("%s-forjj-contribs:/forjj-contribs", a.w.Organization))
 } else {
   cmd_args = append(cmd_args, "-v", fmt.Sprintf("%s:/forjj-contribs", a.contrib_repo_path))

 cmd_args = a.GetDriversParameters(cmd_args, "common")
 cmd_args = a.GetDriversParameters(cmd_args, action)
*/
