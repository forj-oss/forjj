package forjfile

import (
	"fmt"
)

// Deployments is a collection deployment
type Deployments map[string]*DeploymentStruct

// GetDeploymentType return a list of deployment type.
func (d Deployments) GetDeploymentType(deployType string) (v Deployments, found bool) {
	v = make(Deployments)

	for name, deploy := range d {
		if deploy.Type == deployType {
			v[name] = deploy
			found = true
		}
	}
	return
}

// GetDeploymentPROType return the PRO deployment structure
func (d Deployments) GetDeploymentPROType() (v *DeploymentStruct, err error) {

	if deployObjs, _ := d.GetDeploymentType("PRO"); len(deployObjs) != 1 {
		err = fmt.Errorf("Found more than one PRO environment")
	} else {
		for k := range deployObjs {
			v = deployObjs[k]
			break
		}
	}

	return
}

// GetADeployment return the Deployment Object wanted
func (d Deployments) GetADeployment(deploy string) (v *DeploymentStruct, found bool) {
	if deploy == "" {
		return
	}
	v, found = d[deploy]
	return
}

func (d Deployments) One() (v *DeploymentStruct) {
	for _, deploy := range d {
		return deploy
	}
	return
}