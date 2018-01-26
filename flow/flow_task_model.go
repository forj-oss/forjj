package flow

import "forjj/forjfile"

type FlowTaskModel struct {
	Current *forjfile.RepoStruct
	Forjfile *forjfile.ForgeYaml
	List map[string]interface{}
}

func New_FlowTaskModel(current *forjfile.RepoStruct, forjf *forjfile.ForgeYaml) (ret *FlowTaskModel) {
	ret = new(FlowTaskModel)

	ret.Current = current
	ret.Forjfile = forjf
	return
}
