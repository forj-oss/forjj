package forjfile

type DeployForgeModel struct {
	// LocalSettings should not be used from a Forjfile except if this one is a template one.
	LocalSettings WorkspaceStruct
	ForjSettings  ForjSettingsStruct
	Infra         RepoModel
	Deploy        RepoModel
	Repos         map[string]RepoModel
	Apps          map[string]AppModel
	Users         UsersStruct
	Groups        GroupsStruct
	// Collection of Object/Name/Keys=values
	More map[string]map[string]ForjValues `yaml:",inline,omitempty"`
}

// NewDeployForgeModel return a DeployForgeYaml model object
func NewDeployForgeModel(forge *DeployForgeYaml) (ret *DeployForgeModel) {
	ret = new(DeployForgeModel)
	ret.LocalSettings = forge.LocalSettings
	ret.ForjSettings = forge.ForjSettings
	ret.Infra = forge.Infra.Model()
	ret.Repos = make(map[string]RepoModel)
	for name, repo := range forge.Repos {
		ret.Repos[name] = repo.Model()
		if repo.IsCurrentDeploy() {
			ret.Deploy = repo.Model()
		}
	}
	ret.Apps = make(map[string]AppModel)
	for name, app := range forge.Apps {
		ret.Apps[name] = AppModel{
			app: app,
		}
	}

	// TODO: Modelize USERS and GROUPS
	ret.Users = forge.Users
	ret.Groups = forge.Groups

	ret.More = forge.More
	return
}
