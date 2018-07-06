package main

import (
	"forjj/scandrivers"
	"forjj/utils"
	"io/ioutil"
	"os"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

type secretsEdit struct {
	editor   *string
	cmd      *kingpin.CmdClause
	key      *string
	password string

	elements map[string]secretInfo
}

func (e *secretsEdit) init(parent *kingpin.CmdClause) {

	e.cmd = parent.Command("edit", "edit a credential stored in forjj secrets")
	e.key = e.cmd.Arg("key", "Key path. Format is <objectType>/<objectInstance>/<key>.)").Required().String()
	e.editor = e.cmd.Flag("editor", "editor to execute").Envar("EDITOR").Default("/usr/bin/vi").String()
}

// doSet register a password to the path given.
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
		gotrace.Error("'%s' is not a valid secret path. check with `forjj secrets`")
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

	keyPath := strings.Split(*s.key, "/")

	s.password, _, _, _ = forj_app.s.GetString(keyPath[0], keyPath[1], keyPath[2])

	_, err = tmpFile.WriteString(s.password)
	if err != nil {
		gotrace.Error("Unable to write temporary file in /tmp. Exiting.")
		return
	}
	fileName := tmpFile.Name()
	tmpFile.Close()

	if utils.RunCmd(*s.editor, fileName) != 0 {
		gotrace.Error("Unable to start the editor on file %s. Exiting.", tmpFile.Name())
		return
	}

	if data, err := ioutil.ReadFile(fileName); err != nil {
		gotrace.Error("Unable to read the editor file %s. Exiting.", tmpFile.Name())
		return
	} else {
		s.password = strings.Trim(string(data), " \n")
	}

	if len(s.password) == 0 {
		gotrace.Info("File is empty. Update ignored.")
		return
	}
	os.Remove(fileName)
	v := goforjj.ValueStruct{}
	v.Set(s.password)
	env := forj_app.f.GetDeployment()
	if !forj_app.s.SetObjectValue(env, "forjj", keyPath[0], keyPath[1], keyPath[2], &v) {
		gotrace.Info("'%s' secret text not updated.", *s.key)
		return
	}

	forj_app.s.SaveEnv(env)
	gotrace.Info("'%s' secret text saved in '%s' deployment environment.", *s.key, env)
}
