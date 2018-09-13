package forjfile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/forj-oss/goforjj"
)

// WorkspaceData contains the structured data saved as json
type WorkspaceData struct {
	Organization    string              // Workspace Organization name
	Driver          string              // Infra upstream driver name
	Instance        string              // Infra upstream instance name
	Infra           *goforjj.PluginRepo // Infra-repo definition
	WorkspaceStruct                     // Struct shared with local-settings of a Forjfile model
}

// getString return the data of the requested field.
func (w *WorkspaceData) getString(field string) (value string) {
	switch field {
	case "docker-bin-path":
		return w.DockerBinPath
	case "contrib-repo-path":
		return w.Contrib_repo_path
	case "flow-repo-path":
		return w.Flow_repo_path
	case "repotemplate-repo-path":
		return w.Repotemplate_repo_path
	case "plugins-socket-dirs-path":
		return w.SocketDir
	case "organization":
		return w.Organization
	case "infra-instance-name":
		return w.Instance
	case "infra-driver-name":
		return w.Driver
	}
	value, _ = w.More[field]
	return
}

// get return the value of the requested field and found if was found.
func (w *WorkspaceData) get(field string) (value string, found bool) {
	if value, found = w.More[field]; found {
		return
	}
	value = w.getString(field)
	found = (value != "")
	return
}

// set update a field in the Workspace data
func (w *WorkspaceData) set(field, value string) (updated bool) {
	switch field {
	case "docker-bin-path":
		updated = (w.DockerBinPath != value)
		w.DockerBinPath = value
		return
	case "contrib-repo-path":
		updated = (w.Contrib_repo_path != value)
		w.Contrib_repo_path = value
		return
	case "flow-repo-path":
		updated = (w.Flow_repo_path != value)
		w.Flow_repo_path = value
		return
	case "repotemplate-repo-path":
		updated = (w.Repotemplate_repo_path != value)
		w.Repotemplate_repo_path = value
		return
	case "plugins-socket-dirs-path":
		updated = (w.SocketDir != value)
		w.SocketDir = value
		return
	case "organization":
		updated = (w.Organization != value)
		w.Organization = value
		return
	case "infra-instance-name":
		updated = (w.Instance != value)
		w.Instance = value
		return
	case "infra-driver-name":
		updated = (w.Driver != value)
		w.Driver = value
		return
	}
	if v, found := w.More[field]; found {
		updated = (v != value)
	} else {
		updated = true
	}
	if w.More == nil {
		w.More = make(map[string]string)
	}
	w.More[field] = value
	return

}

// save
func (w *WorkspaceData) save(fjson string) error {
	if w == nil {
		return nil
	}

	djson, err := json.MarshalIndent(w, "", " ")
	if err != nil {
		return fmt.Errorf("Issue to encode in json. %s", err)
	}

	err = ioutil.WriteFile(fjson, djson, 0644)
	if err != nil {
		return fmt.Errorf("Unable to create/update '%s'. %s", fjson, err)
	}
	return nil
}
