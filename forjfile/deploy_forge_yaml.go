package forjfile

import (
	"fmt"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

// DeployForgeYaml represents a dedicated deployed Forge.
type DeployForgeYaml struct {
	forge *ForgeYaml
	// LocalSettings should not be used from a Forjfile except if this one is a template one.
	LocalSettings WorkspaceStruct    `yaml:"local-settings,omitempty"` // ignored if Normal Forjfile
	ForjSettings  ForjSettingsStruct `yaml:"forj-settings"`
	Infra         *RepoStruct
	Repos         ReposStruct `yaml:"repositories"`
	Apps          AppsStruct  `yaml:"applications"`
	Users         UsersStruct
	Groups        GroupsStruct
	// Collection of Object/Name/Keys=values
	More map[string]map[string]ForjValues `yaml:",inline,omitempty"`
}

// NewDeployForgeYaml creates an empty pre-initialized object.
func NewDeployForgeYaml() (result *DeployForgeYaml) {
	result = new(DeployForgeYaml)
	if result.Apps == nil {
		result.Apps = make(AppsStruct)
	}
	if result.Groups == nil {
		result.Groups = make(GroupsStruct)
	}
	if result.Users == nil {
		result.Users = make(UsersStruct)
	}
	if result.Repos == nil {
		result.Repos = make(ReposStruct)
	}
	if result.More == nil {
		result.More = make(map[string]map[string]ForjValues)
	}
	return

}

// Init ensure all object are well initialized to avoid core dump
func (f *DeployForgeYaml) Init(forge *ForgeYaml) bool {
	if f == nil {
		return false
	}

	f.forge = forge
	return f.init()
}

// GetString return the value in String. It ignore the origin of the data (default or not)
func (f *DeployForgeYaml) GetString(object, instance, key string) (string, bool) {
	v, found := f.Get(object, instance, key)
	return v.GetString(), found
}

// GetInstances return a list of instances from an object type.
func (f *DeployForgeYaml) GetInstances(object string) (ret []string) {
	if !f.Init(f.forge) {
		return nil
	}
	ret = []string{}
	switch object {
	case "user":
		if f.Users == nil {
			return
		}

		ret = make([]string, len(f.Users))
		iCount := 0
		for user := range f.Users {
			ret[iCount] = user
			iCount++
		}
		return
	case "group":
		if f.Groups == nil {
			return
		}

		ret = make([]string, len(f.Groups))
		iCount := 0
		for group := range f.Groups {
			ret[iCount] = group
			iCount++
		}
		return
	case "app":
		if f.Apps == nil {
			return
		}

		ret = make([]string, len(f.Apps))
		iCount := 0
		for app := range f.Apps {
			ret[iCount] = app
			iCount++
		}
		return
	case "repo":
		if f.Repos == nil {
			return
		}

		ret = make([]string, len(f.Repos))
		iCount := 0
		for repo := range f.Repos {
			ret[iCount] = repo
			iCount++
		}
		return
	case "infra", "settings":
		return
	default:
		if instances, found := f.More[object]; found {
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

// Get return the value of the object instance key as ValueStruct.
func (f *DeployForgeYaml) Get(object, instance, key string) (value *goforjj.ValueStruct, _ bool) {
	if !f.init() {
		return
	}
	switch object {
	case "infra":
		if f.Infra == nil {
			return
		}
		if key == "name" {
			if f.Infra.More == nil {
				return
			}
			if v, found := f.Infra.More["name"]; found && v != "" {
				return value.Set(v), true
			}
			if f.Infra.name != "" {
				return value.Set(f.Infra.name), true
			}
		}
		return f.Infra.Get(key)
	case "user":
		if f.Users == nil {
			return
		}
		if user, found := f.Users[instance]; found {
			return user.Get(key)
		}
	case "group":
		if f.Groups == nil {
			return
		}
		if group, found := f.Groups[instance]; found {
			return group.Get(key)
		}
	case "app":
		if f.Apps == nil {
			return
		}
		if app, found := f.Apps[instance]; found {
			return app.Get(key)
		}
	case "repo":
		if f.Repos == nil {
			return
		}
		if repo, found := f.Repos[instance]; found {
			if key != "" {
				return repo.Get(key)
			}
			return nil, found
		}
	case "settings":
		return f.ForjSettings.Get(instance, key)
	default:
		return f.get(object, instance, key)
	}
	return
}

// Set a value to the object instance key.
func (f *DeployForgeYaml) Set(object, name, key, value string) {
	from := func(string) (string, bool) {
		return value, (value != "")
	}
	f.SetHandler(object, name, from, (*ForjValue).Set, key)
}

// Remove a key value found in the object instance
func (f *DeployForgeYaml) Remove(object, name, key string) {
	from := func(string) (_ string, _ bool) {
		return "", true
	}

	f.SetHandler(object, name, from, (*ForjValue).Clean, key)
}

// SetDefault a value as default value in the object instance key.
func (f *DeployForgeYaml) SetDefault(object, name, key, value string) {
	from := func(string) (string, bool) {
		return value, (value != "")
	}
	f.SetHandler(object, name, from, (*ForjValue).SetDefault, key)
}

// SetHandler define the core set function, and call `from` function to determine the source of the data. (set/default or not found)
func (f *DeployForgeYaml) SetHandler(object, name string, from func(string) (string, bool), set func(*ForjValue, string) bool, keys ...string) {
	if !f.init() {
		return
	}
	switch object {
	case "infra":
		f.Infra.SetHandler(from, keys...)
	case "user":
		if f.Users == nil {
			f.Users = make(map[string]*UserStruct)
		}
		if user, found := f.Users[name]; found {
			user.SetHandler(from, keys...)
		} else {
			newuser := UserStruct{}
			newuser.set_forge(f.forge)
			f.Users[name] = &newuser
		}
	case "group":
		if f.Groups == nil {
			f.Groups = make(map[string]*GroupStruct)
		}
		if group, found := f.Groups[name]; found {
			group.SetHandler(from, keys...)
		} else {
			newgroup := GroupStruct{}
			newgroup.set_forge(f.forge)
			f.Groups[name] = &newgroup
		}
	case "app":
		if f.Apps == nil {
			f.Apps = make(map[string]*AppStruct)
		}
		if app, found := f.Apps[name]; found {
			app.SetHandler(from, set, keys...)
		} else {
			newapp := AppStruct{}
			newapp.set_forge(f.forge)
			f.Apps[name] = &newapp
		}
	case "repo":
		if f.Repos == nil {
			f.Repos = make(map[string]*RepoStruct)
		}
		if repo, found := f.Repos[name]; found {
			repo.SetHandler(from, keys...)
		} else {
			newrepo := RepoStruct{}
			newrepo.set_forge(f.forge)
			f.Repos[name] = &newrepo
			newrepo.SetHandler(func(string) (string, bool) {
				return name, true
			}, FieldRepoName)
			newrepo.SetHandler(from, keys...)
		}
	case "settings", "forj-settings":
		f.ForjSettings.SetHandler(name, from, keys...)
	default:
		f.setHandler(object, name, from, set, keys...)
	}
}

// ObjectLen returns the number of instances of object type.
func (f *DeployForgeYaml) ObjectLen(object string) int {
	if !f.init() {
		return 0
	}
	switch object {
	case "infra":
		return 1
	case "user":
		if f.Users == nil {
			return 0
		}
		return len(f.Users)
	case "group":
		if f.Groups == nil {
			return 0
		}
		return len(f.Groups)
	case "app":
		if f.Apps == nil {
			return 0
		}
		return len(f.Apps)
	case "repo":
		if f.Repos == nil {
			return 0
		}
		return len(f.Repos)
	case "settings":
		return 1
	default:
		if v, found := f.More[object]; found {
			return len(v)
		}
		return 0
	}
	return 0
}

// HasApps return a bool if rules are all true on at least one application.
// a rule is a string formatted as '<key>:<value>'
// a rule is true on an application if it has the key value set to <value>
//
// If the rule is not well formatted, an error is returned.
// If the repo has no application, HasApps return false.
// If no rules are provided and at least one application exist, HasApps return true.
//
// TODO: Write Unit test of HasApps
func (f *DeployForgeYaml) HasApps(rules ...string) (found bool, err error) {
	if f.Apps == nil {
		return
	}
	for _, app := range f.Apps {
		found = true
		for _, rule := range rules {
			ruleToCheck := strings.Split(rule, ":")
			if len(ruleToCheck) != 2 {
				err = fmt.Errorf("rule '%s' is invalid. Format supported is '<key>:<value>'", rule)
				return
			}
			if v, found2 := app.Get(ruleToCheck[0]); found2 && v.GetString() != ruleToCheck[1] {
				found = false
				break
			}
		}
		if found {
			gotrace.Trace("Found an application which meets '%s'", rules)
			return
		}
	}
	gotrace.Trace("NO application found which meets '%s'", rules)
	return
}

// ----------------- Create objects functions

// NewRepoStruct create a new Repo in the Forjfile.
func (f *DeployForgeYaml) NewRepoStruct(name string) (repo *RepoStruct) {
	if f == nil {
		return
	}
	if v, found := f.Repos[name]; found {
		return v
	}
	repo = new(RepoStruct)
	repo.forge = f.forge
	repo.name = name
	repo.initApp()
	return
}

// ---------------- private functions

func (f *DeployForgeYaml) get(object, instance, key string) (value *goforjj.ValueStruct, found bool) {
	if obj, f1 := f.More[object]; f1 {
		if instance, f2 := obj[instance]; f2 {
			v, f3 := instance[key]
			value, found = value.SetIfFound(v.Get(), f3)
		}
	}
	return
}

func (f *DeployForgeYaml) setHandler(object, instance string, from func(string) (string, bool), set func(*ForjValue, string) bool, keys ...string) {
	var objectData map[string]ForjValues
	var instanceData ForjValues

	if o, found := f.More[object]; found && o != nil {
		objectData = o
	} else {
		f.forge.updated = true
		objectData = make(map[string]ForjValues)
		f.More[object] = objectData
	}
	if i, found := objectData[instance]; found && i != nil {
		instanceData = i
	} else {
		f.forge.updated = true
		instanceData = make(map[string]ForjValue)
		objectData[instance] = instanceData
	}
	for _, key := range keys {
		var value string
		if v, found := from(key); !found {
			continue
		} else {
			value = v
		}
		if v, found := instanceData[key]; found && v.Get() != value {
			set(&v, value)
			instanceData[key] = v
			f.forge.updated = true
		} else {
			if !found {
				set(&v, value)
				instanceData[key] = v
				f.forge.updated = true
			}
		}

	}
}

func (f *DeployForgeYaml) init() bool {
	if f == nil {
		return false
	}
	if f.Infra == nil {
		f.Infra = new(RepoStruct)
	}
	if f.Infra.More == nil {
		f.Infra.More = make(map[string]string)
	}
	if f.More == nil {
		f.More = make(map[string]map[string]ForjValues)
	}
	if f.Apps == nil {
		f.Apps = make(map[string]*AppStruct)
	}
	if f.Repos == nil {
		f.Repos = make(ReposStruct)
	}
	if f.Groups == nil {
		f.Groups = make(map[string]*GroupStruct)
	}
	if f.Users == nil {
		f.Users = make(map[string]*UserStruct)
	}

	return true
}

func (f *DeployForgeYaml) mergeFrom(from *DeployForgeYaml) error {
	f.Apps.mergeFrom(from.Apps)
	f.ForjSettings.mergeFrom(&from.ForjSettings)
	f.Groups.mergeFrom(from.Groups)
	f.Infra.mergeFrom(from.Infra)
	f.Users.mergeFrom(from.Users)
	f.Repos.mergeFrom(from.Repos)
	for k, v := range from.More {
		f.More[k] = v
	}
	return nil
}

func (f *DeployForgeYaml) initDefaults(forge *ForgeYaml) {
	// Cleanup LocalSettings to ensure no local setting remain in a Forjfile
	if f == nil {
		return
	}
	f.forge = forge

	f.LocalSettings = WorkspaceStruct{}

	if f.Apps != nil {
		for name, app := range f.Apps {
			if app == nil {
				continue
			}
			app.name = name
			if app.Driver == "" {
				app.Driver = name
			}
			app.set_forge(forge)
			f.Apps[name] = app
		}
	}
	if f.Repos != nil {
		for name, repo := range f.Repos {
			if repo == nil {
				// Repo can be nil if we did not defined any fields under his name.
				// ie : forjj-modules:
				// or
				// forjj-modules: nil
				repo = new(RepoStruct)
			}
			repo.name = name
			repo.set_forge(forge)
			f.Repos[name] = repo
		}
	}
	if f.Users != nil {
		for name, user := range f.Users {
			if user == nil {
				continue
			}
			user.set_forge(forge)
			f.Users[name] = user
		}
	}
	if f.Groups != nil {
		for name, group := range f.Groups {
			if group == nil {
				continue
			}
			group.set_forge(forge)
			f.Groups[name] = group
		}
	}
	if f.Infra == nil {
		f.Infra = new(RepoStruct)
	}
	f.Infra.set_forge(forge)
	f.ForjSettings.set_forge(forge)
}

// GetRepo return the object found
func (f *DeployForgeYaml) GetRepo(name string) (r *RepoStruct, found bool) {
	r, found = f.Repos[name]
	return
}

func (f *DeployForgeYaml) Model() ForgeModel {
	model := ForgeModel{
		forge: f,
	}

	return model
}
func (f *DeployForgeYaml) GetObjectInstance(object, instance string) interface{} {
	f.initDefaults(f.forge)
	
	switch object {
	case "user":
		if f.Users == nil {
			return nil
		}
		if user, found := f.Users[instance]; found {
			return user
		}
	case "group":
		if f.Groups == nil {
			return nil
		}
		if group, found := f.Groups[instance]; found {
			return group
		}
	case "app":
		if f.Apps == nil {
			return nil
		}
		if app, found := f.Apps[instance]; found {
			return app
		}
	case "repo":
		if f.Repos == nil {
			return nil
		}
		if repo, found := f.Repos[instance]; found {
			return repo
		}
	case "settings":
		return f.ForjSettings.GetInstance(instance)
	default:
		return f.getInstance(object, instance)
	}
	return nil
}

func (f *DeployForgeYaml) getInstance(object, instance string) (_ map[string]ForjValue) {
	f.initDefaults(f.forge)

	if obj, f1 := f.More[object]; f1 {
		if i, f2 := obj[instance]; f2 {
			return i
		}
	}
	return
}