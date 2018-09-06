package workspace

import (
	"strings"

	"github.com/alecthomas/kingpin"
)

type Workspace struct {
	workspaceCmd *kingpin.CmdClause
	context Context
	common  wsCommon

	list wsList
	edit wsEdit
	set  wsSet
	unset wsUnset

}

func (w *Workspace) init(app *kingpin.Application) {
	if w == nil || app == nil {
		return
	}

	w.secrets = app.Command("workspace", "Manage forjj workspace data")
	w.context.init()
	w.common.init(&s.context, s.secrets)
	w.list.init(s.secrets, &s.common)
	w.set.init(s.secrets, &s.common)
	w.edit.init(s.secrets, &s.common)

	w.get.cmd = s.secrets.Command("get", "Get value of a credential unencrypted")
	w.get.key = s.get.cmd.Arg("key", "Full key path").Required().String()

	w.unset.init(s.secrets, &s.common)
}

func (w *Workspace) action(action string) {
	actions := strings.Split(action, " ")
	switch actions[1] {
	case "list":
		w.list.showList()
	case "set":
		w.set.doSet()
	case "edit":
		w.edit.doEdit()
	case "unset":
		w.unset.doUnset()
	}
}
