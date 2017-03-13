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

// ForjfileTmpl is the Memory expansion of an external Forjfile (used to create a Forge)
type ForjfileTmpl struct {
	Workspace WorkspaceStruct `yaml:"local-settings"` // See workspace.go
	ForgeYaml `yaml:",inline"`
}

// Forge is the Memory expand of a repository Forjfile.
type Forge struct {
	updated_msg string
	infra_path string // Infra path used to create/save/load Forjfile
	file_name string // Relative path to the Forjfile.
	yaml *ForgeYaml
}

type ForgeYaml struct {
	updated bool
	ForjSettings ForjSettingsStruct `yaml:"forj-settings"`
	Infra InfraRepoStruct
	Repos map[string]RepoStruct `yaml:"repositories"`
	Apps map[string]AppStruct `yaml:"applications"`
	Users map[string]UserStruct
	Groups map[string]GroupStruct
	// Collection of Object/Name/Keys=values
	More map[string]map[string]map[string]string `yaml:",inline"`

}

type InfraRepoStruct struct {
	Name string
	RepoStruct `yaml:",inline"`
}

type WorkspaceStruct struct {
	updated bool
	DockerBinPath          string `yaml:"docker-exe-path"`    // Docker static binary path
	Contrib_repo_path      string `yaml:"contribs-repo"`      // Contrib Repo path used.
	Flow_repo_path         string `yaml:"flows-repo"`         // Flow repo path used.
	Repotemplate_repo_path string `yaml:"repotemplates-repo"` // Repotemplate Path used.
	More                   map[string]string `yaml:",inline"`
}

const file_name = "Forjfile"

// TODO: Load multiple templates that will be merged.

