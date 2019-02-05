package secrets

import (
	"fmt"
	"forjj/creds"
	"forjj/drivers"
	"forjj/forjfile"
	"forjj/scandrivers"
	"forjj/utils"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

type sList struct {
	cmd      *kingpin.CmdClause
	show     *bool
	elements map[string]sInfo
	common   *common

	forjfile *forjfile.Forge
	drivers  *drivers.Drivers
	secrets  *creds.Secure
}

func (l *sList) init(parentCmd *kingpin.CmdClause, common *common, forjfile *forjfile.Forge, drivers *drivers.Drivers, secrets *creds.Secure) {
	l.cmd = parentCmd.Command("list", "Show all credentials of the factory").Default()
	l.show = l.cmd.Flag("show", "Show password unencrypted.").Bool()
	l.common = common

	l.forjfile = forjfile
	l.drivers = drivers
	l.secrets = secrets
}

// Display the list of secrets
func (l *sList) showList() {
	ffd := l.forjfile.InMemForjfile()

	scan := scandrivers.NewScanDrivers(ffd, l.drivers)
	l.elements = make(map[string]sInfo)

	// Retrieve secrets
	scan.SetScanObjFlag(func(objectName, instanceName, flagPrefix, name string, flag goforjj.YamlFlag) error {
		if flag.Options.Secure {
			info := sInfo{}
			info.keyPath = objectName + "/" + instanceName + "/"
			keyName := name
			if flagPrefix != "" {
				keyName = flagPrefix + name
			}
			info.keyPath += keyName

			var value *creds.Value
			if *l.common.common {
				value, info.found, _, info.env = l.secrets.GetGlobal(objectName, instanceName, keyName)
			} else {
				value, info.found, _, info.env = l.secrets.Get(objectName, instanceName, keyName)
			}

			if v, err := value.GetString(); err != nil {
				info.value = fmt.Sprintf("Warning! %s", err)
			} else {
				info.value = v
			}
			info.source = value.GetSource()

			l.elements[info.keyPath] = info
		}
		return nil
	})
	scan.DoScanDriversObject()

	// Create terminal array
	array := utils.NewTerminalArray(len(l.elements), 4)

	// Define Columns
	array.SetCol(0, "Path")
	array.SetCol(1, "Environment")
	array.SetCol(2, "Source")
	array.SetCol(3, "Secret")

	// Evaluate Array size
	value := "***"
	for secretPath, secretValue := range l.elements {
		if *l.show {
			value = strings.Replace(secretValue.value, "\n", "", -1)
		}
		array.EvalLine(secretPath,
			len(secretPath),
			len(secretValue.env),
			len(secretValue.source),
			len(value))
	}

	fmt.Printf("List of secrets in forjj: (Deployment environment = '%s')\n\n", l.forjfile.GetDeployment())

	// Print the array
	iFound := 0
	iTotal := 0
	array.Print(
		func(key string, compressedMax int) []interface{} {
			secretValue, found := l.elements[key]
			if !found {
				return nil
			}

			iTotal++
			value := ""
			if secretValue.found {
				value = "***"
				if *l.show {
					value = strings.Replace(secretValue.value, "\n", "", -1)
				}

				iFound++
			}
			return []interface{}{
				key,
				secretValue.env,
				secretValue.source,
				utils.StringCompress(value, 0, compressedMax),
			}
		},
	)

	gotrace.Info("%d/%d secrets found", iFound, iTotal)

}
