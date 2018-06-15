package forjfile

// DeploymentStruct represent the data structure of all deployment.
type DeploymentStruct struct {
	DeploymentCoreStruct `yaml:",inline"`
	Details              *DeployForgeYaml `yaml:"define,omitempty"`
}

// MarshalYAML provides the encoding part for DeploymentStruct
//
// In short we do not want to encode forjj deployment details) info except the core.
func (d DeploymentStruct) MarshalYAML() (interface{}, error) {
	return d.DeploymentCoreStruct, nil
}

// UpdateDeploymentCoreData set all DeploymentCore data
func (d *DeploymentStruct) UpdateDeploymentCoreData(data DeploymentCoreStruct) {
	d.DeploymentCoreStruct = data
}

// RunInContext run GIT commands in the GIT repo context.
func (d *DeploymentStruct) RunInContext(doRun func() error) (err error) {
	return d.DeploymentCoreStruct.RunInContext(doRun)
}
