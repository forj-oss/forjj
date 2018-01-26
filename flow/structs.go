package flow

type FlowPluginTypeDef struct {
	MaxInstances int `yaml:"max_instances"`
	Roles        []string
}

type FlowTaskDef struct {
	Description string

	If []FlowTaskIf

	List []*FlowTaskList `yaml:"loop-on-list"`

	Set FlowTaskSet // key1: object, key2: instance, key3: value key, then value
}
