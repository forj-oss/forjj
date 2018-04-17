package forjfile

import (
	"fmt"
	"forjj/utils"
	"io/ioutil"
	"os"
	"path"

	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
	"gopkg.in/yaml.v2"
)

// ForjfileTmpl is the Memory expansion of an external Forjfile (used to create a Forge)
type ForjfileTmpl struct {
	file_loaded string
	Workspace   WorkspaceStruct // See workspace.go
	yaml        ForgeYaml
}

// Forge is the Memory expand of a repository Forjfile.
type Forge struct {
	file_loaded      string
	deployFileLoaded string
	tmplfile_loaded  string
	updated_msg      string
	infra_path       string // Infra path used to create/save/load Forjfile
	file_name        string // Relative path to the Forjfile.
	yaml             *ForgeYaml
}

// ForgeYaml represents the master Forjfile or a piece of the Forjfile template.
type ForgeYaml struct {
	updated     bool
	Deployments map[string]*DeploymentStruct
	ForjCore    DeployForgeYaml `yaml:",inline"`
}

// WorkspaceStruct represents the yaml structure of a workspace.
type WorkspaceStruct struct {
	updated                bool
	DockerBinPath          string            `yaml:"docker-exe-path"`    // Docker static binary path
	Contrib_repo_path      string            `yaml:"contribs-repo"`      // Contrib Repo path used.
	Flow_repo_path         string            `yaml:"flows-repo"`         // Flow repo path used.
	Repotemplate_repo_path string            `yaml:"repotemplates-repo"` // Repotemplate Path used.
	More                   map[string]string `yaml:",inline"`
}

const forjfileName = "Forjfile"

// TODO: Load multiple templates that will be merged.

// LoadTmpl Search for Forjfile in `aPath` and load it.
// This file combines the Forjfile in the infra repository and the Workspace
func LoadTmpl(aPath string) (f *ForjfileTmpl, loaded bool, err error) {
	var (
		yaml_data []byte
	)

	var forj_path string
	forj_path, err = utils.Abs(aPath)
	if err != nil {
		return
	}
	if forj_path != "." {
		if fi, e := os.Stat(forj_path); e != nil {
			err = e
			return
		} else {
			if !fi.Mode().IsDir() {
				return f, loaded, fmt.Errorf("'%s' must be a path to '%s'.", aPath, forjfileName)
			}
		}
	}

	file := path.Join(forj_path, forjfileName)

	if _, e := os.Stat(file); os.IsNotExist(e) {
		return
	}

	if fi, d, e := loadFile(file); e != nil {
		err = e
		return
	} else {
		yaml_data = d
		file = fi
	}

	f = new(ForjfileTmpl)

	f.file_loaded = file
	if e := yaml.Unmarshal(yaml_data, &f.yaml); e != nil {
		err = fmt.Errorf("Unable to load %s. %s", file, e)
		return
	}

	f.Workspace = f.yaml.ForjCore.LocalSettings
	gotrace.Trace("Forjfile template '%s' has been loaded.", file)
	// Setting defaults
	f.yaml.set_defaults()
	loaded = true
	return
}

func (f *Forge) GetForjfileTemplateFileLoaded() string {
	return f.tmplfile_loaded
}

func (f *Forge) GetForjfileFileLoaded() string {
	return f.file_loaded
}

func (f *Forge) SetInfraAsRepo() {
	// Copy the infra repo in list of repositories, tagged as infra.
	if !f.Init() {
		return
	}

	var repo *RepoStruct

	if v, found := f.yaml.ForjCore.Infra.More["name"]; found && v != "" {
		f.yaml.ForjCore.Infra.name = v
	}

	if f.yaml.ForjCore.Infra.name == "" || f.yaml.ForjCore.Infra.name == "none" {
		return
	}

	if r, found_repo := f.yaml.ForjCore.Repos[f.yaml.ForjCore.Infra.name]; found_repo {
		repo = r
	}
	if repo == nil {
		repo = new(RepoStruct)
		f.yaml.ForjCore.Repos[f.yaml.ForjCore.Infra.name] = repo
	}
	repo.setFromInfra(f.yaml.ForjCore.Infra)
}

