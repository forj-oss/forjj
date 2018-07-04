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
	deployTo         string // Current deployment name used by forjj
	tmplfile_loaded  string
	updated_msg      string
	infra_path       string // Infra path used to create/save/load Forjfile
	file_name        string // Relative path to the Forjfile.
	yaml             *ForgeYaml
	inMem            *DeployForgeYaml
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

const (
	forjfileName   = "Forjfile"
	DevDeployType  = "DEV"
	testDeployType = "TEST"
	ProDeployType  = "PRO"
)

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
	// Setting internals and some predefined objects
	f.yaml.set_defaults()
	loaded = true

	// Setting default values found in Forjfile/forj-settings/default/...
	f.yaml.defineDefaults(false) // Do not warn if default are set.

	gotrace.Trace("Forjfile template '%s' has been loaded.", file)
	return
}

func (f *Forge) GetForjfileTemplateFileLoaded() string {
	return f.tmplfile_loaded
}

func (f *Forge) GetForjfileFileLoaded() string {
	return f.file_loaded
}

// selectCore do:
// - return inMem if defined
// - or return yaml if defined
// - or return nil
//
func (f *Forge) selectCore() (forge *DeployForgeYaml) {
	forge = f.InMemForjfile()
	if forge != nil {
		return
	}
	forge = f.DeployForjfile()
	return
}

// GetInfraRepo return the infra repository object if defined.
func (f *Forge) GetInfraRepo() *RepoStruct {
	forge := f.selectCore()
	if forge == nil {
		return nil
	}
	return forge.Infra
}

func (f *Forge) attachInfraToDeployment(deploy string) {
	infra := f.GetInfraRepo()
	if infra == nil {
		return
	}
	infra.deployment = deploy
}

func (f *Forge) SetInfraAsRepo() {
	// Copy the infra repo in list of repositories, tagged as infra.
	if !f.Init() {
		return
	}

	var repo *RepoStruct
	forge := f.selectCore()
	if forge == nil {
		return
	}

	if v, found := forge.Infra.More["name"]; found && v != "" {
		forge.Infra.name = v
	}

	if forge.Infra.name == "" || forge.Infra.name == "none" {
		return
	}

	if r, found_repo := forge.Repos[forge.Infra.name]; found_repo {
		repo = r
	}
	if repo == nil {
		repo = new(RepoStruct)
		forge.Repos[forge.Infra.name] = repo
	}
	repo.setFromInfra(forge.Infra)
}

func (f *Forge) GetInfraName() string {
	forge := f.selectCore()
	if forge == nil {
		return ""
	}
	return forge.Infra.name
}

// Load : Load Forjfile stored in a Repository.
func (f *Forge) Load(deployTo string) (loaded bool, err error) {
	if !f.Init() {
		return false, fmt.Errorf("Forge is nil.")
	}

	if f.infra_path != "" {
		if _, err = os.Stat(f.infra_path); err != nil {
			return
		}
	}

	aPath := path.Join(f.infra_path, f.Forjfile_name())

	// Loading master forjfile
	if _, err = f.load(&f.yaml, aPath); err != nil {
		return
	}

	// Loading All Deployments forjfiles
	for deployName, deployData := range f.yaml.Deployments {
		deployDetail := new(DeployForgeYaml)

		aPath = path.Join(f.infra_path, "deployments", deployName, f.Forjfile_name())

		if _, err = f.load(deployDetail, aPath); err != nil {
			return
		}
		deployDetail.Repos.attachToDeploy(deployName)
		if deployData.Type == "PRO" {
			f.attachInfraToDeployment(deployName)
		}
		deployData.Details = deployDetail
	}

	defer f.yaml.set_defaults()

	if deployTo == "global" { // We did not defined a deployment environment.
		f.SetDeployment(deployTo)
		return
	}

	if deployTo != "" {
		if _, found := f.yaml.Deployments.GetADeployment(deployTo); !found {
			err = fmt.Errorf("Deployment '%s' not defined", deployTo)
			return
		}

		gotrace.Info("Deployment environment selected: %s", deployTo)
	}

	if deployTo == "" { // if deploy was not requested, it will use the default dev-deploy if set by defineDefaults or Forjfile.
		if v, found, _ := f.Get("settings", "default", "dev-deploy"); found {
			deployTo = v.GetString()
			gotrace.Info("Default DEV deployment environment selected: %s", deployTo)
		}
	}
	if deployTo == "" {
		if deploys, _ := f.yaml.Deployments.GetDeploymentType("DEV"); len(deploys) == 1 {
			deployTo = deploys.One().Name()
			gotrace.Info("Single DEV deployment environment selected: %s", deployTo)
		}
	}

	if deployTo != "" {
		// Define current deployment loaded in memory.
		f.SetDeployment(deployTo)
	}

	f.yaml.defineDefaults(true) // Do warn if default are set to suggest updating the Forfile instead.

	loaded = true

	return
}

