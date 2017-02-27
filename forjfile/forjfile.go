package forjfile

import (
	"os"
	"path"
	"fmt"
	"io/ioutil"
	"github.com/forj-oss/forjj/.glide/cache/src/https-gopkg.in-yaml.v2"
)

type Forjfile struct {
	Forj ForjDefaultStruct
	Infra RepoStruct
	Workspace WorkspaceStruct `yaml:"local"`
	Repos map[string]RepoStruct
	Apps map[string]map[string]string
	Instances map[string]map[string]map[string]string `yaml:",inline"`
}

type ForjDefaultStruct struct {
	Organization string
	More map[string]string `yaml:",inline"`
}

type RepoStruct struct {
	Name string
	Upstream string
	More map[string]string `yaml:",inline"`
}

type WorkspaceStruct struct {
	ContribsPath string `yml:"contribs-path"`
	More map[string]string `yaml:",inline"`
}

type AppStruct struct {
	Name string
	Type string
	Driver string
	More map[string]string `yaml:",inline"`
}

const forj_file_name = "Forjfile"

// Load: Search for Forjfile in . and load it
func Load(aPath string) (f *Forjfile, loaded bool, err error) {

	forj_path := path.Clean(aPath)

	if forj_path != "." {
		if _, err = os.Stat(forj_path) ; err != nil {
			return
		}
	}

	// TODO: interpret ~ to $HOME and get it from path.Home()
	file := path.Join(forj_path, forj_file_name)
	if fi, e := os.Stat(file) ; e != nil {
		if forj_path != "." {
			err = e
		}
		return
	} else {
		if ! fi.Mode().IsRegular() {
			err = fmt.Errorf("%s must be a regular file.", file)
			return
		}
	}

	var yaml_data []byte
	if fd, e := ioutil.ReadFile(file) ; e != nil {
		err = e
		return
	} else {
		yaml_data = fd
	}

	f = new(Forjfile)

	if e := yaml.Unmarshal(yaml_data, f) ; e != nil {
		err = fmt.Errorf("Unable to load %s. %s", file, e)
		return
	}
	loaded = true
	return
}