func (f *Forge) GetInfraName() string {
	return f.yaml.ForjCore.Infra.name
}

// Load : Load Forjfile stored in a Repository.
func (f *Forge) Load(deployTo string) (loaded bool, err error) {
	var (
		yaml_data []byte
		file      string
	)

	if !f.Init() {
		return false, fmt.Errorf("Forge is nil.")
	}

	if f.infra_path != "" {
		if _, err = os.Stat(f.infra_path); err != nil {
			return
		}
	}

	aPath := path.Join(f.infra_path, f.Forjfile_name())
	if fi, d, e := loadFile(aPath); e != nil {
		err = e
		return
	} else {
		yaml_data = d
		file = fi
	}

	f.file_loaded = aPath
	if e := yaml.Unmarshal(yaml_data, &f.yaml); e != nil {
		err = fmt.Errorf("Unable to load %s. %s", file, e)
		return
	}

	f.yaml.set_defaults()
	loaded = true
	if deployTo != "" {
		f.SetDeployment(deployTo)
	} else {
		gotrace.Trace("Forge loaded from '%s'.", aPath)
	}
	msg := "' and '" + aPath

	if deployTo == "" {
		return
	}
	loaded = false
	// Loading Deployment forjfile
	aPath = path.Join(f.infra_path, "deployments", deployTo, f.Forjfile_name())
	if fi, d, e := loadFile(aPath); e != nil {
		err = e
		return
	} else {
		yaml_data = d
		file = fi
	}

	f.deployFileLoaded = aPath
	var deployData DeployForgeYaml

	if e := yaml.Unmarshal(yaml_data, &deployData); e != nil {
		err = fmt.Errorf("Unable to load deployment file '%s'. %s", file, e)
		return
	}

	err = f.yaml.ForjCore.mergeFrom(&deployData)
	if err != nil {
		return false, fmt.Errorf("Unable to load the Deployment forjfile. %s", err)
	}

	f.yaml.set_defaults()
	loaded = true
	gotrace.Trace("%s deployment forge loaded from '%s'.", deployTo, aPath+msg)

	return
}

func (f *Forge) DeployForjfile() *DeployForgeYaml {
	return &f.yaml.ForjCore
}

func loadFile(aPath string) (file string, yaml_data []byte, err error) {
	file, err = utils.Abs(aPath)
	if err != nil {
		return
	}

	if fi, e := os.Stat(file); e == nil {
		if !fi.Mode().IsRegular() {
			err = fmt.Errorf("%s must be a regular file.", file)
			return
		}
	} else {
		err = e
		return
	}

	if fd, e := ioutil.ReadFile(file); e != nil {
		err = e
		return
	} else {
		yaml_data = fd
	}
	return
}

func (f *Forge) SetFromTemplate(ft *ForjfileTmpl) {
	if !f.Init() {
		return
	}

	*f.yaml = ft.yaml
	f.yaml.updated = true
	f.tmplfile_loaded = ft.file_loaded
}

func (f *Forge) Init() bool {
	if f == nil {
		return false
	}
	if f.yaml == nil {
		f.yaml = new(ForgeYaml)
	}
	if f.yaml.Deployments == nil {
		f.yaml.Deployments = make(map[string]*DeploymentStruct)
	}

	return f.yaml.ForjCore.Init(f.yaml)
}

