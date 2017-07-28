package forjfile

import (
	"os"
	"path"
	"fmt"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"forjj/utils"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

// ForjfileTmpl is the Memory expansion of an external Forjfile (used to create a Forge)
type ForjfileTmpl struct {
	Workspace WorkspaceStruct // See workspace.go
	yaml ForgeYaml
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
	// LocalSettings should not be used from a Forjfile except if this one is a template one.
	LocalSettings WorkspaceStruct `yaml:"local-settings,omitempty"` // ignored if Normal Forjfile
	ForjSettings ForjSettingsStruct `yaml:"forj-settings"`
	Infra *RepoStruct
	Repos ReposStruct `yaml:"repositories"`
	Apps map[string]*AppStruct `yaml:"applications"`
	Users map[string]*UserStruct
	Groups map[string]*GroupStruct
	// Collection of Object/Name/Keys=values
	More map[string]map[string]ForjValues `yaml:",inline,omitempty"`

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

	if e := yaml.Unmarshal(yaml_data, &f.yaml) ; e != nil {
		err = fmt.Errorf("Unable to load %s. %s", file, e)
		return
	}

	f.Workspace = f.yaml.LocalSettings
	gotrace.Trace("Forjfile template '%s' has been loaded.", file)
	// Setting defaults
	f.yaml.set_defaults()
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

func (f *Forge)GetInfraName() string {
	return f.yaml.Infra.name
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

	if e := yaml.Unmarshal(yaml_data, &f.yaml) ; e != nil {
		err = fmt.Errorf("Unable to load %s. %s", file, e)
		return
	}

	f.yaml.set_defaults()
	loaded = true
	gotrace.Trace("Forge loaded from '%s'.", aPath)
	return
}

func (f *Forge)Forjfile() *ForgeYaml {
	return f.yaml
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

	*f.yaml = ft.yaml
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
		f.yaml.More = make(map[string]map[string]ForjValues)
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

func (f *Forge) GetString(object, instance, key string) (string, bool) {
	v, found := f.Get(object, instance, key)
	return v.GetString(), found
}

func (f *Forge) Get(object, instance, key string) (value *goforjj.ValueStruct, _ bool) {
	if ! f.Init() { return }
	switch object {
	case "infra":
		if f.yaml.Infra == nil {
			return
		}
		if key == "name" {
			if f.yaml.Infra.More == nil {
				return
			}
			if v, found := f.yaml.Infra.More["name"] ; found && v != "" {
				return value.Set(v), true
			}
			if f.yaml.Infra.name != "" {
				return value.Set(f.yaml.Infra.name), true
			}
		}
		return f.yaml.Infra.Get(key)
	case "user":
		if f.yaml.Users == nil {
			return
		}
		if user, found := f.yaml.Users[instance] ; found {
			return user.Get(key)
		}
	case "group":
		if f.yaml.Groups == nil {
			return
		}
		if group, found := f.yaml.Groups[instance]; found {
			return group.Get(key)
		}
	case "app":
		if f.yaml.Apps == nil {
			return
		}
		if app, found := f.yaml.Apps[instance] ; found {
			return app.Get(key)
		}
	case "repo":
		if f.yaml.Repos == nil {
			return
		}
		if repo, found := f.yaml.Repos[instance]; found {
			return repo.Get(key)
		}
	case "settings":
		return f.yaml.ForjSettings.Get(instance, key)
	default:
		return f.get(object, instance, key)
	}
	return
}

func (f *Forge) GetObjectInstance(object, instance string) interface{} {
	if ! f.Init() { return nil }
	switch object {
	case "user":
		if f.yaml.Users == nil {
			return nil
		}
		if user, found := f.yaml.Users[instance] ; found {
			return user
		}
	case "group":
		if f.yaml.Groups == nil {
			return nil
		}
		if group, found := f.yaml.Groups[instance]; found {
			return group
		}
	case "app":
		if f.yaml.Apps == nil {
			return nil
		}
		if app, found := f.yaml.Apps[instance] ; found {
			return app
		}
	case "repo":
		if f.yaml.Repos == nil {
			return nil
		}
		if repo, found := f.yaml.Repos[instance]; found {
			return repo
		}
	case "settings":
		return f.yaml.ForjSettings.GetInstance(instance)
	default:
		return f.getInstance(object, instance)
	}
	return nil
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

func (f *Forge) get(object, instance, key string)(value *goforjj.ValueStruct, found bool)  {
	if ! f.Init() { return }
	if obj, f1 := f.yaml.More[object] ; f1 {
		if instance, f2 := obj[instance] ; f2 {
			v, f3 := instance[key]
			value, found = value.SetIfFound(v.Get(), f3)
		}
	}
	return
}

func (f *Forge) getInstance(object, instance string) (_ map[string]ForjValue) {
	if ! f.Init() { return }
	if obj, f1 := f.yaml.More[object] ; f1 {
		if i, f2 := obj[instance] ; f2 {
			return i
		}
	}
	return
}


func (f *Forge) SetHandler(object, name string, from func(string) (string, bool), set func(*ForjValue, string) (bool), keys ...string) {
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
			app.SetHandler(from, set, keys...)
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
		f.setHandler(object, name, from, set, keys...)
	}
}

func (f *Forge) Remove(object, name, key string) {
	from := func(string) (_ string, _ bool) {
		return "", true
	}

	f.SetHandler(object, name, from, (*ForjValue).Clean, key)
}

func (f *Forge) Set(object, name, key, value string) {
	from := func(string) (string, bool) {
		return value, (value != "")
	}
	f.SetHandler(object, name, from, (*ForjValue).Set, key)
}

func (f *Forge) SetDefault(object, name, key, value string) {
	from := func(string) (string, bool) {
		return value, (value != "")
	}
	f.SetHandler(object, name, from, (*ForjValue).SetDefault, key)
}

func (f *Forge) setHandler(object, instance string, from func(string) (string, bool), set func(*ForjValue, string) (bool), keys ...string)  {
	var object_d map[string]ForjValues
	var instance_d ForjValues

	if ! f.Init() { return }

	if o, found := f.yaml.More[object] ; found && o != nil {
		object_d = o
	} else {
		f.yaml.updated = true
		object_d = make(map[string]ForjValues)
	}
	if i, found := object_d[instance] ; found && i != nil {
		instance_d = i
	} else {
		f.yaml.updated = true
		instance_d = make(map[string]ForjValue)
		object_d[instance] = instance_d
	}
	for _, key := range keys {
		var value string
		if v, found := from(key) ; !found {
			continue
		} else {
			value = v
		}
		if v, found := instance_d[key] ; found && v.Get() != value {
			set(&v, value)
			instance_d[key] = v
			f.yaml.updated = true
		} else {
			if !found {
				set(&v, value)
				instance_d[key] = v
				f.yaml.updated = true
			}
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
		f.More = make(map[string]map[string]ForjValues)
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
	// Cleanup LocalSettings to ensure no local setting remain in a Forjfile
	f.LocalSettings = WorkspaceStruct{}

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
