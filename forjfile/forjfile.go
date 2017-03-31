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
	Infra *RepoStruct
	Repos ReposStruct `yaml:"repositories"`
	Apps map[string]*AppStruct `yaml:"applications"`
	Users map[string]*UserStruct
	Groups map[string]*GroupStruct
	// Collection of Object/Name/Keys=values
	More map[string]map[string]map[string]string `yaml:",inline"`

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

	var forj_path string
	forj_path, err = utils.Abs(aPath) ; if err != nil { return }
	if  forj_path != "." {
		if fi, e := os.Stat(forj_path) ; err != nil {
			err = e
			return
		} else {
			if !fi.Mode().IsDir() {
				return f, loaded, fmt.Errorf("'%s' must be a path to '%s'.", aPath, file_name)
			}
		}
	}

	file := path.Join(forj_path, file_name)

	if _, e := os.Stat(file); os.IsNotExist(e) {
		return
	}

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

	gotrace.Trace("Forjfile template '%s' has been loaded.", file)
	// Setting defaults
	f.ForgeYaml.set_defaults()
	loaded = true
	return
}

func (f *Forge)SetInfraAsRepo() {
	// Copy the infra repo in list of repositories, tagged as infra.
	if !f.Init() { return }

	var repo *RepoStruct

	if v, found := f.yaml.Infra.More["name"] ; found && v != "" {
		f.yaml.Infra.name = v
	}

	if f.yaml.Infra.name == "" || f.yaml.Infra.name == "none" {
		return
	}

	if r, found_repo := f.yaml.Repos[f.yaml.Infra.name]; found_repo {
		repo = r
	}
	if repo == nil {
		repo = new(RepoStruct)
		f.yaml.Repos[f.yaml.Infra.name] = repo
	}
	repo.setFromInfra(f.yaml.Infra)
}

// Load : Load Forjfile stored in a Repository.
func (f *Forge)Load() (loaded bool, err error) {
	var (
		yaml_data []byte
		file string
	)

	if ! f.Init() { return false, fmt.Errorf("Forge is nil.") }

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

	var forjfile_tmpl ForjfileTmpl
	if e := yaml.Unmarshal(yaml_data, &forjfile_tmpl) ; e != nil {
		err = fmt.Errorf("Unable to load %s. %s", file, e)
		return
	}

	f.yaml = &forjfile_tmpl.ForgeYaml
	f.yaml.set_defaults()
	loaded = true
	gotrace.Trace("Forge loaded from '%s'.", aPath)
	return
}


func loadFile(aPath string) (file string, yaml_data[]byte, err error) {
	file, err = utils.Abs(aPath)
	if err != nil { return }

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
	if !f.Init() { return }

	*f.yaml = ft.ForgeYaml
	f.yaml.updated = true
}

