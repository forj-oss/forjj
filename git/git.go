package git

import (
	"fmt"
	"forjj/utils"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"
)

type gitContext struct {
	end    string
	indent string
}

var context gitContext
var logFunc func(string)

func init() {
	logFunc = logOut
}

func logOut(text string) {
	log.Print(text)
}

// Define the internal Log system. By default it uses log.Print
func SetLogFunc(aLogFunc func(string)) {
	logFunc = aLogFunc
}

// Do Call git command with arguments. All print out displayed. It returns git Return code.
func Do(opts ...string) int {
	colorCyan, colorReset := utils.DefColor(36)
	logFunc(fmt.Sprintf("%s%sgit %s%s\n", colorCyan, context.indent, strings.Join(opts, " "), colorReset))
	return utils.RunCmd("git", opts...)
}

// Indent permit to display several command indented within a section tag.
func Indent(begin, indent, end string) {
	colorCyan, colorReset := utils.DefColor(36)
	logFunc(fmt.Sprintf("%s%s%s\n", colorCyan, begin, colorReset))
	context.end = end
	context.indent = indent
}

// UnIndent revert Indent.
func UnIndent() {
	colorCyan, colorReset := utils.DefColor(36)
	logFunc(fmt.Sprintf("%s%s%s\n", colorCyan, context.end, colorReset))
}

// ShowGitPath display the current GI path
func ShowGitPath() (msg string) {
	msg = " - Unable to determine the current directory."
	if p, err := os.Getwd(); err == nil {
		msg = " - " + p
	} else {
		msg += err.Error()
	}
	return
}

// GetStatus return an GitStatus struct with the list of files, added, updated and
func GetStatus() (gs *Status) {
	gs = new(Status)

	gs.Ready = make(map[string][]string)
	gs.Ready.init(false)
	gs.NotReady = make(map[string][]string)
	gs.NotReady.init(true)

	ReadyRE, _ := regexp.Compile("^([ADM])  (.*)$")
	NotReadyRE, _ := regexp.Compile("^ ([?ADM]) (.*)$")

	var s string

	s, gs.Err = Get("status", "--porcelain")
	if gs.Err != nil || s == "" {
		return
	}

	lines := strings.Split(s, "\n")

	for _, line := range lines {
		if m := ReadyRE.FindStringSubmatch(line); m != nil {
			gs.Ready.add(m[1], m[2])
		}
		if m := NotReadyRE.FindStringSubmatch(line); m != nil {
			gs.Ready.add(m[1], m[2])
		}
	}
	return
}

// Get Call a git command and get the output as string output.
func Get(opts ...string) (string, error) {
	gotrace.Trace("RUNNING: git %s", strings.Join(opts, " "))
	out, err := exec.Command("git", opts...).Output()
	return strings.Trim(string(out), " \n"), err
}

// GetWithStatusCode Call a git command and get the output as string output.
func GetWithStatusCode(opts ...string) (string, int) {
	colorCyan, colorReset := utils.DefColor(36)
	logFunc(fmt.Sprintf("%s%sgit %s%s\n", colorCyan, context.indent, strings.Join(opts, " "), colorReset))
	return utils.RunCmdOutput("git", opts...)
}

// Commit Do a git commit
func Commit(msg string, errorIfEmpty bool) (err error) {
	s := GetStatus()
	if s.Ready.CountTracked() == 0 {
		if errorIfEmpty {
			err = fmt.Errorf("No files to commit. Please check")
		}
		return
	}
	if Do("commit", "-m", msg) > 0 {
		return fmt.Errorf("Unable to commit")
	}
	return nil
}

// Push Push latest commits
func Push() error {
	if Do("push") > 0 {
		return fmt.Errorf("Unable to push commits")
	}
	return nil
}

// Add call git add
func Add(files []string) int {
	cmd := make([]string, 1, len(files)+1)
	cmd[0] = "add"
	cmd = append(cmd, files...)
	return Do(cmd...)
}

// Branches retrieved the list of branch from git branch
func Branches() ([]string, error) {
	v, err := Get("branch")
	if err != nil || v == "" {
		return []string{}, err
	}
	return strings.Split(v, "\n"), nil
}

// RemoteBranches returns the list of Remote branches found
// Formatted as <remote>/<branchName>
func RemoteBranches() ([]string, error) {
	v, err := Get("branch", "-r")
	if err != nil || v == "" {
		return []string{}, err
	}
	return strings.Split(v, "\n"), nil
}

