package main

import (
	"forjj/forjfile"
)

type forjDeployContext struct {
	to string
	obj *forjfile.DeploymentStruct
	repoAttached *forjfile.RepoStruct
}

func (c *forjDeployContext) repoName() (ret string) {
	if c.repoAttached != nil {
		ret = c.repoAttached.Name()
	}
	return
}