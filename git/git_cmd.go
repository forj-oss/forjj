package git

import (
	"fmt"
	"forjj/utils"
	"os/exec"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"
)

// Do Call git command with arguments. All print out displayed. It returns git Return code.
func Do(opts ...string) int {
	return defaultCmd.do(opts...)
}

// Get Call a git command and get the output as string output.
func Get(opts ...string) (string, error) {
	return defaultCmd.get(opts ...)
}

// GetWithStatusCode Call a git command and get the output as string output.
func GetWithStatusCode(opts ...string) (string, int) {
	return defaultCmd.getWithStatusCode(opts ...)
}

type gitCmd struct {}

var defaultCmd GitCmdInterface

// GetWithStatusCode Call a git command and get the output as string output.
func (c gitCmd) getWithStatusCode(opts ...string) (string, int) {
	colorCyan, colorReset := utils.DefColor(36)
	logFunc(fmt.Sprintf("%s%sgit %s%s\n", colorCyan, context.indent, strings.Join(opts, " "), colorReset))
	return utils.RunCmdOutput("git", opts...)
}

func (c gitCmd) get(opts ...string) (string, error) {
	gotrace.Trace("RUNNING: git %s", strings.Join(opts, " "))
	out, err := exec.Command("git", opts...).Output()
	return strings.Trim(string(out), "\n"), err
}

func (c gitCmd) do(opts ...string) int {
	colorCyan, colorReset := utils.DefColor(36)
	logFunc(fmt.Sprintf("%s%sgit %s%s\n", colorCyan, context.indent, strings.Join(opts, " "), colorReset))
	return utils.RunCmd("git", opts...)
}

func init() {
	defaultCmd = gitCmd{}
}