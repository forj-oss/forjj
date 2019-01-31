package secrets

import (
	"forjj/creds"
	"forjj/drivers"
	"forjj/forjfile"
	"forjj/scandrivers"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/goforjj"
)

type sUnset struct {
	cmd    *kingpin.CmdClause
	key    *string
	common *common

	elements map[string]sInfo

	forjfile *forjfile.Forge
	drivers  *drivers.Drivers
	secrets  *creds.Secure
}

func (s *sUnset) init(parent *kingpin.CmdClause, common *common, forjfile *forjfile.Forge, drivers *drivers.Drivers, secrets *creds.Secure) {
	s.cmd = parent.Command("unset", "remove a credential stored in forjj secrets")
	s.key = s.cmd.Arg("key", "Key path to remove. Format is <objectType>/<objectInstance>/<key>.)").Required().String()
	s.common = common

	s.forjfile = forjfile
	s.drivers = drivers
	s.secrets = secrets
}

// doSet register a password to the path given.
// Only supported path are recognized.
func (s *sUnset) doUnset() {
	ffd := s.forjfile.InMemForjfile()

	scan := scandrivers.NewScanDrivers(ffd, s.drivers)
	s.elements = make(map[string]sInfo)

	// Retrieve secrets path
	scan.SetScanObjFlag(func(objectName, instanceName, flagPrefix, name string, flag goforjj.YamlFlag) error {
		if flag.Options.Secure {
			info := sInfo{}
			info.keyPath = objectName + "/" + instanceName + "/"
			keyName := name
			if flagPrefix != "" {
				keyName = flagPrefix + name
			}
			info.keyPath += keyName

			info.value, info.found, info.source, info.env = s.secrets.GetString(objectName, instanceName, keyName)

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

	env := s.forjfile.GetDeployment()
	if *s.common.common {
		env = creds.Global
	}
	if !s.secrets.UnsetObjectValue(env, "forjj", keyPath[0], keyPath[1], keyPath[2]) {
		gotrace.Info("'%s' secret text not updated.", *s.key)
		return
	}

	s.secrets.SaveEnv(env)
	gotrace.Info("'%s' secret text removed from '%s' deployment environment.", *s.key, env)
}
