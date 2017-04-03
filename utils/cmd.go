package utils

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"github.com/kvz/logstreamer"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/alecthomas/kingpin"
)

// Simple function to call a shell command and display to stdout
// stdout is displayed as is when it arrives, while stderr is displayed in Red, line per line.
func RunCmd(command string, args ...string) int {
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
	if err := cmd.Start(); err != nil {
		kingpin.Errorf("ERROR could not spawn command. %s.", err.Error())
		return 255
	}

	if err := cmd.Wait(); err != nil {
		kingpin.Errorf("\nERROR: wait failure - %s: %s.", command, err)
		return 1
	}
	gotrace.Trace("Command done")
	if status := cmd.ProcessState.Sys().(syscall.WaitStatus); status.ExitStatus() != 0 {
		kingpin.Errorf("\n%s ERROR: Unable to get process status - %s: %s", command, cmd.ProcessState.String())
		return status.ExitStatus()
	}
	return 0
}
