package main

import (
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/cli/clier"
)

type secrets struct {
	secrets *kingpin.CmdClause
	context secretsContext
	common  secretsCommon

	list secretsList

	get struct {
		cmd *kingpin.CmdClause
		key *string
	}

	set secretsSet

	edit secretsEdit

	unset secretsUnset
}

func (s *secrets) init(app *kingpin.Application) {
	if s == nil || app == nil {
		return
	}

	s.secrets = app.Command("secrets", "Manage forjj secrets")
	s.context.init()
	s.common.init(&s.context, s.secrets)
	s.list.init(s.secrets, &s.common)
	s.set.init(s.secrets, &s.common)
	s.edit.init(s.secrets, &s.common)

	s.get.cmd = s.secrets.Command("get", "Get value of a credential unencrypted")
	s.get.key = s.get.cmd.Arg("key", "Full key path").Required().String()

	s.unset.init(s.secrets, &s.common)
}

func (s *secrets) action(action string) {
	actions := strings.Split(action, " ")
	switch actions[1] {
	case "list":
		s.list.showList()
	case "set":
		s.set.doSet()
	case "edit":
		s.edit.doEdit()
	case "unset":
		s.unset.doUnset()
	case "show":
	}
}

// DefineContext define cli Context to permit ParseContext to retrieve
// common variable set.
func (s *secrets) DefineContext(context clier.ParseContexter) {
	s.context.defineContext(context)
}

// GetStringValue Return a field value from the given context (parse time, or after)
func (s *secrets) GetStringValue(field string) (value string, found, isDefault bool, _ error) {
	return s.context.GetStringValue(field)
}
