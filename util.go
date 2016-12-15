package main

import (
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/kvz/logstreamer"
	"github.com/forj-oss/forjj-modules/trace"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

type GitStatus struct {
	Added     []string
	Modified  []string
	Untracked []string
	err       error
}

// Simple function to convert a dynamic type to bool
// it returns false by default except if the internal type is:
// - bool. value as is
// - string: call https://golang.org/pkg/strconv/#ParseBool
//
func to_bool(v interface{}) bool {
	switch v.(type) {
	case bool:
		return v.(bool)
	case string:
		s := v.(string)
		if b, err := strconv.ParseBool(s); err == nil {
			return b
		}
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

// Do a git commit
func git_commit(msg string, errorIfEmpty bool) error {
	s := git_status()
	if len(s.Added) == 0 {
		return fmt.Errorf("No files added. Please check.")
	}
	if git("commit", "-m", msg) > 0 {
		return fmt.Errorf("Unable to commit.")
	}
	return nil
}

// Push latest commits
func gitPush() error {
	if git("push") > 0 {
		return fmt.Errorf("Unable to push commits.")
	}
	return nil
}

func DefColor(colorNum int) (color, reset string) {
	if strings.HasPrefix(os.Getenv("TERM"), "xterm") {
		color = fmt.Sprintf("\x1b[%dm", colorNum)
		reset = "\x1b[0m"
	}
	return
}

// Call git command with arguments. All print out displayed. It returns git Return code.
func git(opts ...string) int {
	colorCyan, colorReset := DefColor(36)
	log.Printf("%sgit %s%s\n", colorCyan, strings.Join(opts, " "), colorReset)
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
		if m := untracked_re.FindStringSubmatch(line); m != nil {
			gs.Untracked = append(gs.Untracked, m[1])
		}
		if m := modified_re.FindStringSubmatch(line); m != nil {
			gs.Modified = append(gs.Modified, m[1])
		}
		if m := added_re.FindStringSubmatch(line); m != nil {
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
func run_cmd(command string, args ...string) int {
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

func touch(file string) error {
	if fd, err := os.Create(file); err == nil {
		return fd.Close()
	} else {
		return err
	}
}

// Function to read a document from a url like github raw or directly from a local path
func read_document_from(s *url.URL) (yaml_data []byte, err error) {
	if s.Scheme == "" {
		// File to read locally
		return read_document_from_fs(s.Path)
	}
	// File to read from an url. Usually, a raw from github.
	return read_document_from_url(s.String())
}

// Read from the filesystem. If the path start with ~, replaced by the user homedir. In some context, this won't work well, like in container.
func read_document_from_fs(source string) (yaml_data []byte, err error) {
	// File to read locally
	if source[:1] == "~" {
		cur_user := &user.User{}
		if cur_user, err = user.Current(); err != nil {
			err = fmt.Errorf("Unable to get your user. %s. Consider to replace ~ by $HOME\n", err)
			return
		}
		source = string(regexp.MustCompile("^~").ReplaceAll([]byte(source), []byte(cur_user.HomeDir)))
	}
	gotrace.Trace("Load file definition at '%s'", source)
	return ioutil.ReadFile(source)
}

// Read from the URL string. Data is returned is content type is of text/plain
func read_document_from_url(source string) (yaml_data []byte, err error) {
	gotrace.Trace("Load file definition at '%s'", source)

	var resp *http.Response
	if resp, err = http.Get(source); err != nil {
		err = fmt.Errorf("Unable to read '%s'. %s\n", source, err)
		return
	}
	defer resp.Body.Close()

	var d []byte
	if d, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	if strings.Contains(http.DetectContentType(d), "text/plain") {
		yaml_data = d
	}
	return
}

func MapBoolKeys(m map[string]bool) (a []string) {
	a = make([]string, 0, len(m))
	for key := range m {
		a = append(a, key)
	}
	return a
}

func arrayStringDelete(a []string, element string) []string {
	for index, value := range a {
		if value == element {
			return append(a[:index], a[index+1:]...)
		}
	}
	return a
}

func inStringList(element string, elements ...string) string {
	for _, value := range elements {
		if element == value {
			return value
		}
	}
	return ""
}
