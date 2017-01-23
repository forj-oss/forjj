package main

type ForjFileCore struct {
	Workspace map[string]string		  `yaml:"local-settings`
	Apps map[string]string			  `yaml:"applications"`
	Repos map[string]string			  `yaml:"repositories"`
    Objects map[string]ForjFileObject `yaml:"forj"`
}

type ForjFileObject map[string]interface{}


