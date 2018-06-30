package main

import (
	"fmt"
	"forjj/scandrivers"
	"forjj/utils"
	"os"
	"sort"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/cli/interface"
	"github.com/forj-oss/forjj-modules/cli/kingpinCli"
	"github.com/forj-oss/goforjj"
	"golang.org/x/crypto/ssh/terminal"
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

	scan.SetScanObjFlag(func(objectName, instanceName, flagPrefix, name string, flag goforjj.YamlFlag) error {
		if flag.Options.Secure {
			info := secretInfo{}
			info.keyPath = objectName + "/" + instanceName + "/"
			keyName := name
			if flagPrefix != "" {
				keyName = flagPrefix + name
			}
			info.keyPath += keyName

			info.value, info.found, info.source = forj_app.s.GetString(objectName, instanceName, keyName)

			s.list.elements[info.keyPath] = info
		}
		return nil
	})

	scan.DoScanDriversObject()

	gotrace.Trace("secrets elements found: %d", len(s.list.elements))

	stdin := int(os.Stdin.Fd())
	var terminalMax int
	if terminal.IsTerminal(stdin) {
		terminalMax, _, _ = terminal.GetSize(stdin)
	}
	if terminalMax < 80 {
		terminalMax = 80
	}

	sortedList := make([]string, len(s.list.elements))
	max := utils.NewEvalValues(5)
	max.Eval(0, len("path"))
	max.Eval(1, len("Source"))
	max.Eval(2, len("secret"))
	for secretPath, secretValue := range s.list.elements {
		sortedList[max.CountOf(0)-1] = secretPath
		max.Eval(0, len(secretPath))
		max.Eval(1, len(secretValue.source))

		if *s.list.show {
			max.Eval(2, len(strings.Replace(secretValue.value, "\n", "", -1)))
		}
		if secretValue.found {
			max.Eval(4, 1)
		}
	}

	colSize := 3
	if max.ValueOf(0)+max.ValueOf(1)+max.ValueOf(2)+colSize*2 > terminalMax {
		max.Eval(3, terminalMax-(max.ValueOf(0)+max.ValueOf(1)+colSize*2))
		max.Eval(3, utils.StringCompressMin)
	} else {
		max.Eval(3, max.ValueOf(2))
	}

	printFormat := max.PrintfFormat("%%-%ds | %%-%ds | %%-%ds\n", 0, 1, 3)
	tableFormat := max.PrintfFormat("%%%ds-+-%%%ds-+-%%%ds\n", 0, 1, 3)

	sort.Strings(sortedList)

	fmt.Print("List of secrets in forjj:\n\n")

	fmt.Printf(printFormat, "Path", "Source", "Secret")
	fmt.Printf(tableFormat,
		strings.Repeat("-", max.ValueOf(0)),
		strings.Repeat("-", max.ValueOf(1)),
		strings.Repeat("-", max.ValueOf(3)),
	)
	for _, secretPath := range sortedList {
		value := ""
		secretValue := s.list.elements[secretPath]
		if secretValue.found {
			if *s.list.show {
				value = utils.StringCompress(strings.Replace(secretValue.value, "\n", "", -1), 0, max.ValueOf(3))
			} else {
				value = "***"
			}

		}

		fmt.Printf(printFormat, secretPath, secretValue.source, value)
	}
	fmt.Printf(tableFormat,
		strings.Repeat("-", max.ValueOf(0)),
		strings.Repeat("-", max.ValueOf(1)),
		strings.Repeat("-", max.ValueOf(3)),
	)

	gotrace.Info("%d/%d secrets found", max.CountOf(4), len(s.list.elements))

}
