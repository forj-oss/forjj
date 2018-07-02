package main

import (
	"strings"

	"github.com/alecthomas/kingpin"
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

	unset struct {
		cmd *kingpin.CmdClause
		key *string
	}
}

func (s *secrets) init(app *kingpin.Application) {
	if s == nil || app == nil {
		return
	}

	s.secrets = app.Command("secrets", "Manage forjj secrets")
	s.context.init()
	s.common.init(&s.context, s.secrets)
	s.list.init(s.secrets)
	s.set.init(s.secrets)

	s.get.cmd = s.secrets.Command("get", "Get value of a credential unencrypted")
	s.get.key = s.get.cmd.Arg("key", "Full key path").Required().String()

	s.unset.cmd = s.secrets.Command("unset", "remove a credential key path value from forjj secrets")
	s.unset.key = s.unset.cmd.Arg("key", "Full key path").Required().String()

}

func (s *secrets) action(action string) {
	actions := strings.Split(action, " ")
	switch actions[1] {
	case "list":
		s.list.showList()
	case "set":
		s.set.doSet()
	case "unset":
	case "show":
	}
}
