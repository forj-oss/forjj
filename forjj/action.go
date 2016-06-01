package main

import (
        "fmt"
        "os"
        "strings"
        "os/exec"
        "path"
        "bufio"
        "io"
        "syscall"
)

// Call docker to create the Solution from scratch with validated parameters.
func (v *ForjContext) Create(organization string, contrib_repo string, ssh_dir string) {
 v.Organization = path.Base(organization)
 v.Orga_path = path.Dir(organization)


 cmd_args := append([]string{}, "sudo", "docker", "run", "-i", "--rm")
 cmd_args = append(cmd_args, "-v")
 cmd_args = append(cmd_args, fmt.Sprintf("%s:/home/devops/.ssh", ssh_dir))
 cmd_args = append(cmd_args, "-v")
 cmd_args = append(cmd_args, fmt.Sprintf("%s:/devops", v.Organization))

 cmd_args = append(cmd_args, "-v")
 if contrib_repo == "" {
   cmd_args = append(cmd_args, fmt.Sprintf("%s-forjj-contribs:/forjj-contribs", v.Organization))
 } else {
   cmd_args = append(cmd_args, fmt.Sprintf("%s:/forjj-contribs", contrib_repo))
 }

 cmd_args = append(cmd_args, Docker_image)
 cmd_args = append(cmd_args, "/usr/local/bin/forjj-create.sh")

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
    fmt.Printf("\nFORJJ - create %s ERROR.\nCommand status: %s\n", v.Organization, cmd.ProcessState.String())
    os.Exit(status.ExitStatus())
 }
 println("FORJJ - create", v.Organization, "DONE") // , cmd.ProcessState.Sys().WaitStatus)
}

