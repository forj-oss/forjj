package forjfile

import (
	"os"
	"path"
	"fmt"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type ForjfileTmpl struct {
	Workspace WorkspaceStruct `yaml:"local-settings"` // See workspace.go
	Forge `yaml:",inline"`
}

type Forge struct {
	Forj ForjDefaultStruct
	Infra RepoStruct
	Repos map[string]RepoStruct `yaml:"repositories"`
	Apps map[string]AppStruct `yaml:"applications"`
	Instances map[string]map[string]map[string]string `yaml:",inline"`
}

type WorkspaceStruct struct {
	DockerBinPath          string `yml:"docker-exe-path"`    // Docker static binary path
	Contrib_repo_path      string `yml:"contribs-path"`      // Contrib Repo path used.
	Flow_repo_path         string `yml:"flows-path"`         // Flow repo path used.
	Repotemplate_repo_path string `yml:"repotemplates-path"` // Repotemplate Path used.
	More                   map[string]string `yaml:",inline"`
}

type ForjDefaultStruct struct {
	Organization string
	Users map[string]UserStruct
	Groups map[string]GroupStruct
	More map[string]map[string]map[string]string `yaml:",inline"`
}

type UserStruct struct {
	Email string
}

type GroupStruct struct {
	Members []string
}

type RepoStruct struct {
	Name string
	Upstream string
	More map[string]string `yaml:",inline"`
}

type AppStruct struct {
	Name string
	Type string
	Driver string
	More map[string]string `yaml:",inline"`
}

const forj_file_name = "Forjfile"

// TODO: Load multiple templates that will be merged.

// LoadTmpl: Search for Forjfile in `aPath` and load it.
// This file combines the Forjfile in the infra repository and the Workspace
func LoadTmpl(aPath string) (f *ForjfileTmpl, loaded bool, err error) {
	var (
		yaml_data []byte
		file string
	)

	if fi, d, e := loadFile(aPath) ; e != nil {
		err = e
		return
	} else {
		yaml_data = d
		file = fi
	}

	f = new(ForjfileTmpl)

	if e := yaml.Unmarshal(yaml_data, f) ; e != nil {
		err = fmt.Errorf("Unable to load %s. %s", file, e)
		return
	}
	loaded = true
	fmt.Printf("%#v\n", f)
	return
}

// Load : Load Forjfile stored in a Repository.
func Load(aPath string) (f *Forge, loaded bool, err error) {
	var (
		yaml_data []byte
		file string
	)

	if fi, d, e := loadFile(aPath) ; e != nil {
		err = e
		return
	} else {
		yaml_data = d
		file = fi
	}

	f = new(Forge)

	if e := yaml.Unmarshal(yaml_data, f) ; e != nil {
		err = fmt.Errorf("Unable to load %s. %s", file, e)
		return
	}
	loaded = true
	fmt.Printf("%#v\n", f)
	return
}

func loadFile(aPath string) (file string, yaml_data[]byte, err error) {
		forj_path := path.Clean(aPath)

	if forj_path != "." {
		if _, err = os.Stat(forj_path) ; err != nil {
			return
		}
	}

	// TODO: interpret ~ to $HOME and get it from path.Home()
	file = path.Join(forj_path, forj_file_name)
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

	if fd, e := ioutil.ReadFile(file) ; e != nil {
		err = e
		return
	} else {
		yaml_data = fd
	}
	return
}
