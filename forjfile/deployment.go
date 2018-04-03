package forjfile

// DeploymentStruct represent the data structure of all deployment.
type DeploymentStruct struct {
	DeploymentCoreStruct
	More *DeployForgeYaml `yaml:",inline"`
}

// DeploymentCoreStruct contains only deployment information. anything others kind of information
type DeploymentCoreStruct struct {
	Desc string            `yaml:"description,omitempty"`
	Pars map[string]string `yaml:"parameters"`
}

// MarshalYAML provides the encoding part for DeploymentStruct
//
// In short we do not want to encode forjj deployment details) info except the core.
func (d DeploymentStruct) MarshalYAML() (interface{}, error) {
	return d.DeploymentCoreStruct, nil
}
