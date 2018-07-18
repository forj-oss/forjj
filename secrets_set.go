package main

import (
	"forjj/creds"
	"fmt"
	"forjj/scandrivers"
	"os"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/goforjj"
)

type secretsSet struct {
	cmd      *kingpin.CmdClause
	key      *string
	password *string
	common   *secretsCommon

	elements map[string]secretInfo
}

func (s *secretsSet) init(parent *kingpin.CmdClause, common *secretsCommon) {
	s.cmd = parent.Command("set", "store a new credential in forjj secrets")
	s.key = s.cmd.Arg("key", "Key path. Format is <objectType>/<objectInstance>/<key>.)").Required().String()
	s.password = s.cmd.Flag("password", "Secret key value").Short('P').String()
	s.common = common
}

// doSet register a password to the path given.
// Only supported path are recognized.
func (s *secretsSet) doSet() {
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

			if *s.common.common {
				info.value, info.found, info.source, info.env = forj_app.s.GetGlobalString(objectName, instanceName, keyName)
			} else {
				info.value, info.found, info.source, info.env = forj_app.s.GetString(objectName, instanceName, keyName)
			}

			s.elements[info.keyPath] = info
		}
		return nil
	})
	scan.DoScanDriversObject()

	if _, found := s.elements[*s.key]; !found {
		gotrace.Error("'%s' is not a valid secret path. check with `forjj secrets`", *s.key)
		return
	}

	if *s.password == "" {
		fmt.Printf("INPUT: --  %s  --\nPlease, enter the secret text to store:\n", *s.key)
		if v, err := terminal.ReadPassword(int(os.Stdout.Fd())); err != nil {
			gotrace.Error("Password read issue. %s", err)
			return
		} else {
			*s.password = string(v)
			fmt.Println()
		}
	}

	keyPath := strings.Split(*s.key, "/")

	v := goforjj.ValueStruct{}
	v.Set(*s.password)
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
