package main

import (
    "os"
    "os/exec"
    "strconv"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "strings"
    "syscall"
    "github.com/kvz/logstreamer"
    "log"
    "regexp"
    "gopkg.in/alecthomas/kingpin.v2"
)


type GitStatus struct {
    Added []string
    Modified []string
    Untracked []string
    err error
}

// Simple function to convert a dynamic type to bool
// it returns false by default except if the internal type is:
// - bool. value as is
// - string: call https://golang.org/pkg/strconv/#ParseBool
//
func to_bool(v interface{}) (bool) {
 switch v.(type) {
   case bool:
     return v.(bool)
   case string:
     s := v.(string)
     if b, err := strconv.ParseBool(s) ; err == nil { return b }
     return false
 }
 return false
}

// simply extract string from the dynamic type
// otherwise the returned string is empty.
func to_string(v interface{}) (result string) {
 switch v.(type) {
   case string:
     return v.(string)
 }
 return
}

// Call git command with arguments. All print out displayed. It returns git Return code.
func git(opts ...string) int {
    return run_cmd("git", opts...)
}

func git_status() (gs *GitStatus) {
    gs = new(GitStatus)
    gs.Added = make([]string, 0, 2)
    gs.Modified = make([]string, 0, 2)
    gs.Untracked = make([]string, 0, 2)
    added_re, _ := regexp.Compile("^[A-Z]  (.*)$")
    modified_re, _ := regexp.Compile("^ [A-Z] (.*)$")
    untracked_re, _ := regexp.Compile(`^\?\? (.*)$`)

    var s string

    s, gs.err = git_get("status", "--porcelain")
    if gs.err != nil {
        return
    }

    lines := strings.Split(s, "\n")

    for _, line := range lines {
        if m := untracked_re.FindStringSubmatch(line) ; m != nil {
            gs.Untracked = append(gs.Untracked, m[1])
        }
        if m := modified_re.FindStringSubmatch(line) ; m != nil {
            gs.Modified = append(gs.Modified, m[1])
        }
        if m := added_re.FindStringSubmatch(line) ; m != nil {
            gs.Added = append(gs.Added, m[1])
        }
    }
    return
}

// Call a git command and get the output as string output.
func git_get(opts ...string) (string, error) {
    gotrace.Trace("RUNNING: git %s", strings.Join(opts, " "))
    out, err := exec.Command("git", opts...).Output()
    return string(out), err
}

// Simple function to call a shell command and display to stdout
// stdout is displayed as is when it arrives, while stderr is displayed in Red, line per line.
func run_cmd(command string, args ...string) int{
    logger := log.New(os.Stdout, "", log.LstdFlags)
    // Setup a streamer that we'll pipe cmd.Stdout to
    logStreamerOut := logstreamer.NewLogstreamer(logger, "stdout", false)
    defer logStreamerOut.Close()
    // Setup a streamer that we'll pipe cmd.Stderr to.
    // We want to record/buffer anything that's written to this (3rd argument true)
    logStreamerErr := logstreamer.NewLogstreamer(logger, "stderr", true)
    defer logStreamerErr.Close()

    cmd := exec.Command(command, args...)
    gotrace.Trace("RUNNING: %s %s", command, strings.Join(args, " "))

    cmd.Stderr = logStreamerErr
    cmd.Stdout = logStreamerOut

    // Reset any error we recorded
    logStreamerErr.FlushRecord()
    // Execute command
    if err := cmd.Start() ; err != nil {
        kingpin.Errorf("ERROR could not spawn command. %s.", err.Error())
        return 255
    }

    if err := cmd.Wait(); err != nil {
        kingpin.Errorf("\nERROR: wait failure - %s: %s.", command, err)
        return 1
    }
    gotrace.Trace("Command done")
    if status := cmd.ProcessState.Sys().(syscall.WaitStatus) ; status.ExitStatus() != 0 {
        kingpin.Errorf("\n%s ERROR: Unable to get process status - %s: %s", command, cmd.ProcessState.String())
        return status.ExitStatus()
    }
    return 0
}

func touch(file string) error {
    if fd, err := os.Create(file) ; err == nil {
        return fd.Close()
    } else {
        return err
    }
}
