package main

import (
	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/cli/interface"
)

type secretsCommon struct {
	env         *string
	secretKey   *string
	infraPath   *string
	cli_context clier.ParseContexter
}

func (c *secretsCommon) init(context *secretsContext, cmd *kingpin.CmdClause) {
	c.env = cmd.Flag("env", "Environ used to query/manage secrets. Use 'common' for all common secrets.").String()
	c.secretKey = cmd.Flag("secrets-key", "Base64 secrets symetric key. Note that the key is not stored.").Envar("FORJJ_SECRETS_KEY").String()
	c.infraPath = context.flag(infra_path_f, cmd.Flag(infra_path_f, infra_path_help)).Envar("FORJJ_INFRA").Short('W').String()
	context.flag("contribs-repo", cmd.Flag("contribs-repo", contribs_repo_help)).Envar("CONTRIBS_REPO").Default(defaultContribsRepo).String()
	context.flag("flows-repo", cmd.Flag("flows-repo", flows_repo_help)).Envar("FLOWS_REPO").Default(defaultFlowRepo).String()
	context.flag("repotemplates-repo", cmd.Flag("repotemplates-repo", repotemplates_repo_help)).Envar("REPOTEMPLATES_REPO").Default(defaultRepoTemplate).String()
}
