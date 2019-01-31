package workspace

import (
	"forjj/forjfile"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/cli/clier"
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

// Init configure the workspace cli command
func (w *Workspace) Init(app *kingpin.Application, data *forjfile.Workspace, isParsePhase func() bool, initCommon func(context *Context, cmd *kingpin.CmdClause)) {
	if w == nil || app == nil {
		return
	}

	w.workspaceCmd = app.Command("workspace", "Manage forjj workspace data")
	w.context.init(isParsePhase)
	initCommon(&w.context, w.workspaceCmd)
	w.list.init(w.workspaceCmd, data)
	w.set.init(w.workspaceCmd, data)
	w.edit.init(w.workspaceCmd, data)
	w.unset.init(w.workspaceCmd, data)
}

// Action executed for workspace cli command
func (w *Workspace) Action(action string) {
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
	default:
		w.list.showList()
	}
}

// DefineContext define cli Context to permit ParseContext to retrieve 
// common variable set.
func (w *Workspace) DefineContext(context clier.ParseContexter) {
	w.context.defineContext(context)
}

// GetStringValue Return a field value from the given context (parse time, or after)
func (w *Workspace) GetStringValue(field string) (value string, found, isDefault bool, _ error) {
	return w.context.GetStringValue(field)
}
