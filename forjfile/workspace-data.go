package forjfile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/forj-oss/goforjj"
)

const (
	DockerBinPathField = "docker-bin-path"
	ContribRepoPathField = "contrib-repo-path"
	FlowRepoPathField = "flow-repo-path"
	RepoTemplateRepoPathField = "repotemplate-repo-path"
	PluginsSocketDirsPathField = "plugins-socket-dirs-path"
	OrganizationField = "organization"
	InfraInstanceNameField = "infra-instance-name"
	InfraDriverNameField = "infra-driver-name"
)

var (
	stdWsField = [...]string{
		DockerBinPathField,
		ContribRepoPathField,
		FlowRepoPathField,
		RepoTemplateRepoPathField,
		PluginsSocketDirsPathField,
		OrganizationField,
		InfraInstanceNameField,
		InfraDriverNameField,
	}
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
	case DockerBinPathField:
		return w.DockerBinPath
	case ContribRepoPathField:
		return w.Contrib_repo_path
	case FlowRepoPathField:
		return w.Flow_repo_path
	case RepoTemplateRepoPathField:
		return w.Repotemplate_repo_path
	case PluginsSocketDirsPathField:
		return w.SocketDir
	case OrganizationField:
		return w.Organization
	case InfraInstanceNameField:
		return w.Instance
	case InfraDriverNameField:
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

	for _, key := range stdWsField {
		if key == field {
			found = true
			break
		}
	}
	if !found {
		return
	}
	value = w.getString(field)
	return
}

// init initialize the workspace data from an initFunc function given
func (w *WorkspaceData) init(initFunc func(string)string) {
	for _, field := range stdWsField {
		if value := initFunc(field) ; value != "" {
			w.set(field, value)
		}
	}
}

// set update a field in the Workspace data
func (w *WorkspaceData) set(field, value string) (updated bool) {
	switch field {
	case DockerBinPathField:
		updated = (w.DockerBinPath != value)
		w.DockerBinPath = value
		return
	case ContribRepoPathField:
		updated = (w.Contrib_repo_path != value)
		w.Contrib_repo_path = value
		return
	case FlowRepoPathField:
		updated = (w.Flow_repo_path != value)
		w.Flow_repo_path = value
		return
	case RepoTemplateRepoPathField:
		updated = (w.Repotemplate_repo_path != value)
		w.Repotemplate_repo_path = value
		return
	case PluginsSocketDirsPathField:
		updated = (w.SocketDir != value)
		w.SocketDir = value
		return
	case OrganizationField:
		updated = (w.Organization != value)
		w.Organization = value
		return
	case InfraInstanceNameField:
		updated = (w.Instance != value)
		w.Instance = value
		return
	case InfraDriverNameField:
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
