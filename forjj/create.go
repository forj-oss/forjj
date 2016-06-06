package main

import (
        "fmt"
        "os"
        "strings"
        "os/exec"
        "bufio"
        "io"
        "syscall"
)

// Call docker to create the Solution source code from scratch with validated parameters.
// This container do the real stuff (git/call drivers)
// I would expect to have this go tool to have a do_create to replace the shell script.
// But this would be a next version and needs to be validated before this decision is made.
func (a *Forj) Create() {
 cmd_args := append([]string{}, "sudo", "docker", "run", "-i", "--rm")
 cmd_args = append(cmd_args, "-v", fmt.Sprintf("%s:/home/devops/.ssh", *a.CurrentCommand.flagsv[ssh_dir_flag_name]))
 cmd_args = append(cmd_args, "-v", fmt.Sprintf("%s/%s:/devops", a.Workspace_path, a.Workspace))

 if a.contrib_repo_path == "" {
   cmd_args = append(cmd_args, "-v", fmt.Sprintf("%s-forjj-contribs:/forjj-contribs", *a.Orga_name))
 } else {
   cmd_args = append(cmd_args, "-v", fmt.Sprintf("%s:/forjj-contribs", a.contrib_repo_path))
 }

 cmd_args = append(cmd_args, Docker_image, "create")

 cmd_args = a.GetDriversParameters(cmd_args, "common")
 cmd_args = a.GetDriversParameters(cmd_args, "create")

 cmd := exec.Command(cmd_args[0], cmd_args[1:]...)
 fmt.Printf("FORJJ - RUNNING: %s\n\n", strings.Join(cmd_args, " "))

 stdout, _ := cmd.StdoutPipe()
 stderr, _ := cmd.StderrPipe()

 in := bufio.NewScanner(io.MultiReader(stdout, stderr))
 cmd.Start()

 for in.Scan() {
   println(in.Text())
  }
 cmd.Wait()
 if status := cmd.ProcessState.Sys().(syscall.WaitStatus) ; status.ExitStatus() != 0 {
    fmt.Printf("\nFORJJ - create %s ERROR.\nCommand status: %s\n", *a.Orga_name, cmd.ProcessState.String())
    os.Exit(status.ExitStatus())
 }
 println("FORJJ - create", *a.Orga_name, "DONE") // , cmd.ProcessState.Sys().WaitStatus)
}

