package git

type gitCmdMock struct {
	stdout string
	stderr string
	combined string
	code int
	err error
}

// GetWithStatusCode Call a git command and get the output as string output.
func (c gitCmdMock)getWithStatusCode(opts ...string) (string, int) {
	return c.stdout, c.code
}

func (c gitCmdMock) get(opts ...string) (string, error) {
	return c.combined, c.err
}

func (c gitCmdMock) do(opts ...string) int {
	return c.code
}