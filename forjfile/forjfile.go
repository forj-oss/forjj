package forjfile

import (
	"os"
	"path"
	"fmt"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"forjj/utils"
	"github.com/forj-oss/forjj-modules/trace"
)

type ForjfileTmpl struct {
	Workspace WorkspaceStruct `yaml:"local-settings"` // See workspace.go
	Forge `yaml:",inline"`
}

type Forge struct {
	updated bool
	updated_msg string
	Forj ForjStruct
	Infra RepoStruct
	Repos map[string]RepoStruct `yaml:"repositories"`
	Apps map[string]AppStruct `yaml:"applications"`
	Instances map[string]map[string]map[string]string `yaml:",inline"`
	infra_path string // Infra path used to create/save/load Forjfile
}

type WorkspaceStruct struct {
	updated bool
	DockerBinPath          string `yaml:"docker-exe-path"`    // Docker static binary path
	Contrib_repo_path      string `yaml:"contribs-repo"`      // Contrib Repo path used.
	Flow_repo_path         string `yaml:"flows-repo"`         // Flow repo path used.
	Repotemplate_repo_path string `yaml:"repotemplates-repo"` // Repotemplate Path used.
	More                   map[string]string `yaml:",inline"`
}

type ForjStruct struct {
	Settings ForjSettingsStruct
	Users map[string]UserStruct
	Groups map[string]GroupStruct
	More map[string]map[string]map[string]string `yaml:",inline"`
}

type ForjSettingsStruct struct {
	is_template bool
	Organization string
	ForjSettingsStructTmpl `yaml:",inline"`
}

type ForjSettingsStructTmpl struct {
	More map[string]string `yaml:",inline"`
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
	name string
	Type string
	Driver string
	More map[string]string `yaml:",inline"`
}

const forj_file_name = "Forjfile"

func (a *AppStruct) UnmarshalYAML(unmarchal func(interface{}) error) error {
	var app struct {
		name string
		Type string
		Driver string
		More map[string]string `yaml:",inline"`
	}

	if err := unmarchal(&app); err != nil {
		return err
	}
	if app.Type == "" {
		return fmt.Errorf("Application type value is required.")
	}

	*a = app
	return nil
}

func (a *AppStruct)Name() string {
	return a.name
}

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
	// Setting defaults
	for name, app := range f.Apps {
		app.name = name
		if app.Driver == "" {
			app.Driver = name
		}
		f.Apps[name] = app
	}
	loaded = true
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

func (f *Forge)SetFromTemplate(ft *ForjfileTmpl) {
	*f = ft.Forge
}

// Initialize the forge. (Forjfile in repository infra)
func (f *Forge)Init() {
	if f.Forj.Groups == nil {
		f.Forj.Groups = make(map[string]GroupStruct)
	}
	if f.Forj.Users == nil {
		f.Forj.Users = make(map[string]UserStruct)
	}
	if f.Forj.More == nil {
		f.Forj.More = make(map[string]map[string]map[string]string)
	}

	if f.Infra.More == nil {
		f.Infra.More = make(map[string]string)
	}

	if f.Repos == nil {
		f.Repos = make(map[string]RepoStruct)
	}

	if f.Apps == nil {
		f.Apps = make(map[string]AppStruct)
	}

	if f.Instances == nil {
		f.Instances = make(map[string]map[string]map[string]string)
	}
}

func (f *Forge) SetPath(infraPath string) error {
	aPath, err := utils.Abs(infraPath)
	if err != nil {
		return err
	}
	f.infra_path = aPath
	return nil
}

func (f *Forge) InfraPath() string {
	return f.infra_path
}

func (f *Forge) Save() error {
	return f.save(path.Join(f.infra_path, File_name))
}

func (f *Forge) save(file string) error {
	yaml_data, err := yaml.Marshal(f)
	if err != nil {
		return err
	}

	if f.infra_path != "" {
		if _, err = os.Stat(f.infra_path); err != nil {
			return nil
		}
	}

	if err := ioutil.WriteFile(file, yaml_data, 0644); err != nil {
		return err
	}
	gotrace.Trace("File name saved: %s", file)

	return nil
}

// SaveTmpl provide Forjfile template export from a Forge.
func SaveTmpl(aPath string, f *Forge) error {
	forge := new(Forge)
	*forge = *f
	forge.Forj.Settings.is_template = true
	return forge.save(aPath)
}

func (f *ForjSettingsStruct) MarshalYAML() (interface{}, error) {
	return f.ForjSettingsStructTmpl, nil
}
