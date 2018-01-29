package flow

import (
	"forjj/forjfile"
	"github.com/forj-oss/forjj-modules/trace"
	"fmt"
	"forjj/utils"
)

type FlowDefine struct { // Yaml structure
	Name   string
	Title  string // Flow title
	Define map[string]FlowPluginTypeDef
	OnRepo map[string]FlowTaskDef `yaml:"on-repo-do"`
}

func (fd *FlowDefine)apply(repo *forjfile.RepoStruct, Forjfile *forjfile.Forge) error {
	for _, flowTask := range fd.OnRepo {
		gotrace.Trace("flow '%s': %s on repository %s", fd.Name, flowTask.Description, repo.GetString("name"))

		task_to_set, err := flowTask.if_section(repo, Forjfile)
		if err != nil {
			return fmt.Errorf("Flow: '%s'. Unable to apply flow task '%s'.", fd.Name, err)
		}

		if ! task_to_set {
			gotrace.Trace("Flow task not applied to repo '%s'. if condition fails.", repo.GetString("name"))
			continue
		}

		gotrace.Trace("Flow task '%s' applying to repo '%s'.", fd.Name, repo.GetString("name"))

		tmpl_data := New_FlowTaskModel(repo, Forjfile.Forjfile())

		if flowTask.List == nil {
			if err := flowTask.Set.apply(tmpl_data, Forjfile); err != nil {
				return fmt.Errorf("Unable to apply flow task '%s' on repo '%s'. %s", fd.Name, repo.GetString("name"), err)
			}
			gotrace.Trace("flow task '%s' applied on repo '%s'.", fd.Name, repo.GetString("name"))
			continue
		}

		// Load list
		max := make([]int, len(flowTask.List))

		for index, taskList := range flowTask.List {
			taskList.list = taskList.Get(repo, Forjfile)
			max[index] = len(taskList.list)
		}

		// Loop on list and set CurrentList
		looplist := utils.NewMLoop(max...)
		tmpl_data.List = make(map[string]interface{})
		for !looplist.Eol() {
			for index, pos := range looplist.Cur() {
				flowTaskList := flowTask.List[index]
				tmpl_data.List[flowTaskList.Name] = flowTaskList.list[pos]
			}

			if err := flowTask.Set.apply(tmpl_data, Forjfile); err != nil {
				return fmt.Errorf("Unable to apply flow task '%s' on repo '%s'. %s", fd.Name, repo.GetString("name"), err)
			}
			gotrace.Trace("flow task '%s' applied on repo '%s'.", fd.Name, repo.GetString("name"))

			looplist.Increment()
		}
	}
	return nil
}

func (ftd *FlowTaskDef)if_section(repo *forjfile.RepoStruct, Forjfile *forjfile.Forge) (task_to_set bool, _ error) {
	task_to_set = true
	if ftd.If != nil {
		for _, ftif := range ftd.If {
			if v, err := ftif.IfEvaluate(repo, Forjfile); err != nil {
				return false, err
			} else if !v {
				task_to_set = false
				break
			}
		}
	}
	return
}
