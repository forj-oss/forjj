package forjfile

import (
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
	cli        WorkspaceData // Data retrieved from cli
	dirty      bool          // True is persistent data has been updated
}

type WorkspaceExport struct {
	Value     string
	IsDefault bool
}

// Init initialize the Workspace object
func (w *Workspace) Init(cliSetup func(string) string, non_ws_entries ...string) {
	if w == nil {
		return
	}
	w.internal.Infra = goforjj.NewRepo()
	w.clean_entries = non_ws_entries
	w.cli.init(cliSetup, w.GetString)
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
	result[DockerBinPathField] = w.exportData(DockerBinPathField)
	result[ContribRepoPathField] = w.exportData(ContribRepoPathField)
	result[FlowRepoPathField] = w.exportData(FlowRepoPathField)
	result[RepoTemplateRepoPathField] = w.exportData(RepoTemplateRepoPathField)
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
func (w *Workspace) Set(field, value string, persistent bool) bool {
	return w.set(field, value, persistent, w.GetString)
}

func (w *Workspace) set(field, value string, persistent bool, getString func(string) string) (updated bool) {
	updated = w.internal.set(field, value, getString)
	if persistent {
		if w.persistent.set(field, value, getString) {
			w.dirty = true
		}
	}

	return
}

// SetDefault save field/value pair in the workspace as default value.
// This value is set to the internal if not set or if unset
func (w *Workspace) SetDefault(field, value string) {
	if _, found := w.internal.get(field); !found {
		w.internal.set(field, value, w.GetString)
	}
	w.def.set(field, value, w.GetString)
}

// Unset remove value of the given field in the workspace.
// The default value can be restored if it was originally set
func (w *Workspace) Unset(field string) (updated bool) {
	w.internal.set(field, "", w.GetString)
	if w.persistent.set(field, "", w.GetString) {
		w.dirty = true
		updated = true
	}
	if v, found := w.def.get(field); !found {
		w.internal.set(field, v, w.GetString)
	}
	return
}

// GetString return the data of the requested field.
// If not found, it return the default value
func (w *Workspace) GetString(field string) (value string) {
	var found bool
	if value, found = w.cli.get(field); found && value != "" {
		return
	}
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
	socketPath = w.GetString(PluginsSocketDirsPathField)
	if socketPath == "" {
		var err error
		socketPath, err = w.createNewSocketDir()
		kingpin.FatalIfError(err, "%s", err)
	} else {
		err := utils.EnsureDir(socketPath, "Socket directory")
		kingpin.FatalIfError(err, "%s", err)
		gotrace.Info("Using saved Socket Path: %s", socketPath)
	}
	return
}

// createNewSocketDir is called when the socket Dir was not created and saved in the workspace.
// It will create the base dir if this one is in /tmp
// Then it will create the Base Name directory under SockerDirName
func (w *Workspace) createNewSocketDir() (socketPath string, _ error) {
	baseDir := forjjSocketBaseDir
	status := "Using default Socket Path: %s"

	if w.internal.getString(PluginsSocketDirField) == "" {
		w.set(PluginsSocketDirField, baseDir, true, w.internal.getString)
	}
	if value := w.GetString(PluginsSocketDirField); value != "" {
		// Get SocketDirName from workspace or cli.
		baseDir = value
		status = "Using Socket Path: %s"
	}

	if err := utils.EnsureDir(baseDir, "Socket base directory"); err != nil {
		return "", err
	}

	socketPath, err := ioutil.TempDir(baseDir, "forjj-")
	if err != nil {
		return "", fmt.Errorf("Unable to create temporary dir in '%s'", "/tmp")
	}
	w.set(PluginsSocketBaseField, path.Base(socketPath), true, w.internal.getString) // Store default in workspace. ie SocketDir calculated from dir and base.
	w.Set(PluginsSocketBaseField, path.Base(socketPath), false)                      // but use the cli setup if defined internally. ie SocketDir calculated from dir/base defined by cli or workspace if set.
	gotrace.Info(status, socketPath)
	return
}

// Name Provide the workspace Name
func (w *Workspace) Name() string {
	if w == nil {
		return ""
	}

	return w.workspace
}

// EnsureExist Ensure workspace path exists. So, if missing, it will be created.
// The current path (pwd) is moved to the existing workspace path.
func (w *Workspace) EnsureExist() (wPath string, err error) {
	if w == nil {
		return "", fmt.Errorf("Workspace object is nil")
	}

	wPath = w.Path()
	if err = utils.EnsureDir(wPath, "initial workspace tree"); err != nil {
		return
	}
	os.Chdir(wPath)
	return
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
		w.dirty = true            // Force to save it as workspace file doesn't exist.
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

	if err := w.loadData(djson); err != nil {
		return fmt.Errorf("Unable to load '%s'. %s", fjson, err)
	}
	gotrace.Trace("Workspace data loaded from '%s'.", fjson)
	return nil
}

func (w *Workspace) loadData(djson []byte) error {
	if err := w.persistent.load(djson); err != nil {
		return err
	}

	if infra := w.persistent.Infra; infra == nil { // Ensure infra is already set even if not properly defined in the json file.
		w.persistent.Infra = w.internal.Infra
	}
	w.internal = w.persistent
	w.cli.computeSocketPath(w.GetString(PluginsSocketDirField), w.GetString(PluginsSocketBaseField))
	return nil
}
