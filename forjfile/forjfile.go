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
	Forge `yaml:",inline"`
}

// Forge is the Memory expand of a repository Forjfile.
type Forge struct {
	updated bool
	updated_msg string
	infra_path string // Infra path used to create/save/load Forjfile

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
	RepoStruct
}

type WorkspaceStruct struct {
	updated bool
	DockerBinPath          string `yaml:"docker-exe-path"`    // Docker static binary path
	Contrib_repo_path      string `yaml:"contribs-repo"`      // Contrib Repo path used.
	Flow_repo_path         string `yaml:"flows-repo"`         // Flow repo path used.
	Repotemplate_repo_path string `yaml:"repotemplates-repo"` // Repotemplate Path used.
	More                   map[string]string `yaml:",inline"`
}


const File_name = "Forjfile"

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
	f.Forge.set_defaults()
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

	aPath := path.Join(f.infra_path, File_name)
	if _, err = os.Stat(aPath); err != nil {
		return
	}
	if fi, d, e := loadFile(aPath) ; e != nil {
		err = e
		return
	} else {
		yaml_data = d
		file = fi
	}

	if e := yaml.Unmarshal(yaml_data, f) ; e != nil {
		err = fmt.Errorf("Unable to load %s. %s", file, e)
		return
	}

	f.set_defaults()
	loaded = true
	return
}

func (f *Forge)set_defaults() {
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

func loadFile(aPath string) (file string, yaml_data[]byte, err error) {
		forj_path := path.Clean(aPath)

	if forj_path != "." {
		if _, err = os.Stat(forj_path) ; err != nil {
			return
		}
	}

	// TODO: interpret ~ to $HOME and get it from path.Home()
	file = path.Join(forj_path, File_name)
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
	forge.ForjSettings.is_template = true
	return forge.save(aPath)
}

func (f *Forge) Get(object, instance, key string) (string, bool) {
	switch object {
	case "infra":
		if key == "name" && f.Infra.Name != "" {
			return f.Infra.Name, true
		}
		return f.Infra.Get(key)
	case "user":
		if f.Users == nil {
			return "", false
		}
		if user, found := f.Users[instance] ; found {
			return user.Get(key)
		}
	case "group":
		if f.Groups == nil {
			return "", false
		}
		if group, found := f.Groups[instance]; found {
			return group.Get(key)
		}
	case "app":
		if f.Apps == nil {
			return "", false
		}
		if app, found := f.Apps[instance] ; found {
			return app.Get(key)
		}
	case "repo":
		if f.Repos == nil {
			return "", false
		}
		if repo, found := f.Repos[instance]; found {
			return repo.Get(key)
		}
	case "settings":
		return f.ForjSettings.Get(key)
	default:
		return f.get(object, instance, key)
	}
	return "", false
}

func (f *Forge) get(object, instance, key string)(value string, found bool)  {
	if obj, f1 := f.More[object] ; f1 {
		if instance, f2 := obj[instance] ; f2 {
			value, found = instance[key]
		}
	}
	return
}

func (f *Forge) Set(object, name, key, value string) {
	switch object {
	case "infra":
		f.Infra.Set(key, value)
	case "user":
		if f.Users == nil {
			f.Users = make(map[string]UserStruct)
		}
		if user, found := f.Users[name]; found {
			user.Set(key, value)
		} else {
			newuser := UserStruct{}
			newuser.set_forge(f)
			f.Users[name] = newuser
		}
	case "group":
		if f.Groups == nil {
			f.Groups = make(map[string]GroupStruct)
		}
		if group, found := f.Groups[name]; found {
			group.Set(key, value)
		} else {
			newgroup := GroupStruct{}
			newgroup.set_forge(f)
			f.Groups[name] = newgroup
		}
	case "app":
		if f.Apps == nil {
			f.Apps = make(map[string]AppStruct)
		}
		if app, found := f.Apps[name]; found {
			app.Set(key, value)
		} else {
			newapp := AppStruct{}
			newapp.set_forge(f)
			f.Apps[name] = newapp
		}
	case "repo":
		if f.Repos == nil {
			f.Repos = make(map[string]RepoStruct)
		}
		if repo, found := f.Repos[name]; found {
			repo.Set(key, value)
		} else {
			newrepo := RepoStruct{}
			newrepo.set_forge(f)
			f.Repos[name] = newrepo
		}
	default:
		f.set(object, name, key, value)
	}
}

func (f *Forge) set(object, instance, key, value string)  {
	var object_d map[string]map[string]string
	var instance_d map[string]string

	if o, found := f.More[object] ; found && o != nil {
		object_d = o
	} else {
		f.updated = true
		object_d = make(map[string]map[string]string)
	}
	if i, found := object_d[instance] ; found && i != nil {
		instance_d = i
	} else {
		f.updated = true
		instance_d = make(map[string]string)
		object_d[instance] = instance_d
	}
	if v, found := instance_d[key] ; found && v != value {
		instance_d[key] = value
		f.updated = true
	}
}

func (f *Forge) IsDirty() bool {
	return f.updated
}

func (f *Forge) dirty() {
	f.updated = true
}

func (f *Forge) Saved() {
	f.updated = false
}
