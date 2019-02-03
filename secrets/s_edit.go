package secrets

import (
	"forjj/creds"
	"forjj/drivers"
	"forjj/forjfile"
	"forjj/scandrivers"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/alecthomas/kingpin"
	gotrace "github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

type sEdit struct {
	editor          *string
	noTerminalSetup *bool
	cmd             *kingpin.CmdClause
	key             *string
	password        string
	common          *common

	elements map[string]sInfo

	forjfile *forjfile.Forge
	drivers  *drivers.Drivers
	secrets  *creds.Secure
}

func (e *sEdit) init(parent *kingpin.CmdClause, common *common, forjfile *forjfile.Forge, drivers *drivers.Drivers, secrets *creds.Secure) {

	e.cmd = parent.Command("edit", "edit a credential stored in forjj secrets")
	e.key = e.cmd.Arg("key", "Key path. Format is <objectType>/<objectInstance>/<key>.)").Required().String()
	e.editor = e.cmd.Flag("editor", "editor to execute").Envar("EDITOR").Default("/usr/bin/vi").String()
	e.noTerminalSetup = e.cmd.Flag("no-terminal-setup", "do not set terminal i/o to edit the credential.").Short('T').Bool()
	e.common = common

	e.forjfile = forjfile
	e.drivers = drivers
	e.secrets = secrets
}

// doEdit register a password to the path given.
// Only supported path are recognized.
func (e *sEdit) doEdit() {
	ffd := e.forjfile.InMemForjfile()

	scan := scandrivers.NewScanDrivers(ffd, e.drivers)
	e.elements = make(map[string]sInfo)

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

			info.value, info.found, info.source, info.env = e.secrets.GetString(objectName, instanceName, keyName)

			e.elements[info.keyPath] = info
		}
		return nil
	})
	scan.DoScanDriversObject()

	if _, found := e.elements[*e.key]; !found {
		gotrace.Error("'%s' is not a valid secret path. check with `forjj secrets`", *e.key)
		return
	}

	if *e.editor == "" {
		*e.editor = "/usr/bin/vi"
	}

	tmpFile, err := ioutil.TempFile("/tmp", "forjj-")
	if err != nil {
		gotrace.Error("Unable to create temporary file in /tmp. %s. Exiting.", err)
		return
	}

	fileName := tmpFile.Name()
	defer os.Remove(fileName)

	keyPath := strings.Split(*e.key, "/")

	e.password, _, _, _ = e.secrets.GetString(keyPath[0], keyPath[1], keyPath[2])

	_, err = tmpFile.WriteString(e.password)
	if err != nil {
		gotrace.Error("Unable to write temporary file in /tmp. %s. Exiting.", err)
		return
	}
	tmpFile.Close()

	cmd := exec.Command(*e.editor, fileName)
	if !*e.noTerminalSetup {
		if !terminal.IsTerminal(int(os.Stdin.Fd())) {
			gotrace.Error("Unable to edit %s. Not a terminal. Exiting.", *e.key)
			return
		}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err = cmd.Run()

	if err != nil {
		gotrace.Error("Unable to start the editor %s on file %s. %s. Exiting.", *e.editor, fileName, err)
		return
	}

	if data, err := ioutil.ReadFile(fileName); err != nil {
		gotrace.Error("Unable to read the editor file %s. %s. Exiting.", fileName, err)
		return
	} else {
		e.password = strings.Trim(string(data), " \n")
	}

	if len(e.password) == 0 {
		gotrace.Info("File is empty. Update ignored.")
		return
	}

	v := creds.NewObjectsValue("forjj", goforjj.NewValueStruct(e.password))
	env := e.forjfile.GetDeployment()
	if *e.common.common {
		env = creds.Global
	}
	if !e.secrets.SetObjectValue(env, "forjj", keyPath[0], keyPath[1], keyPath[2], v) {
		gotrace.Info("'%s' secret text not updated.", *e.key)
		return
	}

	e.secrets.SaveEnv(env)
	gotrace.Info("'%s' secret text saved in '%s' deployment environment.", *e.key, env)
}
