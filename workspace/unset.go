package workspace

import (
	"forjj/creds"
	"forjj/scandrivers"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/goforjj"
)

type wsUnset struct {
	cmd    *kingpin.CmdClause
	key    *string
	common *wsCommon

	elements map[string]Info
}

func (s *wsUnset) init(parent *kingpin.CmdClause, common *wsCommon) {
	s.cmd = parent.Command("unset", "remove a credential stored in forjj secrets")
	s.key = s.cmd.Arg("key", "Key path to remove. Format is <objectType>/<objectInstance>/<key>.)").Required().String()
	s.common = common
}

// doSet register a password to the path given.
// Only supported path are recognized.
func (s *wsUnset) doUnset() {
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

	keyPath := strings.Split(*s.key, "/")

	env := forj_app.f.GetDeployment()
	if *s.common.common {
		env = creds.Global
	}
	if !forj_app.s.UnsetObjectValue(env, "forjj", keyPath[0], keyPath[1], keyPath[2]) {
		gotrace.Info("'%s' secret text not updated.", *s.key)
		return
	}

	forj_app.s.SaveEnv(env)
	gotrace.Info("'%s' secret text removed from '%s' deployment environment.", *s.key, env)
}
