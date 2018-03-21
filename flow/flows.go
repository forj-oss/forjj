package flow

import (
	"fmt"
	"forjj/forjfile"
	"forjj/utils"
	"net/url"

	"github.com/forj-oss/forjj-modules/trace"
	"gopkg.in/yaml.v2"
)

type Flows struct {
	all   map[string]*FlowDefine
	paths []*url.URL
}

// Load flow the first flow file found.
func (fs *Flows) Load(flows ...string) error {
	if fs.all == nil {
		fs.all = make(map[string]*FlowDefine)
	}

	for _, name := range flows {
		if f, err := fs.loadFlow(name); err != nil {
			return err
		} else {
			fs.all[name] = f
			gotrace.Info("Flow definition '%s' loaded.", name)
		}
	}
	return nil
}

func (fs *Flows) loadFlow(flowName string) (flow *FlowDefine, _ error) {
	if data, err := utils.ReadDocumentFrom(fs.paths, ".yaml", flowName); err == nil {
		flow = new(FlowDefine)
		if err = yaml.Unmarshal(data, flow); err != nil {
			return nil, fmt.Errorf("Unable to load the flow '%s'. %s", flowName, err)
		}
		if flow.Name == "" {
			flow.Name = flowName
		}
	} else {
		return nil, fmt.Errorf("Unable to find '%s'. %s", flowName, err)
	}
	return flow, nil
}

// SetRepoPath set the collection of repositories in the flows object.
func (fs *Flows) SetRepoPath(paths ...*url.URL) {
	if fs == nil {
		return
	}
	fs.paths = paths
}

// AddRepoPath add a repo in the list of repository at the end of the list...
func (fs *Flows) AddRepoPath(pathStr *url.URL) (bool, error) {
	if fs == nil {
		return false, fmt.Errorf("Internal issue: %s", "url is nil.")
	}
	if fs.paths == nil {
		fs.paths = make([]*url.URL, 1, 5)
		fs.paths[0] = pathStr
		gotrace.Info("Flow path '%s' added.", pathStr.String())
		return true, nil
	}
	for _, value := range fs.paths {
		if value.String() == pathStr.String() {
			return false, nil
		}
	}
	var paths []*url.URL
	if cap(fs.paths) <= len(fs.paths) {
		paths = make([]*url.URL, len(fs.paths), len(fs.paths)+5)
	} else {
		paths = fs.paths
	}

	fs.paths = append(paths, pathStr)
	gotrace.Info("Flow path '%s' added.", pathStr.String())
	return true, nil
}

// Apply the flow to the Forjfile loaded.
func (fs *Flows) Apply(flowName string, repo *forjfile.RepoStruct, Forjfile *forjfile.Forge) error {
	var flow *FlowDefine

	if af, found := fs.all[flowName]; found {
		flow = af
	} else {
		return fmt.Errorf("Internal Error! Unable to find '%s' flow in memory", flowName)
	}
	return flow.apply(repo, Forjfile)
}
