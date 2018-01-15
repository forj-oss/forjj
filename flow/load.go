package flow

import "net/url"

type Flows struct {
	all map[string]*FlowDefine
	paths []*url.URL
}

type FlowDefine struct { // Yaml structure
	Name   string
	Title  string // Flow title
	Define map[string]FlowPluginTypeDef
	Apps   map[string]FlowTasksDef
}

type FlowTasksDef []FlowTaskDef

type FlowPluginTypeDef struct {
	MaxInstances int `yaml:"max_instances"`
	Roles        []string
}

type FlowTaskDef struct {
	Commit string
	Task   []FlowTaskDo `yaml:"do"`
}

type FlowTaskDo struct {
	Api  FlowTaskAPIData
	Apps map[string]FlowTaskAPIData
}

type FlowTaskAPIData map[string]string

//
func (fs* Flows)Load(flows ...string) error {
	if fs.all == nil {
		fs.all = make(map[string]*FlowDefine)
	}

	for _, name := range flows {
		if f, err := fs.load_flow(name) ; err != nil {
			return err
		} else {
			fs.all[name] = f
		}
	}
	return nil
}

func (fs* Flows)load_flow(flow_name string) (flow *FlowDefine, err error) {
	return
}

func (fs* Flows)SetRepoPath(paths ...*url.URL) {
	if fs == nil {
		return
	}
	fs.paths = paths
}

func (fs* Flows)AddRepoPath(path_str *url.URL) (bool, error) {
	if fs == nil {
		return
	}
	if fs.paths == nil {
		fs.paths = make([]*url.URL, 1, 5)
		fs.paths[0] = path_str
		return true, nil
	}
	for _, value := range fs.paths {
		if value.String() == path_str.String() {
			return false, nil
		}
	}
	var paths []string
	if cap(fs.paths) <= len(fs.paths) {
		paths = make([]*url.URL, len(fs.paths), len(fs.paths) + 5)
	} else {
		paths = fs.paths
	}

	fs.paths = append(paths, path_str)
	return true, nil
}
