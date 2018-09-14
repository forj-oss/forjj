package workspace

import (
	"forjj/forjfile"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
)

type wsEdit struct {
	editor          *string
	noTerminalSetup *bool
	cmd             *kingpin.CmdClause
	key             *string
	password        string

	data *forjfile.Workspace
}

func (s *wsEdit) init(parent *kingpin.CmdClause, data *forjfile.Workspace) {

	s.cmd = parent.Command("edit", "edit a credential stored in forjj secrets")
	s.key = s.cmd.Arg("key", "Key path. Format is <objectType>/<objectInstance>/<key>.)").Required().String()
	s.editor = s.cmd.Flag("editor", "editor to execute").Envar("EDITOR").Default("/usr/bin/vi").String()
	s.noTerminalSetup = s.cmd.Flag("no-terminal-setup", "do not set terminal i/o to edit the credential.").Short('T').Bool()

	s.data = data
}

// doEdit register a password to the path given.
// Only supported path are recognized.
func (s *wsEdit) doEdit() {
	value := ""
	if v, found := s.data.Get(*s.key); !found {
		gotrace.Error("'%s' is not a valid workspace path. check with `forjj workspace`", *s.key)
		return
	} else {
		value = v
	}

	if *s.editor == "" {
		*s.editor = "/usr/bin/vi"
	}

	tmpFile, err := ioutil.TempFile("/tmp", "forjj-ws-")
	if err != nil {
		gotrace.Error("Unable to create temporary file in /tmp. Exiting.")
		return
	}

	fileName := tmpFile.Name()
	defer os.Remove(fileName)

	_, err = tmpFile.WriteString(value)
	if err != nil {
		gotrace.Error("Unable to write temporary file in /tmp. Exiting.")
		return
	}
	tmpFile.Close()

	cmd := exec.Command(*s.editor, fileName)
	if !*s.noTerminalSetup {
		if !terminal.IsTerminal(int(os.Stdin.Fd())) {
			gotrace.Error("Unable to edit %s. Not a terminal. Exiting.", *s.key)
			return
		}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err = cmd.Run()

	if err != nil {
		gotrace.Error("Unable to start the editor %s on file %s. %s. Exiting.", *s.editor, fileName, err)
		return
	}

	if data, err := ioutil.ReadFile(fileName); err != nil {
		gotrace.Error("Unable to read the editor file %s. Exiting.", fileName)
		return
	} else {
		value = strings.Trim(string(data), " \n")
	}

	if len(value) == 0 {
		gotrace.Info("File is empty. Update ignored.")
		return
	}

	if !s.data.Set(*s.key, value, true) {
		gotrace.Info("'%s' secret text not updated.", *s.key)
		return
	}

	s.data.Save()
	gotrace.Info("'%s' workspace text saved.", *s.key)
}
