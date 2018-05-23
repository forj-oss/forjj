package flow

import (
	"fmt"
	"forjj/forjfile"
	"forjj/utils"
	"text/template"

	"github.com/forj-oss/forjj-modules/trace"
)

type FlowTaskSet map[string]map[string]forjfile.ForjValues

func (fts FlowTaskSet) apply(tmpl_data *FlowTaskModel, Forjfile *forjfile.DeployForgeYaml) error {
	tmpl := template.New("flow-set")
	funcs := template.FuncMap{
		"concatenate": fmt.Sprint,
	}
	for object_name, object_data := range fts {
		for instance_name, instance_data := range object_data {
			if v, err := utils.Evaluate(instance_name, tmpl, tmpl_data, funcs); err != nil {
				return fmt.Errorf("Unable to evaluate instance '%s'. %s", instance_name, err)
			} else {
				instance_name = v
			}
			if len(instance_data) == 0 {
				Forjfile.Set(object_name, instance_name, "", "")
				gotrace.Trace("'%s/%s: {}' added.", object_name, instance_name)
				continue
			}
			for key, value := range instance_data {
				if v, err := utils.Evaluate(key, tmpl, tmpl_data, funcs); err != nil {
					return fmt.Errorf("Unable to evaluate instance key '%s'. %s", instance_name, err)
				} else {
					key = v
				}
				if v, err := utils.Evaluate(value.Get(), tmpl, tmpl_data, funcs); err != nil {
					return fmt.Errorf("Unable to evaluate instance '%s' key '%s' value '%s'. %s", instance_name, key, value, err)
				} else {
					if ev := value.Get(); ev != v {
						gotrace.Trace("'%s' has be interpreted as '%s'.", ev, v)
					}
					Forjfile.Set(object_name, instance_name, key, v)
					if v == "" {
						gotrace.Trace("'%s/%s: {}' added. '%s/%s/%s' deleted.",
							object_name, instance_name, object_name, instance_name, key)
					} else {
						gotrace.Trace("'%s/%s/%s=\"%s\"' added.", object_name, instance_name, key, v)
					}

				}
			}
		}
	}
	return nil
}
