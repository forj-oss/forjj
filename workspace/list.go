package workspace

import (
	"fmt"
	"forjj/forjfile"
	"forjj/utils"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
)

type wsList struct {
	cmd  *kingpin.CmdClause
	show *bool
	data *forjfile.Workspace
}

func (l *wsList) init(parentCmd *kingpin.CmdClause, data *forjfile.Workspace) {
	l.cmd = parentCmd.Command("list", "Show all credentials of the factory").Default()
	l.show = l.cmd.Flag("show", "Show password unencrypted.").Bool()

	l.data = data
}

// Display the list of secrets
func (l *wsList) showList() {
	// Create terminal array
	array := utils.NewTerminalArray(l.data.Len(), 2)

	// Define Columns
	array.SetCol(0, "Data")
	array.SetCol(1, "Value")

	// Evaluate Array size
	data := l.data.Data()
	for dataPath, dataValue := range data {
		array.EvalLine(dataPath,
			len(dataPath),
			len(dataValue))
	}

	fmt.Print("List of forjj workspace data: \n\n")

	// Print the array
	iTotal := 0
	array.Print(
		func(key string, compressedMax int) []interface{} {
			value, found := data[key]
			if !found {
				return nil
			}

			iTotal++
			return []interface{}{
				key,
				utils.StringCompress(value, 0, compressedMax),
			}
		},
	)

	gotrace.Info("%d workspace data found", iTotal)

}
