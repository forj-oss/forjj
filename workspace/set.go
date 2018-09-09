package workspace

import (
	"bufio"
	"fmt"
	"forjj/forjfile"
	"os"
	"strings"

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
	s.value = s.cmd.Flag("value", "workspace data value").Short('P').String()

	s.data = data
}

// doSet register a password to the path given.
// Only supported path are recognized.
func (s *wsSet) doSet() {
	if *s.value == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("INPUT: --  %s  --\nPlease, enter the text to store:\n", *s.key)
		value, _ := reader.ReadString('\n')
		*s.value = strings.TrimSpace(value)
	}

	keyPath := strings.Split(*s.key, "/")

	if !s.data.Set(*s.key, *s.value) {
		gotrace.Info("'%s' workspace text not updated.", *s.key)
		return
	}

	s.data.Save()
	gotrace.Info("'%s' workspace text saved.", *s.key)
}
