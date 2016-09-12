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
    Organization           string                         // Workspace Organization name
    Driver                 string                         // Infra upstream driver name
    Instance               string                         // Infra upstream instance name
    Infra                  goforjj.PluginRepo             // Infra-repo definition
    DockerBinPath          string                         // Docker static binary path
    Contrib_repo_path      string                         // Contrib Repo path used.
    Flow_repo_path         string                         // Flow repo path used.
    Repotemplate_repo_path string                         // Repotemplate Path used.
}

func (w *Workspace)Init() {
    if w.Infra.Remotes == nil {
        w.Infra.Remotes = make(map[string]string)
    }
    if w.Infra.BranchConnect == nil {
        w.Infra.BranchConnect = make(map[string]string)
    }
}

func (w *Workspace)Save(app *Forj) {
    var djson []byte

    workspace_path := path.Join(app.Workspace_path, app.Workspace)
    _, err := os.Stat(workspace_path)
    if os.IsNotExist(err) { return }
    kingpin.FatalIfError(err, "Issue with '%s'", workspace_path)

    fjson := path.Join(workspace_path, forjj_workspace_json_file)

    djson, err = json.Marshal(w)
    kingpin.FatalIfError(err, "Issue to encode in json '%s'", djson)

    err = ioutil.WriteFile(fjson, djson, 0644)
    kingpin.FatalIfError(err, "Unable to create/update '%s'", fjson)

    gotrace.Trace("File '%s' saved with '%s'", fjson, djson)
}

func (w *Workspace)Load(app *Forj) error {
    fjson := path.Join(app.Workspace_path, app.Workspace, forjj_workspace_json_file)

    w.Init()

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
