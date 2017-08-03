package forjfile

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
	"io/ioutil"
	"os"
	"path"
	"forjj/utils"
)

const forjj_workspace_json_file = "forjj.json"

// Define the workspace data saved at create/update time.
// Workspace data are not controlled by any git repo. It is local.
// Usually, we stored data to found out where the infra is.
// But it can store any data that is workspace environment specific.
// like where is the docker static binary.
type Workspace struct {
	Organization           string              // Workspace Organization name
	Driver                 string              // Infra upstream driver name
	Instance               string              // Infra upstream instance name
	Infra                  *goforjj.PluginRepo // Infra-repo definition
	workspace              string              // Workspace name
	workspace_path         string              // Workspace directory path.
	error                  error               // Error detected
	is_workspace           bool                // True if instance is the workspace data to save in Workspace path.
	clean_entries          []string            // List of keys to ensure removed.
	WorkspaceStruct
}

/*func (w *WorkspaceStruct)MarshalYAML() (interface{}, error) {

}*/

func (w *Workspace)Init(non_ws_entries ...string) {
	if w == nil {
		return
	}
	w.Infra = goforjj.NewRepo()
	w.clean_entries = non_ws_entries
}

func (w *Workspace) SetPath(Workspace_path string) error {
	if w == nil {
		return fmt.Errorf("Workspace object nil.")
	}
	if Workspace_path == "" {
		return fmt.Errorf("Workspace path not defined.")
	}
	Workspace_path, _ = utils.Abs(path.Clean(Workspace_path))
	w.workspace_path = path.Dir(Workspace_path)
	w.workspace = path.Base(Workspace_path)
	gotrace.Trace("Use workspace : %s (%s / %s)", w.Path(), w.workspace_path, w.workspace)
	return nil
}

func (w *Workspace) GetString(field string) (value string, found bool) {
	switch field {
	case "docker-bin-path":
		return w.DockerBinPath, (w.DockerBinPath != "")
	case "contrib-repo-path":
		return w.Contrib_repo_path, (w.Contrib_repo_path != "")
	case "flow-repo-path":
		return w.Flow_repo_path, (w.Flow_repo_path != "")
	case "repotemplate-repo-path":
		return w.Repotemplate_repo_path, (w.Repotemplate_repo_path != "")
	}
	value, found = w.More[field]
	return
}

func (w *Workspace) RequireWorkspacePath() error {
	if w.workspace == "" {
		return fmt.Errorf("Workspace path not defined.")
	}
	aPath := w.Path()
	if _, err := os.Stat(aPath) ; err != nil {
		if err = os.Mkdir(aPath, 0755) ; err != nil {
			return fmt.Errorf("Unable to create Workspace path '%s'. %s", aPath, err)
		}
		gotrace.Trace("Workspace path '%s' has been created.", aPath)
		return nil
	}
	gotrace.Trace("Workspace path '%s' has been re-used.", aPath)
	return nil
}

func (w *Workspace) SetFrom(aWorkspace WorkspaceStruct) {
	if w == nil {
		return
	}
	w.WorkspaceStruct = aWorkspace
}

// InfraPath Return the path which contains the workspace.
// As the workspace is in the root or the infra repository, that
// path is then the Infra path.
// Note: The infra name is the repository name, ie the upstream
// repo name. This name is not necessarily the base name of the
// Infra path, because we can clone to a different name.
func (w *Workspace) InfraPath() string {
	if w == nil {
		return ""
	}
	return w.workspace_path
}

// Path Provide the workspace absolute path
func (w *Workspace) Path() string {
	if w == nil {
		return ""
	}

	return path.Clean(path.Join(w.workspace_path, w.workspace))
}

// Name Provide the workspace Name
func (w *Workspace) Name() string {
	if w == nil {
		return ""
	}

	return w.workspace
}

// Ensure workspace path exists. So, if missing, it will be created.
// The current path (pwd) is moved to the existing workspace path.
func (w *Workspace) Ensure_exist() (string, error) {
	if w == nil {
		return "", fmt.Errorf("Workspace is nil.")
	}

	w_path := w.Path()
	_, err := os.Stat(w_path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(w_path, 0755); err != nil {
			return "", fmt.Errorf("Unable to create initial workspace tree '%s'. %s", w_path, err)
		}
	}
	os.Chdir(w_path)
	return w_path, nil
}

// Check if a workspace exist or not
func (w *Workspace) Check_exist() (bool, error) {
	if w == nil {
		return false, fmt.Errorf("Workspace is nil.")
	}
	w_path := w.Path()
	_, err := os.Stat(w_path)
	if os.IsNotExist(err) {
		return false, fmt.Errorf("Forjj workspace tree '%s' is inexistent. %s", w_path, err)
	}
	return true, nil

}

func (w *Workspace) Save() {
	if w == nil {
		return
	}
	var djson []byte

	workspace_path, err := w.Ensure_exist()
	kingpin.FatalIfError(err, "Issue with '%s'", workspace_path)

	fjson := path.Join(workspace_path, forjj_workspace_json_file)

	w.CleanUnwantedEntries()

	djson, err = json.Marshal(w)
	kingpin.FatalIfError(err, "Issue to encode in json '%s'", djson)

	err = ioutil.WriteFile(fjson, djson, 0644)
	kingpin.FatalIfError(err, "Unable to create/update '%s'", fjson)

	gotrace.Trace("File '%s' saved with '%s'", fjson, djson)
}

// CleanUnwantedEntries is called before save to remove some unwanted data in the Workspace file.
// Ex: infra-path
func (w *Workspace) CleanUnwantedEntries() {
	for _, key := range w.clean_entries {
		if _, found := w.More[key] ; found {
			delete(w.More, key)
		}
	}
}

func (w *Workspace) Error() error {
	if w == nil {
		return fmt.Errorf("Workspace is nil.")
	}
	return w.error
}

func (w *Workspace) SetError(err error) error{
	if w == nil {
		return fmt.Errorf("Workspace is nil.")
	}
	w.error = err
	return w.error
}

// Load workspace information from the forjj.json
// Workspace path is get from forjj and set kept in the workspace as reference for whole forjj thanks to a.w.Path()
func (w *Workspace) Load() error {
	if w == nil {
		return fmt.Errorf("Workspace is nil.")
	}
	if w.workspace_path == "" || w.workspace == "" {
		return fmt.Errorf("Invalid workspace. name or path are empty.")
	}

	fjson := path.Join(w.Path(), forjj_workspace_json_file)

	_, err := os.Stat(fjson)
	if os.IsNotExist(err) {
		gotrace.Trace("'%s' not found. Workspace data not loaded.", fjson)
		return nil
	}
	if err != nil {
		return fmt.Errorf("Issue to access '%s'. %s", fjson, err)
	}

	var djson []byte
	djson, err = ioutil.ReadFile(fjson)
	if err != nil {
		return fmt.Errorf("Unable to read '%s'. %s", fjson, err)
	}

	if err := json.Unmarshal(djson, &w); err != nil {
		return fmt.Errorf("Unable to load '%s'. %s", fjson, err)
	}
	gotrace.Trace("Workspace data loaded from '%s'.", fjson)
	return nil
}
