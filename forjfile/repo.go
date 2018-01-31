package forjfile

import (
	"github.com/forj-oss/goforjj"
	"forjj/drivers"
	"github.com/forj-oss/forjj-modules/trace"
	"strings"
	"fmt"
)

type ReposStruct map[string]*RepoStruct

func (r ReposStruct) MarshalYAML() (interface{}, error) {
	to_marshal := make(map[string]*RepoStruct)
	for name, repo := range r {
		if repo == nil {
			gotrace.Error("Unable to save Repository '%s'. Repo data is nil.", name)
			continue
		}
		if ! repo.is_infra {
			to_marshal[name] = repo
		}
	}
	return to_marshal, nil
}

func (r ReposStruct) LoadRelApps() {
	for _, repo := range r {
		repo.LoadRelApps()
	}
}

type RepoStruct struct {
	name         string
	is_infra     bool
	forge        *ForgeYaml
	owner        string
	driverOwner  *drivers.Driver
	Upstream     string `yaml:"upstream-app,omitempty"` // Name of the application upstream hosting this repository.
	GitRemote    string `yaml:"git-remote,omitempty"`
	remote       goforjj.PluginRepoRemoteUrl // Git remote string to use/set
	Title        string `yaml:",omitempty"`
	RepoTemplate string `yaml:"repo-template,omitempty"`
	Flow         RepoFlow `yaml:",omitempty"`
	More         map[string]string `yaml:",inline"`
	apps         map[string]*AppStruct // List of applications connected to this repo.
	Apps         map[string]string `yaml:"in-relation-with"`// key: <AppRelName>, value: <appName>
}

type RepoFlow struct {
	Name string
	objects map[string]map[string]string
}

func (r *RepoStruct)Owner() string {
	if r == nil {
		return ""
	}

	return r.owner
}

func (r *RepoStruct)setFromInfra(infra *RepoStruct) {
	if r == nil {
		return
	}
	*r = *infra
	delete(r.More, "name")
	r.is_infra = true
}

func (r *RepoStruct)setToInfra(infra *RepoStruct) {
	if r == nil {
		return
	}

	*infra = *r
	infra.is_infra = false // Unset it to ensure data is saved in yaml
}

func (r *RepoStruct)GetString(field string) (string) {
	if r == nil {
		return ""
	}

	if v, found := r.Get(field) ; found {
		return v.GetString()
	}
	return ""
}

func (r *RepoStruct)RemoteUrl() string {
	if r == nil {
		return ""
	}

	return r.remote.Url
}

func (r *RepoStruct)RemoteGit() string {
	if r == nil {
		return ""
	}

	return r.remote.Ssh
}

func (r *RepoStruct)Get(field string) (value *goforjj.ValueStruct, _ bool) {
	if r == nil {
		return
	}
	switch field {
	case "name":
		return value.SetIfFound(r.name, (r.name != ""))
	case "upstream":
		return value.SetIfFound(r.Upstream, (r.Upstream != ""))
	case "git-remote":
		return value.SetIfFound(r.GitRemote, (r.GitRemote != ""))
	case "remote":
		return value.SetIfFound(r.remote.Ssh, (r.remote.Ssh != ""))
	case "remote-url":
		return value.SetIfFound(r.remote.Url, (r.remote.Url != ""))
	case "title":
		return value.SetIfFound(r.Title, (r.Title != ""))
	case "flow":
		return value.SetIfFound(r.Flow.Name, (r.Flow.Name != ""))
	case "repo-template":
		return value.SetIfFound(r.RepoTemplate, (r.RepoTemplate != ""))
	default:
		v, f := r.More[field]
		return value.SetIfFound(v, f)
	}
	return
}

func (r *RepoStruct)SetHandler(from func(field string)(string, bool), keys...string) {
	if r == nil {
		return
	}

	for _, key := range keys {
		if v, found := from(key) ; found {
			r.Set(key, v)
		}
	}
}

func (r *RepoStruct)SetApp(appRelName, appName string) (_ bool) {
	if r == nil {
		return
	}
	if r.forge == nil {
		return
	}
	if r.apps == nil {
		r.apps = make(map[string]*AppStruct)
	}

	if v, found := r.forge.Apps[appName] ; found {
		r.apps[appRelName] = v
		return true
	}
	return
}

