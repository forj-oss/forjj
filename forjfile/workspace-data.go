package forjfile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/forj-oss/goforjj"
)

const (
	// DockerBinPathField identify full path to the docker binary.
	DockerBinPathField = "docker-bin-path"

	// ContribRepoPathField identify collection of paths of contribution repositories
	ContribRepoPathField = "contrib-repo-path"

	// FlowRepoPathField identify collection of path to the flow repositories.
	FlowRepoPathField = "flow-repo-path"

	// RepoTemplateRepoPathField identify collection of path to the repository template repositories.
	RepoTemplateRepoPathField = "repotemplate-repo-path"

	// PluginsSocketDirsPathField identify full path to the plugins sockets directory
	PluginsSocketDirsPathField = "plugins-socket-dirs-path"

	// PluginsSocketBaseField identify the base path of PluginsSocketDirsPathField
	PluginsSocketBaseField = "plugins-socket-base-path"

	// PluginsSocketDirField identify the dir name of PluginsSocketDirsPathField
	PluginsSocketDirField = "plugins-socket-dir-name"

	// OrganizationField identify the organization name
	OrganizationField = "organization"

	// InfraInstanceNameField identify the instance name
	InfraInstanceNameField = "infra-instance-name"

	//InfraDriverNameField identify the infra driver name
	InfraDriverNameField = "infra-driver-name"
)

var (
	stdWsField = [...]string{
		DockerBinPathField,
		ContribRepoPathField,
		FlowRepoPathField,
		RepoTemplateRepoPathField,
		PluginsSocketDirsPathField,
		PluginsSocketBaseField,
		PluginsSocketDirField,
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

// WorkspaceStruct represents the yaml structure of a workspace.
type WorkspaceStruct struct {
	updated                bool
	DockerBinPath          string            `yaml:"docker-exe-path"`          // Docker static binary path
	Contrib_repo_path      string            `yaml:"contribs-repo"`            // Contrib Repo path used.
	Flow_repo_path         string            `yaml:"flows-repo"`               // Flow repo path used.
	Repotemplate_repo_path string            `yaml:"repotemplates-repo"`       // Repotemplate Path used.
	SocketDir              string            `yaml:"-"`                        // Calculated field. Full path to forjj plugins sockets dir - Shared with plugins containers. Composed by `SocketBaseName`/`SocketDirName`.
	SocketDirName          string            `yaml:"-"`                        // Random dir name containing plugins sockets - Not stored in yaml
	SocketBaseName         string            `yaml:"socket-base-path"`         // Base path containing the sockets directory - Shared in DooD CI context - Not stored in yaml
	More                   map[string]string `yaml:",inline"`
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
	case PluginsSocketDirField:
		return w.SocketDirName
	case PluginsSocketBaseField:
		return w.SocketBaseName
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
func (w *WorkspaceData) init(initFunc func(string) string, getString func(string)string) {
	for _, field := range stdWsField {
		if value := initFunc(field); value != "" {
			w.set(field, value, getString)
		}
	}
}

// set update a field in the Workspace data
// SocketDir is computed field from the workspace.GetString data and value given thanks to computeSocketPath()
func (w *WorkspaceData) set(field, value string, getString func(string)string) (updated bool) {
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
	case PluginsSocketBaseField:
		updated = (w.SocketBaseName != value)
		w.SocketBaseName = value
		w.computeSocketPath(value,getString(PluginsSocketDirField) )
		return
	case PluginsSocketDirField:
		updated = (w.SocketDirName != value)
		w.SocketDirName = value
		w.computeSocketPath(getString(PluginsSocketBaseField), value)
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

// computeSocketPath build if possible the DockerDir field
// It is possible if both Base and Dir are set
//
func  (w *WorkspaceData) computeSocketPath(dir, basename string) {
	if basename != "" && dir != "" {
		w.SocketDir = path.Join(dir, basename)
	} else {
		w.SocketDir = ""
	}
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

// load
func (w *WorkspaceData) load(fjson []byte) error {
	if w == nil {
		return nil
	}

	if err := json.Unmarshal(fjson, &w); err != nil {
		return fmt.Errorf("Error loading json. %s", err)
	}

	w.expandPaths()
	return nil
}

// expandPaths is planned to be called after a yaml load to set SocketDirName and SocketBaseName from SocketDir
func (w *WorkspaceData) expandPaths() {
	if w == nil {
		return
	}
	if w.SocketDir != "" {
		w.SocketBaseName = path.Base(w.SocketDir)
		w.SocketDirName = path.Dir(w.SocketDir)
	}
}
