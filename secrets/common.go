package secrets

import (
	"github.com/alecthomas/kingpin"
)

type common struct {
	env       *string
	common    *bool
	secretKey *string
	infraPath *string
}

func (c *common) init(context *Context, cmd *kingpin.CmdClause, initCommon func(context *Context, cmd *kingpin.CmdClause)) {
	c.env = context.Flag("deploy-to",
		cmd.Flag("deploy-env", "forjj deployment environment used to query/manage secrets. You can set 'FORJJ_DEPLOY_ENV' as environment variable.").Envar("FORJJ_DEPLOY_ENV")).String()
	c.secretKey = cmd.Flag("secrets-key", "Base64 secrets symetric key. Note that the key is not stored.").Envar("FORJJ_SECRETS_KEY").String()
	c.common = cmd.Flag("common", "To manage global secrets used by all deployments. Global secrets doesn't overwrite specific deployment secrets.").Bool()
	
	initCommon(context, cmd)
}
