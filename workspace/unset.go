package workspace

import (
	"forjj/forjfile"

	"github.com/forj-oss/forjj-modules/trace"

	"github.com/alecthomas/kingpin"
)

type wsUnset struct {
	cmd *kingpin.CmdClause
	key *string

	data *forjfile.Workspace
}

func (s *wsUnset) init(parent *kingpin.CmdClause, data *forjfile.Workspace) {
	s.cmd = parent.Command("unset", "remove a credential stored in forjj secrets")
	s.key = s.cmd.Arg("key", "Key path to remove. Format is <objectType>/<objectInstance>/<key>.)").Required().String()
}

// doSet register a password to the path given.
// Only supported path are recognized.
func (s *wsUnset) doUnset() {

	if s.data.Unset(*s.key) {
		gotrace.Info("'%s' workspace text not updated.", *s.key)
		return
	}

	s.data.Save()
	gotrace.Info("'%s' workspace text removed or restored to default value.", *s.key)
}
