package flow

import (
	"forjj/forjfile"
	"text/template"
	"fmt"
	"bytes"
	"strconv"
	"strings"
	"github.com/forj-oss/forjj-modules/trace"
)

type FlowTaskIf struct {
	Rule string
	List map[string]string `yaml:",inline"`
}

// IfEvaluate will interpret
func (fti *FlowTaskIf)IfEvaluate(repo *forjfile.RepoStruct, Forjfile *forjfile.DeployForgeYaml) (_ bool, _ error) {
	if fti.Rule != "" {
		var doc bytes.Buffer

		if t, err:= template.New("flow-eval").Funcs(template.FuncMap{
				}).Parse(fti.Rule); err != nil {
			return false, fmt.Errorf("Error in template evaluation. %s", err)
		} else {
			if err = t.Execute(&doc, New_FlowTaskModel(repo, Forjfile)) ; err != nil {
				return false, fmt.Errorf("Unable to evaluate '%s'. %s", fti.Rule, err)
			}
		}

		result := doc.String()
		gotrace.Trace("'%s' evaluated to '%s'", fti.Rule, result)
		switch strings.ToLower(result) {
		case "", "not found" :
			return
		case "found" :
			return true, nil
		default:
			return strconv.ParseBool(result)
		}
	}

	if fti.List != nil {
		rules := make([]string, 0, len(fti.List))
		for key, value := range fti.List {
			rules = append(rules, key + ":" + value)
		}
		return repo.HasValues(rules ...)
	}
	return true, nil
}
