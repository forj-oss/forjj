package main

type Forjfile struct {
	Forj ForjDefaultStruct
	Infra RepoStruct
	W WorkspaceStruct `yaml:"local"`
	Repos map[string]RepoStruct
	Apps map[string]map[string]string
	Instances map[string]map[string]map[string]string `yaml:",inline"`
}

type ForjDefaultStruct struct {
	Organization string
	More map[string]string `yaml:",inline"`
}

type RepoStruct struct {
	Name string
	Upstream string
	More map[string]string `yaml:",inline"`
}

type WorkspaceStruct struct {
	ContribsPath string `yml:"contribs-path"`
	More map[string]string `yaml:",inline"`
}

type AppStruct struct {
	Name string
	Type string
	Driver string
	More map[string]string `yaml:",inline"`
}
