package git

import (
	"testing"
	"errors"

	"github.com/stretchr/testify/assert"
)

func TestgitCmdInit(t *testing.T) {
	t.Log("Expect Cmd Interface to be set by default.")

	assert := assert.New(t)

	assert.NotNil(defaultCmd)
}

func TestgitCmdFunctions(t *testing.T) {
	t.Log("Expect Cmd function can be set for tests.")

	assert := assert.New(t)

	cmdMock := gitCmdMock{
		code:2,
		stdout: "output test\n",
		stderr: "stderr\n",
		combined: "output test\nstderr\n",
		err: errors.New("error"),
	}
	defaultCmd = cmdMock
	
	code := defaultCmd.do("blabla")
	assert.Equalf(cmdMock.code, code, "expected to get code '%s'", cmdMock.code)

	output, err := defaultCmd.get("blabla")
	assert.Error(err, "expected an error")
	assert.Equal(cmdMock.stdout, output , "expected an error")

	output, code = defaultCmd.getWithStatusCode("blabla")
	assert.Equal(cmdMock.stdout, output , "expected an error")
	assert.Equalf(cmdMock.code, code, "expected to get code '%s'", cmdMock.code)
}