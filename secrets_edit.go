package main

import (
	"forjj/creds"
	"forjj/scandrivers"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

type secretsEdit struct {
	editor   *string
	cmd      *kingpin.CmdClause
	key      *string
	password string
	common   *secretsCommon

	elements map[string]secretInfo
}

func (s *secretsEdit) init(parent *kingpin.CmdClause, common *secretsCommon) {

	s.cmd = parent.Command("edit", "edit a credential stored in forjj secrets")
	s.key = s.cmd.Arg("key", "Key path. Format is <objectType>/<objectInstance>/<key>.)").Required().String()
	s.editor = s.cmd.Flag("editor", "editor to execute").Envar("EDITOR").Default("/usr/bin/vi").String()
	s.common = common
}

// doEdit register a password to the path given.
// Only supported path are recognized.
func (s *secretsEdit) doEdit() {
	ffd := forj_app.f.InMemForjfile()

	scan := scandrivers.NewScanDrivers(ffd, forj_app.drivers)
	s.elements = make(map[string]secretInfo)

	// Retrieve secrets path
	scan.SetScanObjFlag(func(objectName, instanceName, flagPrefix, name string, flag goforjj.YamlFlag) error {
		if flag.Options.Secure {
			info := secretInfo{}
			info.keyPath = objectName + "/" + instanceName + "/"
			keyName := name
			if flagPrefix != "" {
				keyName = flagPrefix + name
			}
			info.keyPath += keyName

			info.value, info.found, info.source, info.env = forj_app.s.GetString(objectName, instanceName, keyName)

			s.elements[info.keyPath] = info
		}
		return nil
	})
	scan.DoScanDriversObject()

	if _, found := s.elements[*s.key]; !found {
		gotrace.Error("'%s' is not a valid secret path. check with `forjj secrets`", *s.key)
		return
	}

	if *s.editor == "" {
		*s.editor = "/usr/bin/vi"
	}

	tmpFile, err := ioutil.TempFile("/tmp", "forjj-")
	if err != nil {
		gotrace.Error("Unable to create temporary file in /tmp. Exiting.")
		return
	}

	fileName := tmpFile.Name()
	defer os.Remove(fileName)

	keyPath := strings.Split(*s.key, "/")

	s.password, _, _, _ = forj_app.s.GetString(keyPath[0], keyPath[1], keyPath[2])

	_, err = tmpFile.WriteString(s.password)
	if err != nil {
		gotrace.Error("Unable to write temporary file in /tmp. Exiting.")
		return
	}
	tmpFile.Close()

	cmd := exec.Command(*s.editor, fileName)
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		gotrace.Error("Unable to edit %s. Not a terminal. Exiting.", *s.key)
		return
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	if err != nil {
		gotrace.Error("Unable to start the editor %s on file %s. %s. Exiting.", *s.editor, fileName, err)
		return
	}

	if data, err := ioutil.ReadFile(fileName); err != nil {
		gotrace.Error("Unable to read the editor file %s. Exiting.", fileName)
		return
	} else {
		s.password = strings.Trim(string(data), " \n")
	}

	if len(s.password) == 0 {
		gotrace.Info("File is empty. Update ignored.")
		return
	}
	v := goforjj.ValueStruct{}
	v.Set(s.password)
	env := forj_app.f.GetDeployment()
	if *s.common.common {
		env = creds.Global
	}
	if !forj_app.s.SetObjectValue(env, "forjj", keyPath[0], keyPath[1], keyPath[2], &v) {
		gotrace.Info("'%s' secret text not updated.", *s.key)
		return
	}

	forj_app.s.SaveEnv(env)
	gotrace.Info("'%s' secret text saved in '%s' deployment environment.", *s.key, env)
}
