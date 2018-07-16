package forjfile

import (
	"github.com/forj-oss/forjj-modules/trace"
)

// ForgeYaml represents the master Forjfile or a piece of the Forjfile model.
type ForgeYaml struct {
	updated     bool
	Deployments Deployments
	ForjCore    DeployForgeYaml `yaml:",inline"`
}

func NewForgeYaml() (ret *ForgeYaml) {
	ret = new(ForgeYaml)
	ret.Init()
	return
}

// Init Initialize the forge. (Forjfile in repository infra)
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

	if f.ForjCore.Infra != nil && f.ForjCore.Infra.More == nil {
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

// defineDefaults update default values defining in Forjfile/forj-settings/default/
//
func (f *ForgeYaml) defineDefaults(warn bool) {

	comm := func(msg string, params ...interface{}) {
		if warn {
			gotrace.Warning(msg+" To eliminate this warning, set the value in Forjfile/forj-settings/default/...", params...)
		} else {
			gotrace.Info(msg, params...)
		}
	}

	for name, deploy := range f.Deployments {
		if deploy.Type == DevDeployType && f.ForjCore.ForjSettings.Default.getDevDeploy() == "" {
			comm("Defining development deployment '%s' as Default (dev-deploy).", name)
			f.ForjCore.ForjSettings.Default.Set("forjj", "dev-deploy", name)
		}
	}
}

// set_defaults
// - set forge in all structures
// - Define a basic Deployment with just 'production' entry
func (f *ForgeYaml) set_defaults() {
	// Cleanup LocalSettings to ensure no local setting remain in a Forjfile
	f.ForjCore.initDefaults(f)

	if len(f.Deployments) == 0 {
		data := DeploymentStruct{}
		data.Desc = "Production environment"
		data.name = "production"
		data.Type = ProDeployType
		f.Deployments = make(map[string]*DeploymentStruct)
		f.Deployments[data.name] = &data
		gotrace.Info("No deployment defined. Created single 'production' deployment. If you want to change that update your forjfile and create a deployment Forfile per deployment under 'deployments/<deploymentName>'.")
	} else {
		for name, deploy := range f.Deployments {
			deploy.name = name
			if deploy.Details != nil {
				deploy.Details.initDefaults(f)
			}
			f.Deployments[name] = deploy
		}
	}
}

func (f *ForgeYaml) dirty() {
	if f == nil {
		return
	}
	f.updated = true
}
