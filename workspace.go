package main

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
	"io/ioutil"
	"log"
	"os"
	"path"
)

const forjj_workspace_json_file = "forjj.json"

// Define the workspace data saved at create/update time.
// Workspace data are not controlled by any git repo. It is local.
// Usually, we stored data to found out where the infra is.
// But it can store any data that is workspace environment specific.
// like where is the docker static binary.
type Workspace struct {
	Organization           string             // Workspace Organization name
	Driver                 string             // Infra upstream driver name
	Instance               string             // Infra upstream instance name
	Infra                  goforjj.PluginRepo // Infra-repo definition
	DockerBinPath          string             // Docker static binary path
	Contrib_repo_path      string             // Contrib Repo path used.
	Flow_repo_path         string             // Flow repo path used.
	Repotemplate_repo_path string             // Repotemplate Path used.
	workspace              string             // Workspace name
	workspace_path         string             // Workspace directory path.
	error                  error              // Error detected
}

func (w *Workspace) Init(Workspace_path string) {
	if Workspace_path == "" {
		return
	}
	Workspace_path, _ = Abs(path.Clean(Workspace_path))
	w.workspace_path = path.Dir(Workspace_path)
	w.workspace = path.Base(Workspace_path)
	if w.Infra.Remotes == nil {
		w.Infra.Remotes = make(map[string]string)
	}
	if w.Infra.BranchConnect == nil {
		w.Infra.BranchConnect = make(map[string]string)
	}
	gotrace.Trace("Use workspace : %s (%s / %s)", w.Path(), w.workspace_path, w.workspace)
}

// Path Provide the workspace absolute path
func (w *Workspace) Path() string {
	return path.Clean(path.Join(w.workspace_path, w.workspace))
}

// Name Provide the workspace Name
func (w *Workspace) Name() string {
	return w.workspace
}

// Ensure workspace path exists. So, if missing, it will be created.
// The current path (pwd) is moved to the existing workspace path.
func (w *Workspace) Ensure_exist() (string, error) {
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
func (w *Workspace) check_exist() (bool, error) {
	w_path := w.Path()
	_, err := os.Stat(w_path)
	if os.IsNotExist(err) {
		return false, fmt.Errorf("Forjj workspace tree '%s' is inexistent. %s", w_path, err)
	}
	return true, nil

}

func (w *Workspace) Save(app *Forj) {
	var djson []byte

	workspace_path, err := w.Ensure_exist()
	kingpin.FatalIfError(err, "Issue with '%s'", workspace_path)

	fjson := path.Join(workspace_path, forjj_workspace_json_file)

	djson, err = json.Marshal(w)
	kingpin.FatalIfError(err, "Issue to encode in json '%s'", djson)

	err = ioutil.WriteFile(fjson, djson, 0644)
	kingpin.FatalIfError(err, "Unable to create/update '%s'", fjson)

	gotrace.Trace("File '%s' saved with '%s'", fjson, djson)
}

// Load workspace information from the forjj.json
// Workspace path is get from forjj and set kept in the workspace as reference for whole forjj thanks to a.w.Path()
func (w *Workspace) Load() error {
	if w.workspace_path == "" || w.workspace == "" {
		return fmt.Errorf("Invalid workspace. name or path are empty.")
	}

	fjson := path.Join(w.Path(), forjj_workspace_json_file)

	_, err := os.Stat(fjson)
	if os.IsNotExist(err) {
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
		log.Fatal(err)
	}
	gotrace.Trace("File '%s' loaded.", fjson)
	return nil
}

// When this function is called, it will
// try to identify if we are in an existing workspace
// It will return the path found.
// You will need to call Init(path) and later Load()
func (w *Workspace) DetectIt() (string, error) {
	var pwd string

	gotrace.Trace("Detecting FORJJ workspace...")
	if v, err := os.Getwd(); err != nil {
		return "", err
	} else {
		pwd = v
	}
	for {
		if _, err := os.Stat(path.Join(pwd, forjj_workspace_json_file)); err == nil {
			gotrace.Trace("Found workspace at '%s'", pwd)
			return pwd, nil
		}
		pwd = path.Dir(pwd)
		if pwd == "/" {
			return "", fmt.Errorf("Unable to find a valid workspace from your path.")
		}
	}
}