func (f *Forge) load(deployData interface{}, aPath string) (loaded bool, err error) {
	var (
		yaml_data []byte
		file      string
	)

	if fi, d, e := loadFile(aPath); e != nil {
		err = e
		return
	} else {
		yaml_data = d
		file = fi
	}

	f.deployFileLoaded = aPath

	if e := yaml.Unmarshal(yaml_data, deployData); e != nil {
		err = fmt.Errorf("Unable to load deployment file '%s'. %s", file, e)
		return
	}
	loaded = true

	f.yaml.set_defaults()
	gotrace.Trace("%s loaded.", aPath)
	return
}

// BuildForjfileInMem return a merge of the Master forjfile with the current deployment.
func (f *Forge) BuildForjfileInMem() (err error) {
	f.inMem, err = f.MergeFromDeployment(f.GetDeployment())
	return
}

// MergeFromDeployment provide a merge between Master and Deployment Forjfile.
func (f *Forge) MergeFromDeployment(deployTo string) (result *DeployForgeYaml, err error) {
	if f == nil {
		return nil, fmt.Errorf("Forge is nil")
	}
	deploy, found := f.GetADeployment(deployTo)
	if !found {
		return nil, fmt.Errorf("Unable to find deployment '%s'", deployTo)
	}
	forge := NewForgeYaml()

	// Keep list of deployment definition, but no details.
	for deployName, deploy := range f.yaml.Deployments {
		newDeploy := new(DeploymentStruct)
		newDeploy.DeploymentCoreStruct = deploy.DeploymentCoreStruct
		forge.Deployments[deployName] = newDeploy
	}

	result = &forge.ForjCore
	if err = result.mergeFrom(&f.yaml.ForjCore); err != nil {
		return nil, fmt.Errorf("Unable to load the master forjfile. %s", err)
	}
	if err = result.mergeFrom(deploy.Details); err != nil {
		return nil, fmt.Errorf("Unable to merge the Deployment forjfile to master one. %s", err)
	}
	result.deployTo = deployTo
	result.initDefaults(forge)
	return
}

// DeployForjfile return the Forjfile master object
func (f *Forge) DeployForjfile() *DeployForgeYaml {
	if f.yaml == nil {
		return nil
	}
	return &f.yaml.ForjCore
}

