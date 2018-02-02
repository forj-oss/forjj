package flow

import "forjj/forjfile"

type FlowTaskModel struct {
	Repo forjfile.RepoModel
	Forjfile forjfile.ForgeModel
	List map[string]interface{}
}

func New_FlowTaskModel(current *forjfile.RepoStruct, forjf *forjfile.Forge) (ret *FlowTaskModel) {
	ret = new(FlowTaskModel)

	ret.Repo = current.Model()
	ret.Forjfile = forjf.Model()
	return
}
