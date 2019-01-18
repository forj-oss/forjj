package git

type GitCmd struct {
	stdout string
	stderr string
	combined string
	code int
	err error
}

var defa

// GetWithStatusCode Call a git command and get the output as string output.
func (c GitCmd)GetWithStatusCode(opts ...string) (string, int) {
}

func (c GitCmdMock) Get(opts ...string) (string, error) {
}

func (c GitCmdMock) Do(opts ...string) int {
	colorCyan, colorReset := utils.DefColor(36)
	logFunc(fmt.Sprintf("%s%sgit %s%s\n", colorCyan, context.indent, strings.Join(opts, " "), colorReset))
	return utils.RunCmd("git", opts...)
}