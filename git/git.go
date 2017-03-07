package git

import (
	"log"
	"strings"
	"regexp"
	"os/exec"
	"fmt"
	"github.com/forj-oss/forjj-modules/trace"
	"forjj/utils"
)

type GitStatus struct {
	Added     []string
	Modified  []string
	Untracked []string
	Err       error
}

// Git Call git command with arguments. All print out displayed. It returns git Return code.
func Do(opts ...string) int {
	colorCyan, colorReset := utils.DefColor(36)
	log.Printf("%sgit %s%s\n", colorCyan, strings.Join(opts, " "), colorReset)
	return utils.RunCmd("git", opts...)
}

func Status() (gs *GitStatus) {
	gs = new(GitStatus)
	gs.Added = make([]string, 0, 2)
	gs.Modified = make([]string, 0, 2)
	gs.Untracked = make([]string, 0, 2)
	added_re, _ := regexp.Compile("^[A-Z]  (.*)$")
	modified_re, _ := regexp.Compile("^ [A-Z] (.*)$")
	untracked_re, _ := regexp.Compile(`^\?\? (.*)$`)

	var s string

	s, gs.Err = Get("status", "--porcelain")
	if gs.Err != nil {
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
func Get(opts ...string) (string, error) {
	gotrace.Trace("RUNNING: git %s", strings.Join(opts, " "))
	out, err := exec.Command("git", opts...).Output()
	return string(out), err
}

// Commit Do a git commit
func Commit(msg string, errorIfEmpty bool) error {
	s := Status()
	if len(s.Added) == 0 {
		return fmt.Errorf("No files added. Please check.")
	}
	if Do("commit", "-m", msg) > 0 {
		return fmt.Errorf("Unable to commit.")
	}
	return nil
}

// GitPush Push latest commits
func GitPush() error {
	if Do("push") > 0 {
		return fmt.Errorf("Unable to push commits.")
	}
	return nil
}