// DeployForjfile return the Forjfile master object
func (f *Forge) InMemForjfile() *DeployForgeYaml {
	return f.inMem
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

		if deployTo.Details == nil {
			gotrace.Warning("The %s deployment info is empty. Forjfile-template:/deployments/%s/define not defined. (%s)", name, name, file)
			deployTo.Details = new(DeployForgeYaml)
		}

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
	forge := f.selectCore()
	if forge == nil {
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
		if forge.Users == nil {
			return
		}

		ret = make([]string, len(forge.Users))
		iCount := 0
		for user := range forge.Users {
			ret[iCount] = user
			iCount++
		}
		return
	case "group":
		if forge.Groups == nil {
			return
		}

		ret = make([]string, len(forge.Groups))
		iCount := 0
		for group := range forge.Groups {
			ret[iCount] = group
			iCount++
		}
		return
	case "app":
		if forge.Apps == nil {
			return
		}

		ret = make([]string, len(forge.Apps))
		iCount := 0
		for app := range forge.Apps {
			ret[iCount] = app
			iCount++
		}
		return
	case "repo":
		if forge.Repos == nil {
			return
		}

		ret = make([]string, len(forge.Repos))
		iCount := 0
		for repo := range forge.Repos {
			ret[iCount] = repo
			iCount++
		}
		return
	case "infra", "settings":
		return
	default:
		if instances, found := forge.More[object]; found {
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
	if f == nil {
		return
	}
	forge := f.selectCore()
	if forge == nil {
		return
	}
	if forge.Infra.apps != nil {
		if v, found := forge.Infra.apps["upstream"]; found && v != nil {
			return v.name
		}
	}
	if forge.Infra.Apps != nil {
		if v, found := forge.Infra.Apps["upstream"]; found {
			return v
		}
	}
	if forge.Infra == nil {
		return
	}
	return forge.Infra.Upstream
}

func (f *Forge) GetString(object, instance, key string) (string, bool, string) {
	v, found, source := f.Get(object, instance, key)
	return v.GetString(), found, source
}

// Get return the value and status of the object instance key
func (f *Forge) Get(object, instance, key string) (value *goforjj.ValueStruct, _ bool, source string) {
	if !f.Init() {
		return
	}
	forge := f.selectCore()
	if forge == nil {
		return
	}
	return forge.Get(object, instance, key)
}

func (f *Forge) GetObjectInstance(object, instance string) interface{} {
	if !f.Init() {
		return nil
	}
	forge := f.selectCore()
	if forge == nil {
		return nil
	}
	return forge.GetObjectInstance(object, instance)
}

func (f *Forge) ObjectLen(object string) int {
	if !f.Init() {
		return 0
	}
	forge := f.selectCore()
	if forge == nil {
		return 0
	}
	switch object {
	case "infra":
		return 1
	case "user":
		if forge.Users == nil {
			return 0
		}
		return len(forge.Users)
	case "group":
		if forge.Groups == nil {
			return 0
		}
		return len(forge.Groups)
	case "app":
		if forge.Apps == nil {
			return 0
		}
		return len(forge.Apps)
	case "repo":
		if forge.Repos == nil {
			return 0
		}
		return len(forge.Repos)
	case "settings":
		return 1
	default:
		if v, found := forge.More[object]; found {
			return len(v)
		}
		return 0
	}
	return 0
}

func (f *Forge) Remove(object, name, key string) {
	forge := f.selectCore()
	if forge == nil {
		return
	}
	forge.Remove(object, name, key)
}

func (f *Forge) Set(source, object, name, key, value string) {
	forge := f.selectCore()
	if forge == nil {
		return
	}
	forge.Set(source, object, name, key, value)
}

func (f *Forge) SetDefault(source, object, name, key, value string) {
	forge := f.selectCore()
	if forge == nil {
		return
	}
	forge.SetDefault(source, object, name, key, value)
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
	forge := f.selectCore()
	if forge == nil {
		return nil
	}

	return forge.Apps
}

// GetDeclaredFlows returns the list of flow to load from the Master Forjfile and the deploy Forjfile.
func (f *Forge) GetDeclaredFlows() (result []string) {
	flows := make(map[string]bool)
	forge := f.selectCore()
	if forge == nil {
		return
	}

	for _, repo := range forge.Repos {
		if repo.Flow.Name != "" {
			flows[repo.Flow.Name] = true
		}
	}
	if deploy, _ := f.GetADeployment(f.GetDeployment()); deploy != nil && deploy.Details != nil {
		for _, repo := range deploy.Details.Repos {
			if repo.Flow.Name != "" {
				flows[repo.Flow.Name] = true
			}
		}
	}

	if flow := forge.ForjSettings.Default.getFlow(); flow != "" {
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

// GetDeployment returns the current deployment environment
func (f *Forge) GetDeployment() string {
	return f.deployTo
}

// SetDeployment defines the current deployment to use.
func (f *Forge) SetDeployment(deployTo string) {
	f.deployTo = deployTo
}

// GetADeployment return the Deployment Object wanted
func (f *Forge) GetADeployment(deploy string) (v *DeploymentStruct, found bool) {
	v, found = f.yaml.Deployments.GetADeployment(deploy)
	return
}

// Validate check if the information in the Forjfile are coherent or not and if code respect some basic rules.
// Validate do not check default values. So, validate can be executed before setting driver default values (forj.ScanAndSetObjectData)
func (f *Forge) Validate() error {
	forge := f.selectCore()
	if forge == nil {
		return fmt.Errorf("No Forjfile Data to validate")
	}

	// ForjSettingsStruct.More

	// Repo connected to a valid deployment
	for _, repo := range forge.Repos {
		if v := repo.Deployment; v != "" {
			if deploy, found := f.yaml.Deployments[v]; !found {
				return fmt.Errorf("Repo '%s': Deployment '%s' doesn't exist. Check deployments section of your Forjfile", repo.name, deploy.name)
			}
		}
	}

	// check if we declared deployment repository in deploy Forjfile.
	for _, deploy := range f.GetDeployments() {
		if deploy.Details == nil || deploy.Details.Repos == nil {
			continue
		}
		for _, repo := range deploy.Details.Repos {
			if repo.Deployment != "" {
				return fmt.Errorf("Unable to declare a deployment Repository in the deployment %s/Forjfile. Remove 'deployment' entry for repo %s", deploy.Name(), repo.name)
			}
		}
	}

	// RepoStruct.More (infra : Repos)

	// AppYamlStruct.More

	// Repository apps connection
	for _, repo := range forge.Repos {
		if repo.Apps == nil {
			continue
		}

		for relAppName, appName := range repo.Apps {
			if _, err := repo.SetInternalRelApp(relAppName, appName); err != nil {
				name, _ := repo.GetString("name")
				return fmt.Errorf("Repo '%s' has an invalid Application reference '%s: %s'. %s", name, relAppName, appName, err)
			}
		}
	}

	// UserStruct.More

	// GroupStruct.More

	// ForgeYaml.More

	// DeploymentStruct
	pro := false
	devDefault := forge.ForjSettings.Default.getDevDeploy()
	devDefaultFound := false
	for name, deploy := range f.yaml.Deployments {
		if deploy.Type == "" {
			return fmt.Errorf("Deployment declaration error in '%s'. Missing type. Provide at least `Type: (PRO|TEST|DEV)`", name)
		}
		if deploy.Type == ProDeployType {
			if pro {
				return fmt.Errorf("Deployment declaration error in '%s'. You cannot have more than 1 deployment of type 'PRO'. Please fix it", name)
			} else {
				pro = true
			}
		}
		if deploy.Type == DevDeployType && devDefault == deploy.name {
			devDefaultFound = true
		}
	}
	if devDefault != "" && !devDefaultFound {
		return fmt.Errorf("Deployment declaration error in '%s'. '%s' is not a valid default DEV deployment name. Please fix it", "forj-settings/default/dev-deploy", devDefault)
	}

	return nil
}

// GetDeployments returns all deployments.
func (f *Forge) GetDeployments() (result Deployments) {
	result = f.yaml.Deployments
	return
}

func (f *Forge) GetDeploymentsModel() (ret *DeploymentsModel) {
	ret = NewDeploymentsModel(f.GetDeployments())
	return
}

// GetDeploymentType return the first found
func (f *Forge) GetDeploymentType(deployType string) (Deployments, bool) {
	return f.yaml.Deployments.GetDeploymentType(deployType)
}

// GetDeploymentPROType return the PRO deployment structure
func (f *Forge) GetDeploymentPROType() (v *DeploymentStruct, err error) {
	return f.yaml.Deployments.GetDeploymentPROType()
}

func (f *Forge) GetUpstreamApps() (v AppsStruct, found bool) {
	v = make(AppsStruct)
	for name, app := range f.Apps() {
		if app.Type == "upstream" {
			v[name] = app
			found = true
		}
	}
	return
}

// GetRepo return the object found
func (f *Forge) GetRepo(name string) (r *RepoStruct, found bool) {
	forge := f.selectCore()
	if forge == nil {
		return
	}
	r, found = forge.GetRepo(name)
	return
}

// Model defines a simple struct to expose Current Application (ie driver instance)
func (f *Forge) Model(instance string) (fModel ForgeModel) {
	if f.inMem == nil {
		fModel.Application.app = f.yaml.ForjCore.Apps[instance]
		return
	}
	fModel.Application.app = f.inMem.Apps[instance]
	return
}