// LoadTmpl: Search for Forjfile in `aPath` and load it.
// This file combines the Forjfile in the infra repository and the Workspace
func LoadTmpl(aPath string) (f *ForjfileTmpl, loaded bool, err error) {
	var (
		yaml_data []byte
	)

	forj_path := path.Clean(aPath)
	if  forj_path != "." {
		if _, err = os.Stat(forj_path) ; err != nil {
			return
		}
	}

	// TODO: interpret ~ to $HOME and get it from path.Home()
	file := path.Join(forj_path, file_name)

	if fi, d, e := loadFile(file) ; e != nil {
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
	gotrace.Trace("Forjfile template '%s' has been loaded.", file)
	f.ForgeYaml.set_defaults()
	loaded = true
	return
}

// Load : Load Forjfile stored in a Repository.
func (f *Forge)Load() (loaded bool, err error) {
	var (
		yaml_data []byte
		file string
	)

	if f.infra_path != "" {
		if _, err = os.Stat(f.infra_path); err != nil {
			return
		}
	}

	aPath := path.Join(f.infra_path, f.Forjfile_name())
	if fi, d, e := loadFile(aPath) ; e != nil {
		err = e
		return
	} else {
		yaml_data = d
		file = fi
	}

	if f.yaml == nil {
		f.Init()
	}
	if e := yaml.Unmarshal(yaml_data, f.yaml) ; e != nil {
		err = fmt.Errorf("Unable to load %s. %s", file, e)
		return
	}

	f.yaml.set_defaults()
	loaded = true
	gotrace.Trace("Forge loaded from '%s'.", aPath)
	return
}


func loadFile(aPath string) (file string, yaml_data[]byte, err error) {
	file = path.Clean(aPath)

	// TODO: interpret ~ to $HOME and get it from path.Home()
	if fi, e := os.Stat(file) ; e == nil {
		if ! fi.Mode().IsRegular() {
			err = fmt.Errorf("%s must be a regular file.", file)
			return
		}
	} else {
		err = e
		return
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
	if f == nil {
		return
	}
	f.Init()
	*f.yaml = ft.ForgeYaml
}

func (f *Forge)Init() {
	if f == nil {
		return
	}
	if f.yaml == nil {
		f.yaml = new(ForgeYaml)
	}
}

func (f *Forge) SetInfraPath(infraPath string) error {
	aPath, err := utils.Abs(infraPath)
	if err != nil {
		return err
	}
	f.infra_path = aPath
	f.file_name = file_name // By default on Repo root directory.
	return nil
}

func (f *Forge) SetRelPath(relPath string) {
	f.file_name = path.Clean(path.Join(relPath, f.file_name))
}

func (f *Forge) InfraPath() string {
	return f.infra_path
}

func (f *Forge) Forjfile_name() string {
	return f.file_name
}

func (f *Forge) Save() error {
	if err := f.save(path.Join(f.infra_path, f.Forjfile_name())) ; err != nil {
		return err
	}
	f.Saved()
	return nil
}

func (f *Forge) save(file string) error {
	f.Init()
	yaml_data, err := yaml.Marshal(f.yaml)
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
	forge.yaml.ForjSettings.is_template = true
	return forge.save(aPath)
}

func (f *Forge) Get(object, instance, key string) (string, bool) {
	f.Init()
	switch object {
	case "infra":
		if key == "name" && f.yaml.Infra.Name != "" {
			return f.yaml.Infra.Name, true
		}
		return f.yaml.Infra.Get(key)
	case "user":
		if f.yaml.Users == nil {
			return "", false
		}
		if user, found := f.yaml.Users[instance] ; found {
			return user.Get(key)
		}
	case "group":
		if f.yaml.Groups == nil {
			return "", false
		}
		if group, found := f.yaml.Groups[instance]; found {
			return group.Get(key)
		}
	case "app":
		if f.yaml.Apps == nil {
			return "", false
		}
		if app, found := f.yaml.Apps[instance] ; found {
			return app.Get(key)
		}
	case "repo":
		if f.yaml.Repos == nil {
			return "", false
		}
		if repo, found := f.yaml.Repos[instance]; found {
			return repo.Get(key)
		}
	case "settings":
		return f.yaml.ForjSettings.Get(key)
	default:
		return f.get(object, instance, key)
	}
	return "", false
}
func (f *Forge) ObjectLen(object string) (int) {
	f.Init()
	switch object {
	case "infra":
		return 1
	case "user":
		if f.yaml.Users == nil {
			return 0
		}
		return len(f.yaml.Users)
	case "group":
		if f.yaml.Groups == nil {
			return 0
		}
		return len(f.yaml.Groups)
	case "app":
		if f.yaml.Apps == nil {
			return 0
		}
		return len(f.yaml.Apps)
	case "repo":
		if f.yaml.Repos == nil {
			return 0
		}
		return len(f.yaml.Repos)
	case "settings":
		return 1
	default:
		if v, found := f.yaml.More[object] ; found {
			return len(v)
		}
		return 0
	}
	return 0
}


func (f *Forge) get(object, instance, key string)(value string, found bool)  {
	f.Init()
	if obj, f1 := f.yaml.More[object] ; f1 {
		if instance, f2 := obj[instance] ; f2 {
			value, found = instance[key]
		}
	}
	return
}

func (f *Forge) SetHandler(object, name string, from func(key string) (string, bool), keys ...string) {
	f.Init()
	switch object {
	case "infra":
		f.yaml.Infra.SetHandler(from, keys...)
	case "user":
		if f.yaml.Users == nil {
			f.yaml.Users = make(map[string]UserStruct)
		}
		if user, found := f.yaml.Users[name]; found {
			user.SetHandler(from, keys...)
		} else {
			newuser := UserStruct{}
			newuser.set_forge(f.yaml)
			f.yaml.Users[name] = newuser
		}
	case "group":
		if f.yaml.Groups == nil {
			f.yaml.Groups = make(map[string]GroupStruct)
		}
		if group, found := f.yaml.Groups[name]; found {
			group.SetHandler(from, keys...)
		} else {
			newgroup := GroupStruct{}
			newgroup.set_forge(f.yaml)
			f.yaml.Groups[name] = newgroup
		}
	case "app":
		if f.yaml.Apps == nil {
			f.yaml.Apps = make(map[string]AppStruct)
		}
		if app, found := f.yaml.Apps[name]; found {
			app.SetHandler(from, keys...)
		} else {
			newapp := AppStruct{}
			newapp.set_forge(f.yaml)
			f.yaml.Apps[name] = newapp
		}
	case "repo":
		if f.yaml.Repos == nil {
			f.yaml.Repos = make(map[string]RepoStruct)
		}
		if repo, found := f.yaml.Repos[name]; found {
			repo.SetHandler(from, keys...)
		} else {
			newrepo := RepoStruct{}
			newrepo.set_forge(f.yaml)
			f.yaml.Repos[name] = newrepo
		}
	default:
		f.setHandler(object, name, from, keys...)
	}
}

func (f *Forge) Set(object, name, key, value string) {
	from := func(string) (string, bool) {
		return value, (value != "")
	}
	f.SetHandler(object, name, from, key)
}

func (f *Forge) setHandler(object, instance string, from func(key string) (string, bool), keys ...string)  {
	var object_d map[string]map[string]string
	var instance_d map[string]string

	f.Init()

	if o, found := f.yaml.More[object] ; found && o != nil {
		object_d = o
	} else {
		f.yaml.updated = true
		object_d = make(map[string]map[string]string)
	}
	if i, found := object_d[instance] ; found && i != nil {
		instance_d = i
	} else {
		f.yaml.updated = true
		instance_d = make(map[string]string)
		object_d[instance] = instance_d
	}
	for _, key := range keys {
		var value string
		if v, found := from(key) ; !found {
			continue
		} else {
			value = v
		}
		if v, found := instance_d[key] ; found && v != value {
			instance_d[key] = value
			f.yaml.updated = true
		}

	}
}


func (f *Forge) IsDirty() bool {
	f.Init()

	return f.yaml.updated
}

func (f *Forge) Saved() {
	f.Init()

	f.yaml.updated = false
}

func (f *Forge) Apps() (map[string]AppStruct) {
	f.Init()

	return f.yaml.Apps
}

// Initialize the forge. (Forjfile in repository infra)
func (f *ForgeYaml)Init() {
	if f.Groups == nil {
		f.Groups = make(map[string]GroupStruct)
	}
	if f.Users == nil {
		f.Users = make(map[string]UserStruct)
	}
	if f.More == nil {
		f.More = make(map[string]map[string]map[string]string)
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

}

func (f *ForgeYaml)set_defaults() {
	if f.Apps != nil {
		for name, app := range f.Apps {
			app.name = name
			if app.Driver == "" {
				app.Driver = name
			}
			f.Apps[name] = app
			app.set_forge(f)
		}
	}
	if f.Repos != nil {
		for _, repo := range f.Repos {
			repo.set_forge(f)
		}
	}
	if f.Users != nil {
		for _, user := range f.Users {
			user.set_forge(f)
		}
	}
	if f.Groups != nil {
		for _, group := range f.Groups {
			group.set_forge(f)
		}
	}
	f.Infra.set_forge(f)
	f.ForjSettings.set_forge(f)
}

func (f *ForgeYaml) dirty() {
	f.updated = true
}
