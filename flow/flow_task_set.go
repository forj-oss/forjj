package flow

import (
	"forjj/forjfile"
	"text/template"
	"forjj/utils"
	"fmt"
	"github.com/forj-oss/forjj-modules/trace"
)

type FlowTaskSet map[string]map[string]forjfile.ForjValues

func (fts FlowTaskSet)apply(tmpl_data *FlowTaskModel, Forjfile *forjfile.Forge) error {
	tmpl := template.New("flow-set")
	funcs := template.FuncMap{}
	for object_name, object_data := range fts {
		for instance_name, instance_data := range object_data {
			if v, err := utils.Evaluate(instance_name, tmpl, tmpl_data, funcs) ; err != nil {
				return fmt.Errorf("Unable to evaluate instance '%s'. %s", instance_name, err)
			} else {
				instance_name = v
			}
			if len(instance_data) == 0 {
				Forjfile.Set(object_name, instance_name, "", "")
				gotrace.Trace("instance '%s/%s' added with no keys.", object_name, instance_name)
				continue
			}
			for key, value:= range instance_data {
				if v, err := utils.Evaluate(key, tmpl, tmpl_data, funcs) ; err != nil {
					return fmt.Errorf("Unable to evaluate instance key '%s'. %s", instance_name, err)
				} else {
					key = v
				}
				if v, err := utils.Evaluate(value.Get(), tmpl, tmpl_data, funcs) ; err != nil {
					return fmt.Errorf("Unable to evaluate instance '%s' key '%s' value '%s'. %s", instance_name, key, value, err)
				} else {
					Forjfile.Set(object_name, instance_name, key, v)
					gotrace.Trace("instance '%s/%s' key '%s' value '%s' added.", object_name, instance_name, key, value.Get())
				}
			}
		}
	}
	return nil
}

