package secrets

import (
	"fmt"
	"forjj/creds"
	"forjj/drivers"
	"forjj/forjfile"
	"forjj/scandrivers"
	"forjj/utils"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/forj-oss/forjj-modules/trace"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/goforjj"
)

type sSet struct {
	cmd      *kingpin.CmdClause
	key      *string
	password *string
	common   *common
	copyFile *string
	linkFile *string

	elements map[string]sInfo

	forjfile *forjfile.Forge
	drivers  *drivers.Drivers
	secrets  *creds.Secure
}

func (s *sSet) init(parent *kingpin.CmdClause, common *common, forjfile *forjfile.Forge, drivers *drivers.Drivers, secrets *creds.Secure) {
	s.cmd = parent.Command("set", "store a new credential in forjj secrets")
	s.key = s.cmd.Arg("key", "Key path. Format is <objectType>/<objectInstance>/<key>.)").Required().String()
	s.password = s.cmd.Flag("password", "Secret key value").Short('P').String()
	s.copyFile = s.cmd.Flag("from-file", "Copy secret in forjj internal from the file name given. The file must exists.").String()
	s.linkFile = s.cmd.Flag("use-file", "Link the secret to the file name given.").String()
	s.common = common

	s.forjfile = forjfile
	s.drivers = drivers
	s.secrets = secrets

}

// doSet register a password to the path given.
// Only supported path are recognized.
func (s *sSet) doSet() {
	env := s.loadSecretsInfo()

	if _, found := s.elements[*s.key]; !found {
		gotrace.Error("'%s' is not a valid secret path. check with `forjj secrets`", *s.key)
		return
	}

	if (*s.copyFile != "" && *s.linkFile != "") || (*s.password != "" && (*s.copyFile != "" || *s.linkFile != "")) {
		gotrace.Error("Incorrect flags combination. flags --copy-file, --link-file and --password are exclusive. Choose the proper one and retry.")
		return
	} else if *s.copyFile != "" {
		s.copyFromFile(env)
	} else if *s.linkFile != "" {
		s.linkToFile(env)
	} else {
		s.setInternal(env)
	}

	s.secrets.SaveEnv(env)
	gotrace.Info("'%s' secret text saved in '%s' deployment environment.", *s.key, env)
}

// loadSecretsInfo load the list of secrets
func (s *sSet) loadSecretsInfo() (env string) {
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

			if *s.common.common {
				info.value, info.found, info.source, info.env = s.secrets.GetGlobalString(objectName, instanceName, keyName)
			} else {
				info.value, info.found, info.source, info.env = s.secrets.GetString(objectName, instanceName, keyName)
			}

			s.elements[info.keyPath] = info
		}
		return nil
	})
	scan.DoScanDriversObject()

	env = s.forjfile.GetDeployment()
	if *s.common.common {
		env = creds.Global
	}
	return
}

// Set an internal secret
func (s *sSet) setInternal(env string) {
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

	v := creds.NewObjectsValue("internal", goforjj.NewValueStruct(*s.password))

	if !s.secrets.SetObjectValue(env, "internal", keyPath[0], keyPath[1], keyPath[2], v) {
		gotrace.Info("'%s' secret text not updated.", *s.key)
		return
	}
}

func (s *sSet) copyFromFile(env string) {
	file := *s.copyFile

	if f, err := utils.Abs(file); err != nil {
		gotrace.Error("Unable to determine file path. %s", err)
		return
	} else {
		file = f
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		gotrace.Error("Unable to copy file content to forjj secret %s", *s.key)
		return
	}

	keyPath := strings.Split(*s.key, "/")

	v := creds.NewObjectsValue("internal", goforjj.NewValueStruct(data))
	v.AddResource("copied-from", file)
	v.AddResource("copied-on", time.Now().String())

	if !s.secrets.SetObjectValue(env, "", keyPath[0], keyPath[1], keyPath[2], v) {
		gotrace.Info("'%s' secret text not updated.", *s.key)
		return
	}
}

func (s *sSet) linkToFile(env string) {
	file := *s.linkFile

	if f, err := utils.Abs(file); err != nil {
		gotrace.Error("Unable to determine file path. %s", err)
		return
	} else {
		file = f
	}

	keyPath := strings.Split(*s.key, "/")

	v := creds.NewObjectsValue("link-to", goforjj.NewValueStruct(""))
	v.AddResource("linked-to", file)
	v.AddResource("created-on", time.Now().String())

	if !s.secrets.SetObjectValue(env, "", keyPath[0], keyPath[1], keyPath[2], v) {
		gotrace.Info("'%s' secret text not updated.", *s.key)
		return
	}

}
