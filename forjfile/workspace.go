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
const forjjSocketBaseDir = "/tmp/forjj"

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
	def        WorkspaceData // Default values
	persistent WorkspaceData // Data saved and loaded in forjj.json
	dirty      bool          // True is persistent data has been updated
}

type WorkspaceExport struct {
	Value     string
	IsDefault bool
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

// TODO: factorize the following function

// Data provides the list of workspace variables stored.
func (w *Workspace) Data() (result map[string]WorkspaceExport) {
	result = make(map[string]WorkspaceExport)
	result["docker-bin-path"] = w.exportData("docker-bin-path")
	result["contrib-repo-path"] = w.exportData("contrib-repo-path")
	result["flow-repo-path"] = w.exportData("flow-repo-path")
	result["repotemplate-repo-path"] = w.exportData("repotemplate-repo-path")
	for key := range w.internal.More {
		result[key] = w.exportData(key)
	}
	return
}

func (w *Workspace) exportData(field string) WorkspaceExport {
	value := w.GetString(field)
	defValue, _ := w.GetDefault(field)
	return WorkspaceExport{
		Value:     value,
		IsDefault: (value == defValue),
	}
}

// Len provides the numbers of workspace data stored.
func (w *Workspace) Len() int {
	return 4 + len(w.internal.More)
}

// Set save field/value pair in the workspace.
// If persistent is true, this data will be stored in the internal persistent workspace data
// Save will check this flag to update the .forj-workspace/forjj.json
func (w *Workspace) Set(field, value string, persistent bool) (updated bool) {
	updated = w.internal.set(field, value)
	if persistent {
		if w.persistent.set(field, value) {
			w.dirty = true
		}
	}

	return
}

// SetDefault save field/value pair in the workspace as default value.
// This value is set to the internal if not set or if unset
func (w *Workspace) SetDefault(field, value string) {
	if _, found := w.internal.get(field); !found {
		w.internal.set(field, value)
	}
	w.def.set(field, value)
}

// Unset remove value of the given field in the workspace.
// The default value can be restored if it was originally set
func (w *Workspace) Unset(field string) (updated bool) {
	w.internal.set(field, "")
	if w.persistent.set(field, "") {
		w.dirty = true
		updated = true
	}
	if v, found := w.def.get(field); !found {
		w.internal.set(field, v)
	}
	return
}

// GetString return the data of the requested field.
// If not found, it return the default value
func (w *Workspace) GetString(field string) (value string) {
	if value, _ = w.internal.get(field); value == "" {
		return w.def.getString(field)
	} else {
		return
	}
}

// Get return the value of the requested field and found if was found.
// Get do not extract the default value if not found
// to get the Default value, use GetDefault()
func (w *Workspace) Get(field string) (value string, found bool) {
	return w.internal.get(field)
}

// GetDefault return the default value of the requested field and found if was found.
func (w *Workspace) GetDefault(field string) (value string, found bool) {
	return w.def.get(field)
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
	w.dirty = true
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

// SocketPath creates a socket path if it doesn't exist.
// This information is stored in the workspace forjj.json file
func (w *Workspace) SocketPath() (socketPath string) {
	socketPath = w.GetString("plugins-socket-dirs-path")
	if socketPath == "" {
		var err error
		os.MkdirAll(forjjSocketBaseDir, 0755)
		socketPath, err = ioutil.TempDir(forjjSocketBaseDir, "forjj-")
		kingpin.FatalIfError(err, "Unable to create temporary dir in '%s'", "/tmp")
		w.Set("plugins-socket-dirs-path", socketPath, true)
	}
	return
}

// Name Provide the workspace Name
func (w *Workspace) Name() string {
	if w == nil {
		return ""
	}

	return w.workspace
}

// Ensure_exist Ensure workspace path exists. So, if missing, it will be created.
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

// checkDataExist create missing workspace path
//
func (w *Workspace) checkDataExist() (fjson string, found bool, err error) {
	if w == nil {
		return
	}

	wPath := w.Path()
	fjson = path.Join(wPath, forjj_workspace_json_file)

	_, err = os.Stat(wPath)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(wPath, 0755); err != nil {
			return
		}
	}

	_, err = os.Stat(fjson)
	found = !os.IsNotExist(err)
	err = nil
	return
}

// Check_exist Check if a workspace exist or not
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

// Save persistent workspace data to the json file
func (w *Workspace) Save() {
	if w == nil {
		return
	}
	fjson, exist, err := w.checkDataExist()
	kingpin.FatalIfError(err, "Issue with '%s'", fjson)

	w.CleanUnwantedEntries()

	if !exist || w.dirty {
		err = w.persistent.save(fjson)
	} else {
		gotrace.Trace("No Workspace updates: File '%s' not saved.'", fjson)
		return
	}

	gotrace.Trace("File '%s' saved.", fjson)
	w.dirty = false
}

// CleanUnwantedEntries is called before save to remove some unwanted data in the Workspace file.
// Ex: infra-path
func (w *Workspace) CleanUnwantedEntries() {
	for _, key := range w.clean_entries {
		if _, found := w.persistent.More[key]; found {
			delete(w.persistent.More, key)
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
		return fmt.Errorf("Workspace is nil")
	}
	if w.workspace_path == "" || w.workspace == "" {
		return fmt.Errorf("Invalid workspace. name or path are empty")
	}

	fjson := path.Join(w.Path(), forjj_workspace_json_file)

	_, err := os.Stat(fjson)
	if os.IsNotExist(err) {
		w.persistent = w.internal // Get default information initialized
		w.dirty = true // Force to save it as workspace file doesn't exist.
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

	if infra := w.persistent.Infra ; infra == nil { // Ensure infra is already set even if not properly defined in the json file.
		w.persistent.Infra = w.internal.Infra
	}
	w.internal = w.persistent
	gotrace.Trace("Workspace data loaded from '%s'.", fjson)
	return nil
}
