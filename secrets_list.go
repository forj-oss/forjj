package main

import (
	"fmt"
	"forjj/scandrivers"
	"forjj/utils"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

type secretsList struct {
	cmd      *kingpin.CmdClause
	show     *bool
	elements map[string]secretInfo
	common   *secretsCommon
}

func (l *secretsList) init(parentCmd *kingpin.CmdClause, common *secretsCommon) {
	l.cmd = parentCmd.Command("list", "Show all credentials of the factory").Default()
	l.show = l.cmd.Flag("show", "Show password unencrypted.").Bool()
	l.common = common
}

// Display the list of secrets
func (l *secretsList) showList() {
	ffd := forj_app.f.InMemForjfile()

	scan := scandrivers.NewScanDrivers(ffd, forj_app.drivers)
	l.elements = make(map[string]secretInfo)

	// Retrieve secrets
	scan.SetScanObjFlag(func(objectName, instanceName, flagPrefix, name string, flag goforjj.YamlFlag) error {
		if flag.Options.Secure {
			info := secretInfo{}
			info.keyPath = objectName + "/" + instanceName + "/"
			keyName := name
			if flagPrefix != "" {
				keyName = flagPrefix + name
			}
			info.keyPath += keyName

			if *l.common.common {
				info.value, info.found, info.source, info.env = forj_app.s.GetGlobalString(objectName, instanceName, keyName)
			} else {
				info.value, info.found, info.source, info.env = forj_app.s.GetString(objectName, instanceName, keyName)
			}

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
			len(secretValue.source),
			len(secretValue.env),
			len(value))
	}

	fmt.Printf("List of secrets in forjj: (Deployment environment = '%s')\n\n", forj_app.f.GetDeployment())

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
