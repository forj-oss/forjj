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

 // Ensure remote upstream exists - calling driver

 // Ensure local repo upstream properly configured

 // git add/commit and push

 println("FORJJ - create", *a.Orga_name, "DONE") // , cmd.ProcessState.Sys().WaitStatus)
}

/*
 cmd_args = append(cmd_args, "-v", fmt.Sprintf("%s:/home/devops/.ssh", *a.CurrentCommand.flagsv[ssh_dir_flag_name]))
 cmd_args = append(cmd_args, "-v", fmt.Sprintf("%s/%s:/devops", a.Workspace_path, a.Workspace))

 if a.contrib_repo_path == "" {
   cmd_args = append(cmd_args, "-v", fmt.Sprintf("%s-forjj-contribs:/forjj-contribs", *a.Orga_name))
 } else {
   cmd_args = append(cmd_args, "-v", fmt.Sprintf("%s:/forjj-contribs", a.contrib_repo_path))

 cmd_args = a.GetDriversParameters(cmd_args, "common")
 cmd_args = a.GetDriversParameters(cmd_args, action)
*/
