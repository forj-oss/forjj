package forjfile

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkspace(t *testing.T) {
	test := assert.New(t)

	/*********************************/
	var workspace *Workspace
	testCase := "when workspace is nil."

	test.NotPanicsf(func() {
		workspace.Init("infra-path")
	}, "Expect no panics %s", testCase)

	test.NotPanicsf(func() {
		err := workspace.SetPath("relative")
		test.EqualErrorf(err, "Workspace object nil.", "Expect an error %s", testCase)
	}, "Expect no panics %s", testCase)

	/*********************************/
	testCase = "when workspace is initialized."

	workspace = new(Workspace)
	workspace.Init("infra-path")

	entries := []string{"infra-path"}
	test.ElementsMatchf(entries, workspace.clean_entries, "Expect entries to be cleaned set %s", testCase)
	test.NotNilf(workspace.Infra, "Expect infra repository to be created empty by goforjj %s", testCase)

	/*********************************/
	testCase = "when workspace path is given empty."

	err := workspace.SetPath("")
	test.EqualErrorf(err, "Workspace path not defined.", "Expect an error %s", testCase)

	/*********************************/
	testCase = "when workspace path is given absolute."

	err = workspace.SetPath("/home/forjj/infra/.workspace")
	test.NoErrorf(err, "Expect no error %s", testCase)

	test.Equalf("/home/forjj/infra", workspace.workspace_path, "Expect workspace root path properly set %s", testCase)
	test.Equalf(".workspace", workspace.workspace, "Expect workspace relative path properly set %s", testCase)

	/*********************************/
	testCase = "when workspace path is given relative."

	err = workspace.SetPath("infra/.workspace")
	test.NoErrorf(err, "Expect no error %s", testCase)

	var curPath string

	curPath, err = os.Getwd()
	test.Equalf(path.Join(curPath, "infra"), workspace.workspace_path, "Expect workspace root path properly set %s", testCase)
	test.Equalf(".workspace", workspace.workspace, "Expect workspace relative path properly set %s", testCase)

	/*********************************/
	testCase = "when workspace is empty."

	for _, field := range []string{"docker-bin-path", "contrib-repo-path", "flow-repo-path", "repotemplate-repo-path", "field1"} {
		value, found := workspace.GetString(field)
		test.False(found, "expect to not found %s %s", field, testCase)
		test.Emptyf(value, "expect to get an empty string for %s %s", field, testCase)
	}

}
