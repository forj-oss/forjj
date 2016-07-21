package main

import (
    "os"
    "os/exec"
    "strconv"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "strings"
    "fmt"
    "syscall"
    "github.com/kvz/logstreamer"
    "log"
)


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



// Simple function to call a shell comand and display to stdout
// stdout is displayed as is when it arrives, while stderr is displayed in Red, line per line.
func run_cmd(command string, args ...string) int{
    logger := log.New(os.Stdout, "", 0)
    // Setup a streamer that we'll pipe cmd.Stdout to
    logStreamerOut := logstreamer.NewLogstreamer(logger, "stdout", false)
    defer logStreamerOut.Close()
    // Setup a streamer that we'll pipe cmd.Stderr to.
    // We want to record/buffer anything that's written to this (3rd argument true)
    logStreamerErr := logstreamer.NewLogstreamer(logger, "stderr", true)
    defer logStreamerErr.Close()

    cmd := exec.Command(command, args...)
    gotrace.Trace("RUNNING: %s %s\n", command, strings.Join(args, " "))

    cmd.Stderr = logStreamerErr
    cmd.Stdout = logStreamerOut

    // Reset any error we recorded
    logStreamerErr.FlushRecord()
    // Execute command
    if err := cmd.Start() ; err != nil {
        fmt.Printf("ERROR could not spawn command.", err.Error())
    }

    if err := cmd.Wait(); err != nil {
        fmt.Printf("\nERROR: wait failure\n%s: %s\n", command, err)
        return 1
    }
    gotrace.Trace("Command done\n")
    if status := cmd.ProcessState.Sys().(syscall.WaitStatus) ; status.ExitStatus() != 0 {
        fmt.Printf("\n%s ERROR: Unable to get process status.\n%s: %s\n", command, cmd.ProcessState.String())
        return status.ExitStatus()
    }
    return 0
}
