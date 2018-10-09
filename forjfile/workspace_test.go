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

	for _, field := range stdWsField {
		value, found := workspace.Get(field)

		test.Truef(found, "expect to be found %s %s", field, testCase)
		test.Emptyf(value, "expect to get an empty string for %s %s", field, testCase)
	}

	for _, field := range []string{"field1"} {
		value, found := workspace.Get(field)

		test.Falsef(found, "expect to not be found %s %s", field, testCase)
		test.Emptyf(value, "expect to get an empty string for %s %s", field, testCase)
	}

	/*********************************/
	testCase = "when workspace is empty but with a default value."

	workspace.SetDefault("contrib-repo-path", "blabla")

	value, found := workspace.Get("contrib-repo-path")
	test.Emptyf(value, "expect Get to be empty %s", testCase)
	test.Truef(found, "expect Get to be found %s", testCase)

	value, found = workspace.GetDefault("contrib-repo-path")
	test.Equalf("blabla", value, "expect GetDefault to return '%s' %s", "blabla", testCase)

	test.Equalf("blabla", workspace.GetString("contrib-repo-path"), "expect GetString to return '%s' %s", "blabla", testCase)

	test.Equalf(4, workspace.Len(), "expect Len to return '%d' %s", "4", testCase)

	/*********************************/
	testCase = "when workspace is nil and loading."

	workspace = nil
	err = workspace.Load()

	test.EqualErrorf(err, "Workspace is nil", "Expect an error %s", testCase)

	/*********************************/
	testCase = "when workspace is intialized and loading."
	workspace = new(Workspace)
	workspace.Init("infra-path")

	err = workspace.Load()

	test.EqualErrorf(err, "Invalid workspace. name or path are empty", "Expect an error %s", testCase)

	/*********************************/
	testCase = "when workspace is intialized and configured then loading."

	workspace.SetPath("/tmp/forjj/infra/workspace")
	test.Nilf(workspace.persistent.Infra, "Expect persistent Infra to be nil %s", testCase)
	test.NotNilf(workspace.internal.Infra, "Expect internal Infra to NOT be nil %s", testCase)

	err = workspace.Load()

	test.NotNilf(workspace.persistent.Infra, "Expect persistent Infra to NOT be nil %s", testCase)
	test.Truef(workspace.dirty, "Expect persistent to be identified as dirty %s", testCase)
	test.Equalf(workspace.persistent.Infra, workspace.internal.Infra, "Expect persistent infra to be same than internal object.")


	// Todo: Set, Data, Save and Load(70%)
}
