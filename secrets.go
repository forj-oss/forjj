package main

import (
	"fmt"
	"forjj/scandrivers"
	"forjj/utils"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/cli/interface"
	"github.com/forj-oss/forjj-modules/cli/kingpinCli"
	"github.com/forj-oss/goforjj"
)

type secrets struct {
	secrets     *kingpin.CmdClause
	params      map[string]cli.ForjParam
	env         *string
	secretKey   *string
	infraPath   *string
	cli_context clier.ParseContexter

	list struct {
		cmd      *kingpin.CmdClause
		show     *bool
		elements map[string]secretInfo
	}

	get struct {
		cmd *kingpin.CmdClause
		key *string
	}

	set struct {
		cmd      *kingpin.CmdClause
		key      *string
		password *string
	}

	unset struct {
		cmd *kingpin.CmdClause
		key *string
	}
}

func (s *secrets) init(app *kingpin.Application) {
	if s == nil || app == nil {
		return
	}

	s.params = make(map[string]cli.ForjParam)

	s.secrets = app.Command("secrets", "Manage forjj secrets")
	s.env = s.secrets.Flag("env", "Environ used to query/manage secrets. Use 'common' for all common secrets.").String()
	s.secretKey = s.secrets.Flag("secrets-key", "Base64 secrets symetric key. Note that the key is not stored.").Envar("FORJJ_SECRETS_KEY").String()
	s.infraPath = s.flag(infra_path_f, s.secrets.Flag(infra_path_f, infra_path_help)).Envar("FORJJ_INFRA").Short('W').String()
	s.flag("contribs-repo", s.secrets.Flag("contribs-repo", contribs_repo_help)).Envar("CONTRIBS_REPO").Default(defaultContribsRepo).String()
	s.flag("flows-repo", s.secrets.Flag("flows-repo", flows_repo_help)).Envar("FLOWS_REPO").Default(defaultFlowRepo).String()
	s.flag("repotemplates-repo", s.secrets.Flag("repotemplates-repo", repotemplates_repo_help)).Envar("REPOTEMPLATES_REPO").Default(defaultRepoTemplate).String()

	s.list.cmd = s.secrets.Command("list", "Show all credentials of the factory").Default()
	s.list.show = s.list.cmd.Flag("show", "Show password unencrypted.").Bool()

	s.get.cmd = s.secrets.Command("get", "Get value of a credential unencrypted")
	s.get.key = s.get.cmd.Arg("key", "Full key path").Required().String()

	s.set.cmd = s.secrets.Command("set", "store a new credential in forjj secrets")
	s.set.key = s.set.cmd.Arg("key", "Full key path").Required().String()
	s.set.password = s.set.cmd.Flag("password", "Secret key value").Short('P').String()

	s.unset.cmd = s.secrets.Command("unset", "remove a credential key path value from forjj secrets")
	s.unset.key = s.unset.cmd.Arg("key", "Full key path").Required().String()

}

func (s *secrets) flag(name string, flag *kingpin.FlagClause) (cliFlag *kingpinCli.FlagClause) {
	if s == nil {
		return nil
	}

	cliFlag = kingpinCli.NewFlag(flag)
	param := cli.NewForjFlag(cliFlag)
	s.params[name] = param
	return
}

func (s *secrets) defineContext(context clier.ParseContexter) {
	if s == nil {
		return
	}
	s.cli_context = context
}

func (s *secrets) getContextFlagValue(name string) (interface{}, bool) {
	if s == nil {
		return nil, false
	}
	param, found := s.params[name]
	if !found {
		return nil, false
	}
	return param.GetContextValue(s.cli_context)
}

// GetStringValue return value and status where the value were found.
//
// WARNING: Default status can be set only in cli load context phase (before parse)
// If we need to incorporate some data feed between real value and default value
// it must be done and saved during load context phase. (ie ParseContext() in cli_context.go)
func (s *secrets) GetStringValue(field string) (value string, found, isDefault bool, _ error) {
	var param cli.ForjParam

	param, found = s.params[field]
	if !found {
		return
	}

	var v interface{}
	if !forj_app.cli.IsParsePhase() {
		v, found = param.GetContextValue(s.cli_context)
		if !found {
			return
		}
		if fieldValue, ok := v.(string); ok {
			value = fieldValue
		} else if fieldDefault, ok := v.(*string); ok {
			value = *fieldDefault
			isDefault = true
		}
	} else {
		found = param.IsFound()
		if !found {
			return
		}
		value = param.GetStringValue()
	}
	return
}

func (s *secrets) action(action string) {
	actions := strings.Split(action, " ")
	switch actions[1] {
	case "list":
		s.showList()
	case "set":
	case "unset":
	case "show":
	}
}

// Display the list of secrets
func (s *secrets) showList() {
	ffd := forj_app.f.InMemForjfile()

	scan := scandrivers.NewScanDrivers(ffd, forj_app.drivers)
	s.list.elements = make(map[string]secretInfo)

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

			info.value, info.found, info.source, info.env = forj_app.s.GetString(objectName, instanceName, keyName)

			s.list.elements[info.keyPath] = info
		}
		return nil
	})
	scan.DoScanDriversObject()

	// Create terminal array
	array := utils.NewTerminalArray(len(s.list.elements), 4)

	// Define Columns
	array.SetCol(0, "Path")
	array.SetCol(1, "Environment")
	array.SetCol(2, "Source")
	array.SetCol(3, "Secret")

	// Evaluate Array size
	value := "***"
	for secretPath, secretValue := range s.list.elements {
		if *s.list.show {
			value = strings.Replace(secretValue.value, "\n", "", -1)
		}
		array.EvalLine(secretPath,
			len(secretPath),
			len(secretValue.source),
			len(secretValue.env),
			len(value))
	}

	fmt.Print("List of secrets in forjj:\n\n")

	// Print the array
	iFound := 0
	iTotal := 0
	array.Print(
		func(key string, compressedMax int) []interface{} {
			secretValue, found := s.list.elements[key]
			if !found {
				return nil
			}

			iTotal++
			value := ""
			if secretValue.found {
				value = "***"
				if *s.list.show {
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