func (f *Forge)Init() bool {
	if f == nil {
		return false
	}
	if f.yaml == nil {
		f.yaml = new(ForgeYaml)
	}
	if f.yaml.Infra == nil {
		f.yaml.Infra = new(RepoStruct)
	}
	if f.yaml.Infra.More == nil {
		f.yaml.Infra.More = make(map[string]string)
	}
	if f.yaml.More == nil {
		f.yaml.More = make(map[string]map[string]map[string]string)
	}
	if f.yaml.Apps == nil {
		f.yaml.Apps = make(map[string]*AppStruct)
	}
	if f.yaml.Repos == nil {
		f.yaml.Repos = make(ReposStruct)
	}
	if f.yaml.Groups == nil {
		f.yaml.Groups = make(map[string]*GroupStruct)
	}
	if f.yaml.Users == nil {
		f.yaml.Users = make(map[string]*UserStruct)
	}

	return true
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
	if ! f.Init() { return fmt.Errorf("Forge is nil.") }
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

func (f *Forge) GetInstances(object string) (ret []string) {
	if ! f.Init() { return nil }
	ret = []string{}
	switch object {
	case "user":
		if f.yaml.Users == nil { return }

		ret = make([]string, len(f.yaml.Users))
		iCount := 0
		for user := range f.yaml.Users {
			ret[iCount] = user
			iCount++
		}
		return
	case "group":
		if f.yaml.Groups == nil { return }

		ret = make([]string, len(f.yaml.Groups))
		iCount := 0
		for group := range f.yaml.Groups {
			ret[iCount] = group
			iCount++
		}
		return
	case "app":
		if f.yaml.Apps == nil { return }

		ret = make([]string, len(f.yaml.Apps))
		iCount := 0
		for app := range f.yaml.Apps {
			ret[iCount] = app
			iCount++
		}
		return
	case "repo":
		if f.yaml.Repos == nil { return }

		ret = make([]string, len(f.yaml.Repos))
		iCount := 0
		for repo := range f.yaml.Repos {
			ret[iCount] = repo
			iCount++
		}
		return
	case "infra", "settings":
		return
	default:
		if instances, found := f.yaml.More[object] ; found {
			ret = make([]string, len(instances))
			iCount := 0
			for instance := range instances {
				ret[iCount] = instance
				iCount++
			}
		}
	}
	return
}

func (f *Forge) GetInfraInstance() string {
	if f == nil { return "" }
	if f.yaml.Infra == nil { return "" }
	return f.yaml.Infra.Upstream
}

func (f *Forge) Get(object, instance, key string) (string, bool) {
	if ! f.Init() { return "", false }
	switch object {
	case "infra":
		if f.yaml.Infra == nil {
			return "", false
		}
		if key == "name" {
			if f.yaml.Infra.More == nil {
				return "", false
			}
			if v, found := f.yaml.Infra.More["name"] ; found && v != "" {
				return v, true
			}
			if f.yaml.Infra.name != "" {
				return f.yaml.Infra.name, true
			}
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
		return f.yaml.ForjSettings.Get(instance, key)
	default:
		return f.get(object, instance, key)
	}
	return "", false
}

func (f *Forge) ObjectLen(object string) (int) {
	if ! f.Init() { return 0 }
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
	if ! f.Init() { return "", false }
	if obj, f1 := f.yaml.More[object] ; f1 {
		if instance, f2 := obj[instance] ; f2 {
			value, found = instance[key]
		}
	}
	return
}

func (f *Forge) SetHandler(object, name string, from func(key string) (string, bool), keys ...string) {
	if ! f.Init() { return }
	switch object {
	case "infra":
		f.yaml.Infra.SetHandler(from, keys...)
	case "user":
		if f.yaml.Users == nil {
			f.yaml.Users = make(map[string]*UserStruct)
		}
		if user, found := f.yaml.Users[name]; found {
			user.SetHandler(from, keys...)
		} else {
			newuser := UserStruct{}
			newuser.set_forge(f.yaml)
			f.yaml.Users[name] = &newuser
		}
	case "group":
		if f.yaml.Groups == nil {
			f.yaml.Groups = make(map[string]*GroupStruct)
		}
		if group, found := f.yaml.Groups[name]; found {
			group.SetHandler(from, keys...)
		} else {
			newgroup := GroupStruct{}
			newgroup.set_forge(f.yaml)
			f.yaml.Groups[name] = &newgroup
		}
	case "app":
		if f.yaml.Apps == nil {
			f.yaml.Apps = make(map[string]*AppStruct)
		}
		if app, found := f.yaml.Apps[name]; found {
			app.SetHandler(from, keys...)
		} else {
			newapp := AppStruct{}
			newapp.set_forge(f.yaml)
			f.yaml.Apps[name] = &newapp
		}
	case "repo":
		if f.yaml.Repos == nil {
			f.yaml.Repos = make(map[string]*RepoStruct)
		}
		if repo, found := f.yaml.Repos[name]; found {
			repo.SetHandler(from, keys...)
		} else {
			newrepo := RepoStruct{}
			newrepo.set_forge(f.yaml)
			f.yaml.Repos[name] = &newrepo
		}
	case "settings":
		f.yaml.ForjSettings.SetHandler(name, from, keys...)
	default:
		f.setHandler(object, name, from, keys...)
	}
}

func (f *Forge) Remove(object, name, key string) {
	from := func(string) (string, bool) {
		return "", true
	}
	f.SetHandler(object, name, from, key)
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

	if ! f.Init() { return }

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
	if ! f.Init() { return false }

	return f.yaml.updated
}

func (f *Forge) Saved() {
	if ! f.Init() { return }

	f.yaml.updated = false
}

func (f *Forge) Apps() (map[string]*AppStruct) {
	if ! f.Init() { return nil }

	return f.yaml.Apps
}

// Initialize the forge. (Forjfile in repository infra)
func (f *ForgeYaml)Init() {
	if f.Groups == nil {
		f.Groups = make(map[string]*GroupStruct)
	}
	if f.Users == nil {
		f.Users = make(map[string]*UserStruct)
	}
	if f.More == nil {
		f.More = make(map[string]map[string]map[string]string)
	}

	if f.Infra.More == nil {
		f.Infra.More = make(map[string]string)
	}

	if f.Repos == nil {
		f.Repos = make(map[string]*RepoStruct)
	}

	if f.Apps == nil {
		f.Apps = make(map[string]*AppStruct)
	}

}

func (f *ForgeYaml)set_defaults() {
	if f.Apps != nil {
		for name, app := range f.Apps {
			if app == nil { continue }
			app.name = name
			if app.Driver == "" {
				app.Driver = name
			}
			app.set_forge(f)
			f.Apps[name] = app
		}
	}
	if f.Repos != nil {
		for name, repo := range f.Repos {
			if repo == nil { continue }
			repo.name = name
			repo.set_forge(f)
			f.Repos[name] = repo
		}
	}
	if f.Users != nil {
		for name, user := range f.Users {
			if user == nil { continue }
			user.set_forge(f)
			f.Users[name] = user
		}
	}
	if f.Groups != nil {
		for name, group := range f.Groups {
			if group == nil { continue }
			group.set_forge(f)
			f.Groups[name] = group
		}
	}
	if f.Infra == nil { f.Infra = new(RepoStruct) }
	f.Infra.set_forge(f)
	f.ForjSettings.set_forge(f)
}

func (f *ForgeYaml) dirty() {
	f.updated = true
}
