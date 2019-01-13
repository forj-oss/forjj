package forjfile

import (
	"fmt"
	"forjj/drivers"
	"forjj/sources_info"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

type RepoStruct struct {
	name            string
	is_infra        bool
	isDeploy        bool
	isCurrentDeploy bool
	forge           *ForgeYaml
	owner           string
	driverOwner     *drivers.Driver
	deployment      string                      // set to deploiment name where the repo is planned to be deployed.
	Deployment      string                      `yaml:"deploy-repo-of,omitempty"`
	Upstream        string                      `yaml:"upstream-app,omitempty"` // Name of the application upstream hosting this repository.
	GitRemote       string                      `yaml:"git-remote,omitempty"`
	remote          goforjj.PluginRepoRemoteUrl // Git remote string to use/set
	Title           string                      `yaml:",omitempty"`
	RepoTemplate    string                      `yaml:"repo-template,omitempty"`
	Flow            RepoFlow                    `yaml:",omitempty"`
	More            map[string]string           `yaml:",inline"`
	apps            map[string]*AppStruct       // List of applications connected to this repo. Defaults are added automatically.
	Apps            map[string]string           `yaml:"in-relation-with,omitempty"` // key: <AppRelName>, value: <appName>
	sources         *sourcesinfo.Sources
}

const (
	FieldRepoName          = "name"
	FieldRepoUpstream      = "upstream"
	FieldRepoApps          = "apps"
	FieldRepoGitRemote     = "git-remote"
	FieldRepoRemote        = "remote"
	FieldRepoRemoteURL     = "remote-url"
	FieldRepoTitle         = "title"
	FieldRepoFlow          = "flow"
	FieldRepoTemplate      = "repo-template"
	FieldRepoDeployName    = "deployment-name"
	FieldRepoDeployType    = "deployment-type"
	FieldRepoRole          = "role"
	FieldCurrentDeployRepo = "current-deployment-repo"
)

// Register will register the repository
func (r *RepoStruct) Register() {
	if r == nil {
		return
	}
	if r.forge == nil {
		return
	}
	r.forge.ForjCore.Repos[r.name] = r
}

// Flags provide the list of existing keys in a repo object.
func (r *RepoStruct) Flags() (flags []string) {
	if r == nil {
		flags = make([]string, 0)
		return
	}
	coreList := []string{
		FieldRepoName,
		FieldRepoUpstream,
		FieldRepoGitRemote,
		FieldRepoRemote,
		FieldRepoRemoteURL,
		FieldRepoTitle,
		FieldRepoFlow,
		FieldRepoTemplate,
		FieldRepoDeployName,
		FieldRepoRole,
		FieldCurrentDeployRepo,
	}
	flags = make([]string, len(coreList), len(coreList)+len(r.More))
	for index, name := range coreList {
		flags[index] = name
	}
	for k := range r.More {
		flags = append(flags, k)
	}
	return

}

func (r *RepoStruct) mergeFrom(from *RepoStruct) *RepoStruct {
	if r == nil {
		return from
	}
	for _, flag := range from.Flags() {
		if v, found, source := from.Get(flag); found {
			r.Set(source, flag, v.GetString())
		}
	}
	return r
}

type RepoFlow struct {
	Name    string
	objects map[string]map[string]string
}

func (r *RepoStruct) Model() RepoModel {
	model := RepoModel{
		Apps: make(map[string]RepoAppModel),
	}

	model.From(r)

	if r == nil {
		return model
	}

	for appRelName, app := range r.apps {
		repoModel := RepoAppModel{
			AppName: app.name,
		}
		if _, found := r.Apps[appRelName]; !found {
			repoModel.Default = true
		}
		model.Apps[appRelName] = repoModel
	}
	return model
}

func (r *RepoStruct) Owner() string {
	if r == nil {
		return ""
	}

	return r.owner
}

func (r *RepoStruct) setFromInfra(infra *RepoStruct) {
	if r == nil {
		return
	}
	infra.is_infra = true
	infra.isDeploy = false
	*r = *infra
	delete(r.More, "name")
}

func (r *RepoStruct) GetString(field string) (_ string, _ string) {
	if r == nil {
		return
	}

	if v, found, source := r.Get(field); found {
		return v.GetString(), source
	}
	return
}

func (r *RepoStruct) RemoteUrl() string {
	if r == nil {
		return ""
	}

	return r.remote.Url
}

func (r *RepoStruct) RemoteGit() string {
	if r == nil {
		return ""
	}

	return r.remote.Ssh
}

// Get return the value for a field.
func (r *RepoStruct) Get(field string) (value *goforjj.ValueStruct, found bool, source string) {
	if r == nil {
		return
	}
	source = r.sources.Get(field)
	switch fieldSel := strings.Split(field, ":"); fieldSel[0] {
	case FieldRepoName:
		value, found = value.SetIfFound(r.name, (r.name != ""))
	case FieldRepoApps, FieldRepoUpstream:
		if field == "upstream" {
			gotrace.Warning("*RepoStruct.Get(): Field '%s' is obsolete. Change the code to use 'apps:upstream'.", field)
		} else if len(fieldSel) > 1 {
			field = fieldSel[1]
		}
		if v, found2 := r.apps[field]; found2 {
			value, found = value.SetIfFound(v.name, v != nil && v.name != "")
			return
		}
		if v, found2 := r.Apps[field]; found2 {
			value, found = value.SetIfFound(v, found2 && (v != ""))
			return
		}
		if field == "upstream" && r.Upstream != "" {
			gotrace.Warning("the '%s' in /repositories/%s is obsolete. Define it as /repositories/%s/apps/%s", field,
				r.name, r.name, field)
			value, found = value.SetIfFound(r.Upstream, (r.Upstream != ""))
			return
		}
		value, found = value.SetIfFound("", false)
	case FieldRepoGitRemote:
		value, found = value.SetIfFound(r.GitRemote, (r.GitRemote != ""))
	case FieldRepoRemote:
		value, found = value.SetIfFound(r.remote.Ssh, (r.remote.Ssh != ""))
	case FieldRepoRemoteURL:
		value, found = value.SetIfFound(r.remote.Url, (r.remote.Url != ""))
	case FieldRepoTitle:
		value, found = value.SetIfFound(r.Title, (r.Title != ""))
	case FieldRepoFlow:
		value, found = value.SetIfFound(r.Flow.Name, (r.Flow.Name != ""))
	case FieldRepoTemplate:
		value, found = value.SetIfFound(r.RepoTemplate, (r.RepoTemplate != ""))
	case FieldRepoDeployName:
		value, found = value.SetIfFound(r.deployment, (r.deployment != ""))
	case FieldRepoDeployType:
		if v, found2 := r.forge.Deployments[r.deployment]; found2 {
			value, found = value.SetIfFound(v.Type, found2)
			return
		}
		value, found = value.SetIfFound("", false)
	case FieldRepoRole:
		value, found = value.SetIfFound(r.Role(), true)
	case FieldCurrentDeployRepo:
		if r.isCurrentDeploy {
			value, found = value.SetIfFound("true", true)
			return
		}
		value, found = value.SetIfFound("false", true)
	default:
		v, f := r.More[field]
		value, found = value.SetIfFound(v, f)
	}
	return
}

// SetHandler is the generic setter function.
func (r *RepoStruct) SetHandler(source string, from func(field string) (string, bool), keys ...string) {
	if r == nil {
		return
	}

	for _, key := range keys {
		if v, found := from(key); found {
			r.Set(source, key, v)
		}
	}
}

// SetInternalRelApp will set the appName connected to the repo
// But if the Forjfile has a setup, this one will forcefully used.
// return:
// updated bool : true if the app has been updated.
// error : set if error has been found. updated is then nil.
func (r *RepoStruct) SetInternalRelApp(appRelName, appName string) (updated *bool, _ error) {
	if err := r.initApp(); err != nil {
		return nil, err
	}

	if v, found := r.Apps[appRelName]; found {
		appName = v // Set always declared one.
	}

	if app, err := r.forge.ForjCore.Apps.Found(appName); err != nil {
		return nil, fmt.Errorf("Unable to set %s:%s. %s.", appRelName, appName, err)
	} else if v, found := r.apps[appRelName]; !found || (found && v.name != appName) {
		r.apps[appRelName] = app
		updated = new(bool)
		*updated = true
	}

	return
}

// SetApp Define the Forjfile application name to link with the repo.
//
// return:
// updated bool : true if the app has been updated.
// error : set if error has been found. updated is then nil.
func (r *RepoStruct) SetApp(appRelName, appName string) (updated *bool, _ error) {
	if err := r.initApp(); err != nil {
		return nil, err
	}

	updated = new(bool)
	if v, found := r.Apps[appRelName]; !found || (found && v != appName) {
		*updated = true
	}
	r.Apps[appRelName] = appName

	if set, err := r.SetInternalRelApp(appRelName, appName); set == nil {
		return set, err
	}
	return
}

func (r *RepoStruct) initApp() (_ error) {
	if r == nil {
		return fmt.Errorf("Internal: repo object is nil.")
	}
	if r.forge == nil {
		return fmt.Errorf("Internal: forge ref not set.")
	}

	if r.apps == nil {
		r.apps = make(map[string]*AppStruct)
	}
	if r.Apps == nil {
		r.Apps = make(map[string]string)
	}
	return
}

func (r *RepoStruct) Set(source, field, value string) {
	if r == nil {
		return
	}
	switch field_sel := strings.Split(field, ":"); field_sel[0] {
	case FieldRepoName:
		if r.name != value {
			r.name = value
		}
	case FieldRepoApps:
		if len(field_sel) > 1 {
			r.SetApp(field_sel[1], value)
			r.forge.dirty()
		}
	case FieldRepoUpstream:
		if v, found := r.Apps["upstream"]; !found || (found && v != value) {
			r.SetApp("upstream", value)
			r.forge.dirty()
		}
	case FieldRepoGitRemote:
		if r.GitRemote != value {
			r.GitRemote = value
			r.forge.dirty()
		}
	case FieldRepoRemote:
		if r.remote.Ssh != value {
			r.remote.Ssh = value
		}
	case FieldRepoRemoteURL:
		if r.remote.Url != value {
			r.remote.Url = value
		}
	case FieldRepoTemplate:
		if r.RepoTemplate != value {
			r.RepoTemplate = value
			r.forge.dirty()
		}
	case FieldRepoTitle:
		if r.Title != value {
			r.Title = value
			r.forge.dirty()
		}
	case FieldRepoFlow:
		if r.Flow.Name != value {
			r.Flow.Name = value
			r.forge.dirty()
		}
	case FieldRepoDeployName:
		if r.deployment != value {
			r.deployment = value
		}
	case FieldRepoRole:
		switch value {
		case "infra":
			r.is_infra = true
			r.isDeploy = false
			r.isCurrentDeploy = false
		case "deploy":
			r.is_infra = false
			r.isDeploy = true
		case "code":
			r.is_infra = false
			r.isDeploy = false
			r.isCurrentDeploy = false
		}
	case FieldCurrentDeployRepo:
		r.SetCurrentDeploy()
	default:
		if r.More == nil {
			r.More = make(map[string]string)
		}
		if v, found := r.More[field]; found && value == "" {
			delete(r.More, field)
			r.forge.dirty()
		} else {
			if v != value {
				r.forge.dirty()
				r.More[field] = value
			}

		}
	}
	r.sources = r.sources.Set(source, field, value)
}

func (r *RepoStruct) SetInstanceOwner(owner string) {
	if r == nil {
		return
	}
	r.owner = owner
}

func (r *RepoStruct) SetPluginOwner(d *drivers.Driver) {
	if r == nil {
		return
	}
	r.driverOwner = d
}

func (r *RepoStruct) RemoteType() string {
	if r == nil {
		return "git"
	}
	if r.driverOwner == nil {
		return "git"
	}
	return r.driverOwner.Name
}

func (r *RepoStruct) UpstreamAPIUrl() string {
	if r == nil {
		return ""
	}
	if r.driverOwner == nil {
		return ""
	}
	return r.driverOwner.DriverAPIUrl
}

func (r *RepoStruct) set_forge(f *ForgeYaml) {
	if r == nil {
		return
	}

	r.forge = f
}

// HasApps return a bool if rules are all true on at least one application.
// a rule is a string formatted as '<key>:<value>' or '<key>:*' for any value.
// a rule is true on an application if it has the key value set to <value>
//
// If the rule is not well formatted, an error is returned.
// If the repo has no application, HasApps return false.
// If no rules are provided and at least one application exist, HasApps return true.
// If all rules are true, HasApps return true
//
// TODO: Write Unit test of HasApps
func (r *RepoStruct) HasApps(rules ...string) (found bool, err error) {
	if r.apps == nil {
		return
	}
	for appRelName, app := range r.apps {
		found = true
		for _, rule := range rules {
			ruleToCheck := strings.Split(rule, ":")
			if len(ruleToCheck) != 2 {
				err = fmt.Errorf("rule '%s' is invalid. Format supported is '<key>:<value>'.", rule)
				return
			}
			if ruleToCheck[0] == "appRelName" {
				if appRelName == ruleToCheck[0] {
					continue
				}
				found = false
				break
			}
			v, found2, _ := app.Get(ruleToCheck[0])
			if ruleToCheck[1] == "*" {
				if found2 {
					continue
				}
				found = false
				break
			} else if found2 {
				 if v.GetString() != ruleToCheck[1] {
					found = false
					break
				}
			} else {
				found = false
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

func (r *RepoStruct) GetApps(rules ...string) (apps map[string]*AppStruct, err error) {
	if r.apps == nil {
		return
	}
	apps = make(map[string]*AppStruct)
	found := false
	for app_name, app := range r.apps {
		found = true
		for _, rule := range rules {
			ruleToCheck := strings.Split(rule, ":")
			if len(ruleToCheck) != 2 {
				err = fmt.Errorf("rule '%s' is invalid. Format supported is '<key>:<value>'.", rule)
				return
			}
			if ruleToCheck[0] == "appRelName" {
				if app_name == ruleToCheck[0] {
					continue
				}
				found = false
				break
			}
			v, found2, _ := app.Get(ruleToCheck[0])
			if ruleToCheck[1] == "*" {
				if found2 {
					continue
				}
				found = false
				break
			} else if found2 && v.GetString() != ruleToCheck[1] {
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

func (r *RepoStruct) HasValues(rules ...string) (found bool, err error) {
	found = true
	for _, rule := range rules {
		ruleToCheck := strings.Split(rule, ":")
		if len(ruleToCheck) != 2 {
			err = fmt.Errorf("rule '%s' is invalid. Format supported is '<key>:<value>'.", rule)
			return
		}
		if v, found2, _ := r.Get(ruleToCheck[0]); found2 && v.GetString() != ruleToCheck[1] {
			found = false
			break
		}
	}
	if found {
		return
	}
	return
}

func (r *RepoStruct) IsInfra() bool {
	if r == nil {
		return false
	}
	return r.is_infra
}

func (r *RepoStruct) Role() string {
	switch {
	case r.is_infra:
		return "infra"
	case r.isDeploy:
		return "deploy"
	default:
		return "code"
	}
}

// Name return the internal reponame.
func (r *RepoStruct) Name() string {
	if r == nil {
		return ""
	}
	return r.name
}

// IsCurrentDeploy return True if this repo is the current deployment repository
func (r *RepoStruct) IsCurrentDeploy() (ret bool) {
	if r == nil {
		return
	}
	return r.isCurrentDeploy && r.isDeploy
}

// SetCurrentDeploy set True if this repo IS the current deployment repository
// No check made on other repositories.
// At a time, only one repository should be considered as the current deployment repository
func (r *RepoStruct) SetCurrentDeploy() {
	if r == nil {
		return
	}
	if !r.isDeploy {
		return
	}
	r.isCurrentDeploy = true
}

func (r *RepoStruct) AttachedToDeployment() string {
	return r.deployment
}
