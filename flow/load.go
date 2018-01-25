package flow

import (
	"net/url"
	"path"
	"forjj/utils"
	"fmt"
	"strings"
	"gopkg.in/yaml.v2"
	"github.com/forj-oss/forjj-modules/trace"
	"forjj/forjfile"
)

type Flows struct {
	all map[string]*FlowDefine
	paths []*url.URL
}

type FlowDefine struct { // Yaml structure
	Name   string
	Title  string // Flow title
	Define map[string]FlowPluginTypeDef
	OnForjfile map[string]FlowTaskDef `yaml:"on-forjfile-do"`
}

type FlowPluginTypeDef struct {
	MaxInstances int `yaml:"max_instances"`
	Roles        []string
}

type FlowTaskDef struct {
	Description string
	OnList []FlowTaskList      `yaml:"on_list"`

	IfFound string             `yaml:"if_found"`
	IfEach FlowTaskListIfEach  `yaml:"if_each"`

	Set map[string]map[string]forjfile.ForjValues // key1: object, key2: instance, key3: value key, then value
}

type FlowTaskListIfEach struct {
	List []string
	Filter string
}

type FlowTaskList struct {
	Name string
	Object string
	FilteredBy []FlowListFilter `yaml:"filtered-by"`
	Extract []map[string]string
}

type FlowListFilter struct {
	Filter string
	Fields forjfile.ForjValues `yaml:",inline"`
}

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
			gotrace.Info("Flow definition '%s' loaded.", name)
		}
	}
	return nil
}

func (fs* Flows)load_flow(flow_name string) (*FlowDefine, error) {
	path_list := make([]string, len(fs.paths))
	for index, aUrl := range fs.paths {
		path_list[index] = aUrl.String()
		flow_url, _ := url.Parse(aUrl.String())
		flow_url.Path = path.Join(aUrl.Path, flow_name, flow_name + ".yaml")
		if data, err := utils.ReadDocumentFrom(flow_url) ; err == nil {
			flow := new(FlowDefine)
			if err = yaml.Unmarshal(data, flow) ; err != nil {
				return nil, fmt.Errorf("Unable to load the flow '%s'. %s", flow_url.String(), err)
			}
			return flow, nil
		}
	}
	return nil, fmt.Errorf("Unable to find '%s' from '%s'", flow_name, strings.Join(path_list, "', '"))
}

func (fs* Flows)SetRepoPath(paths ...*url.URL) {
	if fs == nil {
		return
	}
	fs.paths = paths
}

func (fs* Flows)AddRepoPath(path_str *url.URL) (bool, error) {
	if fs == nil {
		return false, fmt.Errorf("Internal issue: %s", "url is nil.")
	}
	if fs.paths == nil {
		fs.paths = make([]*url.URL, 1, 5)
		fs.paths[0] = path_str
		gotrace.Info("Flow path '%s' added.", path_str.String())
		return true, nil
	}
	for _, value := range fs.paths {
		if value.String() == path_str.String() {
			return false, nil
		}
	}
	var paths []*url.URL
	if cap(fs.paths) <= len(fs.paths) {
		paths = make([]*url.URL, len(fs.paths), len(fs.paths) + 5)
	} else {
		paths = fs.paths
	}

	fs.paths = append(paths, path_str)
	gotrace.Info("Flow path '%s' added.", path_str.String())
	return true, nil
}
