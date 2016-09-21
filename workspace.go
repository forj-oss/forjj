package main

import (
    "github.com/alecthomas/kingpin"
    "encoding/json"
    "log"
    "path"
    "io/ioutil"
    "os"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "github.hpe.com/christophe-larsonneur/goforjj"
    "fmt"
)

const forjj_workspace_json_file = "forjj.json"

// Define the workspace data saved at create/update time.
// Workspace data are not controlled by any git repo. It is local.
// Usually, we stored data to found out where the infra is.
// But it can store any data that is workspace environment specific.
// like where is the docker static binary.
type Workspace struct {
    Organization           string               // Workspace Organization name
    Driver                 string               // Infra upstream driver name
    Instance               string               // Infra upstream instance name
    Infra                  goforjj.PluginRepo   // Infra-repo definition
    DockerBinPath          string               // Docker static binary path
    Contrib_repo_path      string               // Contrib Repo path used.
    Flow_repo_path         string               // Flow repo path used.
    Repotemplate_repo_path string               // Repotemplate Path used.
    workspace              string               // Workspace name
    workspace_path         string               // Workspace directory path.
}

func (w *Workspace)Init(Workspace_path, Workspace string) {
    w.workspace_path = Workspace_path
    w.workspace = Workspace
    if w.Infra.Remotes == nil {
        w.Infra.Remotes = make(map[string]string)
    }
    if w.Infra.BranchConnect == nil {
        w.Infra.BranchConnect = make(map[string]string)
    }
    gotrace.Trace("Use workspace : %s (%s / %s)",w.Path(), Workspace_path, Workspace)
}

// Provide the workspace absolute path
func (w *Workspace)Path() (string) {
    return path.Clean(path.Join(w.workspace_path, w.workspace))
}

// Provide the workspace Name
func (w *Workspace)Name() (string) {
    return w.workspace
}


// Ensure workspace path exists. So, if missing, it will be created.
// The current path (pwd) is moved to the existing workspace path.
func (w *Workspace)Ensure_exist() (string, error){
    w_path := w.Path()
    _, err := os.Stat(w_path)
    if os.IsNotExist(err) {
        if err := os.MkdirAll(w_path, 0755) ; err != nil {
            return "", fmt.Errorf("Unable to create initial workspace tree '%s'. %s", w_path, err)
        }
    }
    os.Chdir(w_path)
    return w_path, nil
}

func (w *Workspace)Save(app *Forj) {
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
func (w *Workspace)Load(wsp, wsn string) error {
    w.Init(wsp, wsn)

    fjson := path.Join(w.Path(), forjj_workspace_json_file)

    _, err := os.Stat(fjson)
    if os.IsNotExist(err) { return nil }
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
