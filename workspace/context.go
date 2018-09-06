package workspace

import (
	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/cli/interface"
	"github.com/forj-oss/forjj-modules/cli/kingpinCli"
)

// Following deals with forjj-modules/cli at Parse time
// When some parameters needs to be retrieved at parsetime (Used in ParseContext)
// we need to let ParseContext get those values.
// So, we need to call flag/arg functions at init phase
//
// Ex: 

type Context struct {
	params     map[string]cli.ForjParam
	cliContext clier.ParseContexter
}

func (s *Context) init() {
	s.params = make(map[string]cli.ForjParam)
}

// Create a cli flag from a kingpin flag (for forjj-modules/cli)
func (s *Context) flag(name string, flag *kingpin.FlagClause) (cliFlag *kingpinCli.FlagClause) {
	if s == nil {
		return nil
	}

	cliFlag = kingpinCli.NewFlag(flag)
	param := cli.NewForjFlag(cliFlag)
	s.params[name] = param
	return
}

// defineContext store the forjj-modules/cli context for GetStringValue
func (s *Context) defineContext(context clier.ParseContexter) {
	if s == nil {
		return
	}
	s.cliContext = context
}

// getContextFlagValue Get Flag value from current cli context
func (s *Context) getContextFlagValue(name string) (interface{}, bool) {
	if s == nil {
		return nil, false
	}
	param, found := s.params[name]
	if !found {
		return nil, false
	}
	return param.GetContextValue(s.cliContext)
}

// GetStringValue return value and status where the value were found.
//
// WARNING: Default status can be set only in cli load context phase (before parse)
// If we need to incorporate some data feed between real value and default value
// it must be done and saved during load context phase. (ie ParseContext() in cli_context.go)
func (s *Context) GetStringValue(field string) (value string, found, isDefault bool, _ error) {
	var param cli.ForjParam

	param, found = s.params[field]
	if !found {
		return
	}

	var v interface{}
	if !forj_app.cli.IsParsePhase() {
		v, found = param.GetContextValue(s.cliContext)
		if !found {
			return
		}
		if fieldValue, ok := v.(string); ok {
			value = fieldValue
		} else if fieldDefault, ok := v.(*string); ok {
			value = *fieldDefault
			isDefault = true
		}
	} else {
		found = param.IsFound()
		if !found {
			return
		}
		value = param.GetStringValue()
	}
	return
}
