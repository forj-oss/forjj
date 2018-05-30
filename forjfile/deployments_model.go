package forjfile

// DeploymentsModel is the ;odel structure of Deployments for text/template
type DeploymentsModel struct {
	list Deployments
}

// NewDeploymentsModel return a model of Deployments
func NewDeploymentsModel(list Deployments) (ret *DeploymentsModel) {
	ret = new(DeploymentsModel)
	ret.setDeployments(list)
	return
}

func (dm *DeploymentsModel) setDeployments(list Deployments) {
	dm.list = list
}

// GetFromType get attributes of a collection of deploy of one type
func (dm *DeploymentsModel) GetFromType(deployType, object, instance, key string) (ret map[string]string) {
	ret = make(map[string]string)
	deploys, found := dm.list.GetDeploymentType(deployType)
	if !found {
		return
	}

	for _, deploy := range deploys {
		if v, keyFound := deploy.Details.GetString(object, instance, key); keyFound {
			ret[deploy.name] = v
		}
	}

	return
}

// GetFromPRO get attribute of a PRO deploy type
func (dm *DeploymentsModel) GetFromPRO(object, instance, key string) string {
	deploy, _ := dm.list.GetDeploymentPROType()
	if v, keyFound := deploy.Details.GetString(object, instance, key); keyFound {
		return v
	}
	return ""
}

// GetFromName get attribute of a PRO deploy type
func (dm *DeploymentsModel) GetFromName(deployName, object, instance, key string) (ret string) {
	if deploy, found := dm.list.GetADeployment(deployName); !found {
		return
	} else if v, keyFound := deploy.Details.GetString(object, instance, key); keyFound {
		return v
	}
	return
}

// GetFromName get attribute of a PRO deploy type
func (dm *DeploymentsModel) Get(deployName string) (_ DeploymentCoreStruct) {
	if deploy, found := dm.list.GetADeployment(deployName); found {
		return deploy.DeploymentCoreStruct
	}
	return
}