func (r *RepoStruct)LoadRelApps() (_ error) {
	if r.forge == nil {
		return fmt.Errorf("Internal issue: %s", "Forge reference is missing.")
	}
	for relAppName, appName := range r.Apps {
		if ! r.SetApp(relAppName, appName) {
			return fmt.Errorf("Application '%s' not defined.", appName)
		}
	}

	// set from Defaults
	for relAppName, appName := range r.forge.ForjSettings.RepoApps {
		if v1, found1 := r.forge.Apps[appName]; !found1 {
			return fmt.Errorf("Default repo app: Application '%s' not defined.", appName)
		} else if _, found2 := r.apps[relAppName] ; ! found2 {
			r.apps[relAppName] = v1
		}

	}

	return
}

func (r *RepoStruct)Set(field, value string) {
	if r == nil {
		return
	}
	switch field {
	case "name":
		if r.name != value {
			r.name = value
		}
	case "upstream":
		if r.Upstream != value {
			r.Upstream = value
			r.forge.dirty()
		}
	case "git-remote":
		if r.GitRemote != value {
			r.GitRemote = value
			r.forge.dirty()
		}
	case "remote":
		if r.remote.Ssh != value {
			r.remote.Ssh = value
		}
	case "remote-url":
		if r.remote.Url != value {
			r.remote.Url = value
		}
	case "repo-template":
		if r.RepoTemplate != value {
			r.RepoTemplate = value
			r.forge.dirty()
		}
	case "title":
		if r.Title != value {
			r.Title = value
			r.forge.dirty()
		}
	case "flow":
		if r.Flow.Name != value {
			r.Flow.Name = value
			r.forge.dirty()
		}
	default:
		if r.More == nil {
			r.More = make(map[string]string)
		}
		if v, found := r.More[field] ; found && value == "" {
			delete(r.More,field)
			r.forge.dirty()
		} else {
			if v != value {
				r.forge.dirty()
				r.More[field] = value
			}

		}
	}
}

func (r *RepoStruct)SetInstanceOwner(owner string) {
	if r == nil {
		return
	}
	r.owner = owner
}

func (r *RepoStruct)SetPluginOwner(d *drivers.Driver) {
	if r == nil {
		return
	}
	r.driverOwner = d
}

func (r *RepoStruct)RemoteType() string {
	if r == nil {
		return "git"
	}
	if r.driverOwner == nil {
		return "git"
	}
	return r.driverOwner.Name
}

func (r *RepoStruct)UpstreamAPIUrl() string {
	if r == nil {
		return ""
	}
	if r.driverOwner == nil {
		return ""
	}
	return r.driverOwner.DriverAPIUrl
}

func (r *RepoStruct)set_forge(f *ForgeYaml) {
	if r == nil {
		return
	}

	r.forge = f
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
func (r *RepoStruct)HasApps(rules ...string) (found bool, err error) {
	if r.apps == nil {
		return
	}
	for _, app:= range r.apps {
		found = true
		for _, rule := range rules {
			ruleToCheck := strings.Split(rule, ":")
			if len(ruleToCheck) != 2 {
				err = fmt.Errorf("rule '%s' is invalid. Format supported is '<key>:<value>'.", rule)
				return
			}
			if v, found2 := app.Get(ruleToCheck[0]); found2 && v.GetString() != ruleToCheck[1] {
				found = false
				break
			}
		}
		if found {
			return
		}
	}
	return
}

func (r *RepoStruct)GetApps(rules ...string) (apps map[string]*AppStruct , err error) {
	if r.apps == nil {
		return
	}
	apps = make(map[string]*AppStruct)
	found := false
	for app_name, app:= range r.apps {
		found = true
		for _, rule := range rules {
			ruleToCheck := strings.Split(rule, ":")
			if len(ruleToCheck) != 2 {
				err = fmt.Errorf("rule '%s' is invalid. Format supported is '<key>:<value>'.", rule)
				return
			}
			if v, found2 := app.Get(ruleToCheck[0]); found2 && v.GetString() != ruleToCheck[1] {
				found = false
				break
			}
		}
		if found {
			apps[app_name] = app
		}
	}
	return
}

func (r *RepoStruct)HasValues(rules ...string) (found bool, err error) {
	found = true
	for _, rule := range rules {
		ruleToCheck := strings.Split(rule, ":")
		if len(ruleToCheck) != 2 {
			err = fmt.Errorf("rule '%s' is invalid. Format supported is '<key>:<value>'.", rule)
			return
		}
		if v, found2 := r.Get(ruleToCheck[0]); found2 && v.GetString() != ruleToCheck[1] {
			found = false
			break
		}
	}
	if found {
		return
	}
	return
}
