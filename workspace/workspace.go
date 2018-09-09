package workspace

import (
	"forjj/forjfile"
	"strings"

	"github.com/alecthomas/kingpin"
)

// Workspace represents cli subcommand actions and options
type Workspace struct {
	workspaceCmd *kingpin.CmdClause
	context Context

	list wsList
	edit wsEdit
	set  wsSet
	unset wsUnset

}

func (w *Workspace) Init(app *kingpin.Application, data *forjfile.Workspace, isParsePhase func() bool) {
	if w == nil || app == nil {
		return
	}

	w.workspaceCmd = app.Command("workspace", "Manage forjj workspace data")
	w.context.init(isParsePhase)
	w.list.init(w.workspaceCmd, data)
	w.set.init(w.workspaceCmd, data)
	w.edit.init(w.workspaceCmd, data)
	w.unset.init(w.workspaceCmd, data)
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
