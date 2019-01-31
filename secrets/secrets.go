package secrets

import (
	"forjj/creds"
	"forjj/drivers"
	"forjj/forjfile"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/cli/clier"
)

type Secrets struct {
	secrets *kingpin.CmdClause
	Context Context
	common  common

	list sList

	get struct {
		cmd *kingpin.CmdClause
		key *string
	}

	set sSet

	edit sEdit

	unset sUnset
}

// Init initialize the secrets cli commands
func (s *Secrets) Init(app *kingpin.Application, forjfile *forjfile.Forge, drivers *drivers.Drivers, secrets *creds.Secure, isParsePhase func() bool, initCommon func(context *Context, cmd *kingpin.CmdClause)) {
	if s == nil || app == nil {
		return
	}

	s.secrets = app.Command("secrets", "Manage forjj secrets")
	s.Context.init(isParsePhase)
	s.common.init(&s.Context, s.secrets, initCommon)
	s.list.init(s.secrets, &s.common, forjfile, drivers, secrets)
	s.set.init(s.secrets, &s.common, forjfile, drivers, secrets)
	s.edit.init(s.secrets, &s.common, forjfile, drivers, secrets)

	s.get.cmd = s.secrets.Command("get", "Get value of a credential unencrypted")
	s.get.key = s.get.cmd.Arg("key", "Full key path").Required().String()

	s.unset.init(s.secrets, &s.common, forjfile, drivers, secrets)
}

func (s *Secrets) Action(action string) {
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
func (s *Secrets) DefineContext(context clier.ParseContexter) {
	s.Context.defineContext(context)
}

// GetStringValue Return a field value from the given context (parse time, or after)
func (s *Secrets) GetStringValue(field string) (value string, found, isDefault bool, _ error) {
	return s.Context.GetStringValue(field)
}
