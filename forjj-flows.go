package main

// Define a collection of functions to apply a flow to a repository.
// The flow definition is loaded from forjj-flows

type FlowDef struct {
    Title string            // Flow title
    Define map[string]FlowPluginTypeDef
    Apps map[string]FlowTasksDef
}

type FlowTasksDef []FlowTaskDef

type FlowPluginTypeDef struct {
    MaxInstances int `yaml:"max_instances"`
    Roles []string
}

type FlowTaskDef struct {
    Commit string
    Task []FlowTaskDo `yaml:"do"`
}

type FlowTaskDo struct {
    Api FlowTaskAPIData
    Apps map[string]FlowTaskAPIData
}

type FlowTaskAPIData map[string]string
