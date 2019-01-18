package git

type GitCmdMock struct {
	stdout string
	stderr string
	combined string
	code int
	err error
}

// GetWithStatusCode Call a git command and get the output as string output.
func (c GitCmdMock)GetWithStatusCode(opts ...string) (string, int) {
	return c.stdout, c.code
}

func (c GitCmdMock) Get(opts ...string) (string, error) {
	return c.combined, c.err
}

func (c GitCmdMock) Do(opts ...string) int {
	return c.code
}