// CheckInfraPath will check if:
// - a Forjfile is found
// - is stored in a repository in root path.
func (f *Forge) CheckInfraPath(infraAbsPath string) error {
	if fi, err := os.Stat(infraAbsPath); err != nil {
		return fmt.Errorf("Not a valid infra path '%s': %s", infraAbsPath, err)
	} else if !fi.IsDir() {
		return fmt.Errorf("Not a valid infra path: '%s' must be a directory", infraAbsPath)
	}

	git := path.Join(infraAbsPath, ".git")
	if fi, err := os.Stat(git); err != nil {
		return fmt.Errorf("Not a valid infra path '%s'. Must be a GIT repository: %s", infraAbsPath, err)
	} else if !fi.IsDir() {
		return fmt.Errorf("Not a valid infra path: '%s' must be a directory", git)
	}

	forjfile := path.Join(infraAbsPath, forjfileName)
	if fi, err := os.Stat(forjfile); err != nil {
		return fmt.Errorf("Not a valid infra path '%s'. Must have a Forjfile: %s", infraAbsPath, err)
	} else if fi.IsDir() {
		return fmt.Errorf("Not a valid infra path: '%s' must be a file", forjfile)
	}

	return nil
}

func (f *Forge) SetInfraPath(infraPath string, create_request bool) error {
	aPath, err := utils.Abs(infraPath)
	if err != nil {
		return err
	}

	if create_request {
		if fi, err := os.Stat(aPath); err == nil && !fi.IsDir() {
			return fmt.Errorf("Unable to set infra PATH to '%s'. Must be a directory", aPath)
		}
	} else {
		if err := f.CheckInfraPath(aPath); err != nil {
			return err
		}
	}

	f.infra_path = aPath
	f.file_name = forjfileName // By default on Repo root directory.
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

func (f *Forge) Forjfiles_name() (ret []string) {

	ret = make([]string, 1, 1+len(f.yaml.Deployments))
	ret[0] = f.file_name
	for name := range f.yaml.Deployments {
		ret = append(ret, path.Join("deployments", name))
	}
	return
}

func (f *Forge) Save() error {
	if err := f.save(f.infra_path); err != nil {
		return err
	}
	f.Saved()
	return nil
}

func (f *Forge) save(infraPath string) error {
	if !f.Init() {
		return fmt.Errorf("Forge is nil")
	}

	file := path.Join(infraPath, f.Forjfile_name())
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
	if f.yaml.ForjCore.ForjSettings.is_template {
		return nil
	}
	for name, deployTo := range f.yaml.Deployments {
		filepath := path.Join(infraPath, "deployments", name)

		file = path.Join(infraPath, "deployments", name, f.Forjfile_name())

		yaml_data, err = yaml.Marshal(deployTo.Details)
		if err != nil {
			return err
		}

		if fi, err := os.Stat(filepath); err != nil || !fi.IsDir() {
			if err != nil {
				if err = os.MkdirAll(filepath, 0755); err != nil {
					return fmt.Errorf("Unable to create '%s'. %s", filepath, err)
				}
				gotrace.Trace("'%s' created.", filepath)
			} else {
				return fmt.Errorf("Unable to create '%s'. Already exist but is not a directory", filepath)
			}
		}

		if err := ioutil.WriteFile(file, yaml_data, 0644); err != nil {
			return err
		}
		gotrace.Trace("Deployment file name saved: %s", file)

	}

	return nil
}

// SaveTmpl provide Forjfile template export from a Forge.
func SaveTmpl(aPath string, f *Forge) error {
	forge := new(Forge)
	*forge = *f
	forge.yaml.ForjCore.ForjSettings.is_template = true
	return forge.save(aPath)
}

func (f *Forge) GetInstances(object string) (ret []string) {
	if !f.Init() {
		return nil
	}
	ret = []string{}
	switch object {
	case "deployment":
		if f.yaml.Deployments == nil {
			return
		}
		ret = make([]string, len(f.yaml.Deployments))
		iCount := 0
		for deployment := range f.yaml.Deployments {
			ret[iCount] = deployment
			iCount++
		}
		return

	case "user":
		if f.yaml.ForjCore.Users == nil {
			return
		}

		ret = make([]string, len(f.yaml.ForjCore.Users))
		iCount := 0
		for user := range f.yaml.ForjCore.Users {
			ret[iCount] = user
			iCount++
		}
		return
	case "group":
		if f.yaml.ForjCore.Groups == nil {
			return
		}

		ret = make([]string, len(f.yaml.ForjCore.Groups))
		iCount := 0
		for group := range f.yaml.ForjCore.Groups {
			ret[iCount] = group
			iCount++
		}
		return
	case "app":
		if f.yaml.ForjCore.Apps == nil {
			return
		}

		ret = make([]string, len(f.yaml.ForjCore.Apps))
		iCount := 0
		for app := range f.yaml.ForjCore.Apps {
			ret[iCount] = app
			iCount++
		}
		return
	case "repo":
		if f.yaml.ForjCore.Repos == nil {
			return
		}

		ret = make([]string, len(f.yaml.ForjCore.Repos))
		iCount := 0
		for repo := range f.yaml.ForjCore.Repos {
			ret[iCount] = repo
			iCount++
		}
		return
	case "infra", "settings":
		return
	default:
		if instances, found := f.yaml.ForjCore.More[object]; found {
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

func (f *Forge) GetInfraInstance() (_ string) {
	if f == nil || f.yaml.ForjCore.Infra == nil || f.yaml.ForjCore.Infra.apps == nil {
		return
	}
	if v, found := f.yaml.ForjCore.Infra.apps["upstream"]; found && v != nil {
		return v.name
	}
	if v, found := f.yaml.ForjCore.Infra.Apps["upstream"]; found {
		return v
	}
	return f.yaml.ForjCore.Infra.Upstream
}

func (f *Forge) GetString(object, instance, key string) (string, bool) {
	v, found := f.Get(object, instance, key)
	return v.GetString(), found
}

func (f *Forge) Get(object, instance, key string) (value *goforjj.ValueStruct, _ bool) {
	if !f.Init() {
		return
	}
	return f.yaml.ForjCore.Get(object, instance, key)
}

func (f *Forge) GetObjectInstance(object, instance string) interface{} {
	if !f.Init() {
		return nil
	}
	switch object {
	case "user":
		if f.yaml.ForjCore.Users == nil {
			return nil
		}
		if user, found := f.yaml.ForjCore.Users[instance]; found {
			return user
		}
	case "group":
		if f.yaml.ForjCore.Groups == nil {
			return nil
		}
		if group, found := f.yaml.ForjCore.Groups[instance]; found {
			return group
		}
	case "app":
		if f.yaml.ForjCore.Apps == nil {
			return nil
		}
		if app, found := f.yaml.ForjCore.Apps[instance]; found {
			return app
		}
	case "repo":
		if f.yaml.ForjCore.Repos == nil {
			return nil
		}
		if repo, found := f.yaml.ForjCore.Repos[instance]; found {
			return repo
		}
	case "settings":
		return f.yaml.ForjCore.ForjSettings.GetInstance(instance)
	default:
		return f.getInstance(object, instance)
	}
	return nil
}

func (f *Forge) ObjectLen(object string) int {
	if !f.Init() {
		return 0
	}
	switch object {
	case "infra":
		return 1
	case "user":
		if f.yaml.ForjCore.Users == nil {
			return 0
		}
		return len(f.yaml.ForjCore.Users)
	case "group":
		if f.yaml.ForjCore.Groups == nil {
			return 0
		}
		return len(f.yaml.ForjCore.Groups)
	case "app":
		if f.yaml.ForjCore.Apps == nil {
			return 0
		}
		return len(f.yaml.ForjCore.Apps)
	case "repo":
		if f.yaml.ForjCore.Repos == nil {
			return 0
		}
		return len(f.yaml.ForjCore.Repos)
	case "settings":
		return 1
	default:
		if v, found := f.yaml.ForjCore.More[object]; found {
			return len(v)
		}
		return 0
	}
	return 0
}

func (f *Forge) getInstance(object, instance string) (_ map[string]ForjValue) {
	if !f.Init() {
		return
	}
	if obj, f1 := f.yaml.ForjCore.More[object]; f1 {
		if i, f2 := obj[instance]; f2 {
			return i
		}
	}
	return
}

func (f *Forge) Remove(object, name, key string) {
	from := func(string) (_ string, _ bool) {
		return "", true
	}

	f.yaml.ForjCore.SetHandler(object, name, from, (*ForjValue).Clean, key)
}

func (f *Forge) Set(object, name, key, value string) {
	from := func(string) (string, bool) {
		return value, (value != "")
	}
	f.yaml.ForjCore.SetHandler(object, name, from, (*ForjValue).Set, key)
}

func (f *Forge) SetDefault(object, name, key, value string) {
	from := func(string) (string, bool) {
		return value, (value != "")
	}
	f.yaml.ForjCore.SetHandler(object, name, from, (*ForjValue).SetDefault, key)
}

func (f *Forge) IsDirty() bool {
	if !f.Init() {
		return false
	}

	return f.yaml.updated
}

func (f *Forge) Saved() {
	if !f.Init() {
		return
	}

	f.yaml.updated = false
}

func (f *Forge) Apps() map[string]*AppStruct {
	if !f.Init() {
		return nil
	}

	return f.yaml.ForjCore.Apps
}

// Initialize the forge. (Forjfile in repository infra)
func (f *ForgeYaml) Init() {
	if f.ForjCore.Groups == nil {
		f.ForjCore.Groups = make(map[string]*GroupStruct)
	}
	if f.ForjCore.Users == nil {
		f.ForjCore.Users = make(map[string]*UserStruct)
	}
	if f.ForjCore.More == nil {
		f.ForjCore.More = make(map[string]map[string]ForjValues)
	}

	if f.ForjCore.Infra.More == nil {
		f.ForjCore.Infra.More = make(map[string]string)
	}

	if f.ForjCore.Repos == nil {
		f.ForjCore.Repos = make(map[string]*RepoStruct)
	}

	if f.ForjCore.Apps == nil {
		f.ForjCore.Apps = make(map[string]*AppStruct)
	}

	if f.Deployments == nil {
		f.Deployments = make(map[string]*DeploymentStruct)
	}

}

func (f *ForgeYaml) set_defaults() {
	// Cleanup LocalSettings to ensure no local setting remain in a Forjfile
	f.ForjCore.LocalSettings = WorkspaceStruct{}

	if f.ForjCore.Apps != nil {
		for name, app := range f.ForjCore.Apps {
			if app == nil {
				continue
			}
			app.name = name
			if app.Driver == "" {
				app.Driver = name
			}
			app.set_forge(f)
			f.ForjCore.Apps[name] = app
		}
	}
	if f.ForjCore.Repos != nil {
		for name, repo := range f.ForjCore.Repos {
			if repo == nil {
				// Repo can be nil if we did not defined any fields under his name.
				// ie : forjj-modules:
				// or
				// forjj-modules: nil
				repo = new(RepoStruct)
			}
			repo.name = name
			repo.set_forge(f)
			f.ForjCore.Repos[name] = repo
		}
	}
	if f.ForjCore.Users != nil {
		for name, user := range f.ForjCore.Users {
			if user == nil {
				continue
			}
			user.set_forge(f)
			f.ForjCore.Users[name] = user
		}
	}
	if f.ForjCore.Groups != nil {
		for name, group := range f.ForjCore.Groups {
			if group == nil {
				continue
			}
			group.set_forge(f)
			f.ForjCore.Groups[name] = group
		}
	}
	if f.ForjCore.Infra == nil {
		f.ForjCore.Infra = new(RepoStruct)
	}
	f.ForjCore.Infra.set_forge(f)
	f.ForjCore.ForjSettings.set_forge(f)
	if len(f.Deployments) == 0 {
		data := DeploymentStruct{}
		data.Desc = "Production environment"
		data.name = "production"
		data.Type = "PRO"
		f.Deployments = make(map[string]*DeploymentStruct)
		f.Deployments[data.name] = &data
		gotrace.Info("No deployment defined. Created single 'production' deployment. If you want to change that update your forjfile and create a deployment Forfile per deployment under 'deployments/<deploymentName>'.")
		f.ForjCore.ForjSettings.DeployTo = data.name
	} else {
		for name, deploy := range f.Deployments {
			deploy.name = name
			f.Deployments[name] = deploy
		}
	}
}

func (f *ForgeYaml) dirty() {
	f.updated = true
}

func (f *Forge) GetDeclaredFlows() (result []string) {
	flows := make(map[string]bool)

	for _, repo := range f.yaml.ForjCore.Repos {
		if repo.Flow.Name != "" {
			flows[repo.Flow.Name] = true
		}
	}
	if flow := f.yaml.ForjCore.ForjSettings.Default.getFlow(); flow != "" {
		flows[flow] = true
	}

	if len(flows) == 0 {
		flows["default"] = true // Default is always loaded when nothing is declared.
	}

	result = make([]string, 0, len(flows))
	for name := range flows {
		result = append(result, name)
	}
	return
}

func (f *Forge) Model() ForgeModel {
	model := ForgeModel{
		forge: f,
	}

	return model
}

func (f *Forge) GetDeployment() string {
	return f.yaml.ForjCore.ForjSettings.DeployTo
}

func (f *Forge) SetDeployment(deployTo string) {
	f.yaml.ForjCore.ForjSettings.DeployTo = deployTo
}

// GetADeployment return the Deployment Object wanted
func (f *Forge) GetADeployment(deploy string) (v *DeploymentStruct, found bool) {
	v, found = f.yaml.Deployments[deploy]
	return
}


// Validate check if the information in the Forjfile are coherent or not and if code respect some basic rules.
func (f *Forge) Validate() error {

	// ForjSettingsStruct.More

	// RepoStruct.More (infra : Repos)

	// AppYamlStruct.More

	// Repository apps connection
	for _, repo := range f.yaml.ForjCore.Repos {
		if repo.Apps == nil {
			continue
		}

		for relAppName, appName := range repo.Apps {
			if _, err := repo.SetInternalRelApp(relAppName, appName); err != nil {
				return fmt.Errorf("Repo '%s' has an invalid Application reference '%s: %s'. %s", repo.GetString("name"), relAppName, appName, err)
			}
		}
	}

	// UserStruct.More

	// GroupStruct.More

	// ForgeYaml.More

	// DeploymentStruct
	pro := false
	for name, deploy := range f.yaml.Deployments {
		if deploy.Type == "" {
			return fmt.Errorf("Deployment declaration error in '%s'. Missing type. Provide at least `Type: (PRO|TEST|DEV)`", name)
		}
		if deploy.Type == "PRO" && pro {
			return fmt.Errorf("Deployment declaration error in '%s'. You cannot have more than 1 deployment of type 'PRO'. Please fix it.", name)
		}
	}

	return nil
}

func (f *Forge) GetDeployments() (result map[string]*DeploymentStruct) {
	result = f.yaml.Deployments
	return
}

// GetDeploymentType return the first 
func (f *Forge) GetDeploymentType(deployType string) (v map[string]*DeploymentStruct, found bool) {
	v = make(map[string]*DeploymentStruct)

	for name, deploy := range f.yaml.Deployments {
		if deploy.Type == deployType {
			v[name] = deploy
			found = true
		}
	}
	return
}


// GetDeploymentType return the first 
func (f *Forge) GetDeploymentPROType() (v *DeploymentStruct, err error) {

	if deployObjs, _ := f.GetDeploymentType("PRO") ; len(deployObjs) != 1 {
		err = fmt.Errorf("Found more than one PRO environment")
	} else {
		for k := range deployObjs {
			v = deployObjs[k]
			break
		}
	}

	return
}
