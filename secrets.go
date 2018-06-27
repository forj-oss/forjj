package main

import (
	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/cli"
	"github.com/forj-oss/forjj-modules/cli/interface"
	"github.com/forj-oss/forjj-modules/cli/kingpinCli"
)


type secrets struct {
	secrets   *kingpin.CmdClause
	params    map[string]cli.ForjParam
	env       *string
	secretKey *string
	infraPath *string
	cli_context clier.ParseContexter

	list struct {
		cmd  *kingpin.CmdClause
		show *bool
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

func (s *secrets) flag(name string, flag *kingpin.FlagClause) *kingpin.FlagClause {
	if s == nil {
		return nil
	}

	param := cli.NewForjFlag(kingpinCli.NewFlag(flag))
	s.params[name] = param
	return flag
}

func (s *secrets) defineContext(context clier.ParseContexter) {
	if s == nil {
		return
	}
	s.cli_context = context
}

func (s *secrets) getContextFlagValue(name string) (interface{}, bool){
	if s == nil {
		return nil, false
	}
	param, found := s.params[name]
	if ! found {
		return nil, false
	}
	return param.GetContextValue(s.cli_context)
}