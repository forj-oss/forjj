package forjfile

import (
	"encoding/json"
	"fmt"
	"forjj/utils"
	"io/ioutil"
	"os"
	"path"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

const forjj_workspace_json_file = "forjj.json"

// Define the workspace data saved at create/update time.
// Workspace data are not controlled by any git repo. It is local.
// Usually, we stored data to found out where the infra is.
// But it can store any data that is workspace environment specific.
// like where is the docker static binary.
type Workspace struct {
	workspace      string   // Workspace name
	workspace_path string   // Workspace directory path.
	error          error    // Error detected
	is_workspace   bool     // True if instance is the workspace data to save in Workspace path.
	clean_entries  []string // List of keys to ensure removed.

	internal   WorkspaceData
	persistent WorkspaceData // Data saved and loaded in forjj.json
}

// WorkspaceData contains the structured data saved as json
type WorkspaceData struct {
	Organization    string              // Workspace Organization name
	Driver          string              // Infra upstream driver name
	Instance        string              // Infra upstream instance name
	Infra           *goforjj.PluginRepo // Infra-repo definition
	WorkspaceStruct                     // Struct shared with local-settings of a Forjfile model
}

// Init initialize the Workspace object
func (w *Workspace) Init(non_ws_entries ...string) {
	if w == nil {
		return
	}
	w.internal.Infra = goforjj.NewRepo()
	w.clean_entries = non_ws_entries
}

// SetPath define the workspace path.
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

// GetString return the data of the requested field.
func (w *Workspace) GetString(field string) (value string) {
	switch field {
	case "docker-bin-path":
		return w.internal.DockerBinPath
	case "contrib-repo-path":
		return w.internal.Contrib_repo_path
	case "flow-repo-path":
		return w.internal.Flow_repo_path
	case "repotemplate-repo-path":
		return w.internal.Repotemplate_repo_path
	case "plugins-socket-dirs-path":
		return w.internal.SocketDir
	case "organization":
		return w.internal.Organization
	case "infra-instance-name":
		return w.internal.Instance
	case "infra-driver-name":
		return w.internal.Driver
	}
	value, _ = w.internal.More[field]
	return
}

// Get return the value of the requested field and found if was found.
func (w *Workspace) Get(field string) (value string, found bool) {
	if value, found = w.internal.More[field]; found {
		return
	}
	value = w.GetString(field)
	found = (value != "")
	return
}

func (w *Workspace) Set(field, value string, persistent bool) (updated bool) {
	switch field {
	case "docker-bin-path":
		updated = (w.internal.DockerBinPath != value)
		w.internal.DockerBinPath = value
		return
	case "contrib-repo-path":
		updated = (w.internal.Contrib_repo_path != value)
		w.internal.Contrib_repo_path = value
		return
	case "flow-repo-path":
		updated = (w.internal.Flow_repo_path != value)
		w.internal.Flow_repo_path = value
		return
	case "repotemplate-repo-path":
		updated = (w.internal.Repotemplate_repo_path != value)
		w.internal.Repotemplate_repo_path = value
		return
	case "plugins-socket-dirs-path":
		updated = (w.internal.SocketDir != value)
		w.internal.SocketDir = value
		return
	case "organization":
		updated = (w.internal.Organization != value)
		w.internal.Organization = value
		return
	case "infra-instance-name":
		updated = (w.internal.Instance != value)
		w.internal.Instance = value
		return
	case "infra-driver-name":
		updated = (w.internal.Driver != value)
		w.internal.Driver = value
		return
	}
	if v, found := w.internal.More[field] ; found {
		updated = (v != value)
	} else {
		updated = true
	}
	w.internal.More[field] = value
	return

}

// Infra return the Infra data object
func (w *Workspace) Infra() (ret *goforjj.PluginRepo) {
	return w.internal.Infra
}

// SetInfra save the infra object in the workspace internal data
func (w *Workspace) SetInfra(infra *goforjj.PluginRepo) {
	w.internal.Infra = infra
}

func (w *Workspace) RequireWorkspacePath() error {
	if w.workspace == "" {
		return fmt.Errorf("Workspace path not defined.")
	}
	aPath := w.Path()
	if _, err := os.Stat(aPath); err != nil {
		if err = os.Mkdir(aPath, 0755); err != nil {
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
	w.internal.WorkspaceStruct = aWorkspace
	w.persistent.WorkspaceStruct = aWorkspace
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

	djson, err = json.Marshal(w.persistent)
	kingpin.FatalIfError(err, "Issue to encode in json '%s'", djson)

	err = ioutil.WriteFile(fjson, djson, 0644)
	kingpin.FatalIfError(err, "Unable to create/update '%s'", fjson)

	gotrace.Trace("File '%s' saved with '%s'", fjson, djson)
}

// CleanUnwantedEntries is called before save to remove some unwanted data in the Workspace file.
// Ex: infra-path
func (w *Workspace) CleanUnwantedEntries() {
	for _, key := range w.clean_entries {
		if _, found := w.internal.More[key]; found {
			delete(w.internal.More, key)
		}
	}
}

func (w *Workspace) Error() error {
	if w == nil {
		return fmt.Errorf("Workspace is nil.")
	}
	return w.error
}

func (w *Workspace) SetError(err error) error {
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

	if err := json.Unmarshal(djson, &w.persistent); err != nil {
		return fmt.Errorf("Unable to load '%s'. %s", fjson, err)
	}
	w.internal = w.persistent
	gotrace.Trace("Workspace data loaded from '%s'.", fjson)
	return nil
}
