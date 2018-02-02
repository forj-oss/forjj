package flow

import "forjj/forjfile"

type FlowTaskLists []*FlowTaskList

type FlowTaskList struct {
	Name string
	List string
	Parameters []string
	list []interface{}
}

func (ftl *FlowTaskList)Get(repo *forjfile.RepoStruct, _ *forjfile.Forge) (list []interface{}) {
	list = []interface{}{}

	switch ftl.List {
	case "GetApps":
		if ftl.Parameters == nil {
			ftl.Parameters = []string{}
		}
		if apps , err := repo.GetApps(ftl.Parameters...) ; err == nil {
			list = make([]interface{}, 0, len(apps))
			for _, app := range apps {
				list = append(list, app.Model())
			}
		}
	}
	return
}
