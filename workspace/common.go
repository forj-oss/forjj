package workspace

import (
	"github.com/alecthomas/kingpin"
)

type wsCommon struct {
}

func (c *wsCommon) init(context *Context, cmd *kingpin.CmdClause, initCommon func(context *Context, cmd *kingpin.CmdClause)) {
	initCommon(context, cmd)
}