// RemoteBranchExist check is remote branch if known by GIT.
//
// Remote: Formated as <remote>/<branchName>
func RemoteBranchExist(remote string) (bool, error) {
	branches, err := RemoteBranches()
	if err != nil {
		return false, err
	}

	for _, branch := range branches {
		if branch == remote {
			return true, nil
		}
	}
	return false, nil
}

// BranchExist return true if the branch exist
func BranchExist(remote string) (bool, error) {
	branches, err := Branches()
	if err != nil {
		return false, err
	}

	for _, branch := range branches {
		if branch == remote {
			return true, nil
		}
	}
	return false, nil
}

// RemoteStatus provide a sync status information
func RemoteStatus(remote string) (_ string, err error) {
	var localRev, remoteRev, baseRev string
	localRev, err = Get("rev-parse", "@{0}")
	if err != nil {
		return
	}

	remoteRev, err = Get("rev-parse", remote)
	if err != nil {
		return
	}

	baseRev, err = Get("merge-base", "@{0}", remote)
	if err != nil {
		return
	}

	if localRev == remoteRev {
		return "=", nil
	}
	if localRev == baseRev {
		return "-1", nil
	}
	if remoteRev == baseRev {
		return "+1", nil
	}
	return "-1+1", nil
}

// RemoteExist return true if remote is defined.
func RemoteExist(remote string) (found bool) {
	var remotes []string
	v, err := Get("remote")
	if err != nil {
		return
	}

	remotes = strings.Split(v, "\n")

	for _, aRemote := range remotes {
		if aRemote == remote {
			return true
		}
	}
	return
}

// RemoteURL returns the url of the remote requested.
func RemoteURL(remote string) (string, bool, error) {
	var remotes []string
	v, err := Get("remote", "-v")
	if err != nil {
		return "", false, err
	}

	if v == "" {
		remotes = []string{}
	} else {
		remotes = strings.Split(v, "\n")
	}

	remMatch, _ := regexp.Compile(`^ *(\w+)[ \t]*(.*) \((fetch)\)$`)
	for _, aRemote := range remotes {
		if v := remMatch.FindStringSubmatch(aRemote); v != nil && v[1] == remote {
			return v[1], true, nil
		}
	}
	return "", false, nil
}

// EnsureRemoteIs will update the remote name with the url...
func EnsureRemoteIs(name, url string) error {
	if ru, found, err := RemoteURL(name); err != nil {
		return err
	} else if found {
		if ru != url {
			Do("remote", "set-url", name, url)
		}
	} else {
		Do("remote", "add", name, url)
	}
	return nil
}

// GetCurrentBranch return the current branch name.
// If no branch is detected, it returns "master"
func GetCurrentBranch() (branch string) {
	b, status := GetWithStatusCode("rev-parse", "--abbrev-ref", "HEAD")
	if status == 128 {
		return "master"
	}

	branch = b
	return
}

// EnsureRepoExist ensure a local repo exist.
func EnsureRepoExist(aPath string) error {
	if fi, err := os.Stat(path.Join(aPath, ".git")); err != nil && os.IsNotExist(err) {
		if Do("init", aPath) != 0 {
			return fmt.Errorf("Unable to create the local repository '%s'", aPath)
		}
	} else if err != nil {
		return err
	} else if !fi.IsDir() {
		return fmt.Errorf("'%s' is not a valid GIT repo (.git is not a directory)", aPath)
	}
	return nil
}

// RunInPath run a function in a specificDirectory and restore the current Path.
func RunInPath(gitRepoPath string, runIn func() error) error {

	restore, err := moveTo(gitRepoPath)
	if err != nil {
		return err
	}

	defer func() {
		UnIndent()
		os.Chdir(restore)
	}()

	Indent("---- GIT"+ShowGitPath(), " - ", "--------")

	if err = runIn(); err != nil {
		return err
	}
	return nil
}

func moveTo(gitPath string) (curDir string, err error) {
	if v, err := os.Getwd(); err != nil {
		return "", fmt.Errorf("Unable to get the current directory. %s", err)
	} else {
		curDir = v
	}
	gotrace.Trace("Moving to %s repo", gitPath)
	err = os.Chdir(gitPath)
	return
}
