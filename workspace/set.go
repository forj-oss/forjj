package workspace

import (
	"forjj/forjfile"

	"github.com/forj-oss/forjj-modules/trace"

	"github.com/alecthomas/kingpin"
)

type wsSet struct {
	cmd   *kingpin.CmdClause
	key   *string
	value *string

	data *forjfile.Workspace
}

func (s *wsSet) init(parent *kingpin.CmdClause, data *forjfile.Workspace) {
	s.cmd = parent.Command("set", "store a new credential in forjj secrets")
	s.key = s.cmd.Arg("key", "Workspace key path.)").Required().String()
	s.value = s.cmd.Arg("value", "workspace data value").Required().String()

	s.data = data
}

// doSet register a password to the path given.
// Only supported path are recognized.
func (s *wsSet) doSet() {
	if !s.data.Set(*s.key, *s.value, true) {
		gotrace.Info("'%s' workspace text not updated.", *s.key)
		return
	}

	s.data.Save()
}
