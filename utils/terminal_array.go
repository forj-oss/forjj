package utils

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

type TerminalArray struct {
	cols       []string
	max        *EvalValues
	formatLine string
	formatSep  string
	sortedList []string
	sortIndex  int
}

func NewTerminalArray(linesNum, colsNum int) (ret *TerminalArray) {
	if colsNum <= 0 {
		return nil
	}

	ret = new(TerminalArray)
	ret.cols = make([]string, colsNum)
	ret.max = NewEvalValues(colsNum + 1)
	ret.sortedList = make([]string, linesNum)
	return
}

func (t *TerminalArray) SetCol(index int, header string) {
	if t == nil {
		return
	}
	if index < 0 || index >= len(t.cols) {
		return
	}

	t.cols[index] = header
	t.max.Eval(index, len(header))
}

func (t *TerminalArray) EvalLine(key string, cols ...int) {
	for index, length := range cols {
		if t.sortIndex == index {
			t.sortedList[t.max.CountOf(index)-1] = key
		}
		t.max.Eval(index, length)
	}
}

func (t *TerminalArray) Print(getLineData func(key string, compressedMax int) []interface{}) {
	sort.Strings(t.sortedList)

	colSize := 3

	// Evaluate terminal width
	stdout := int(os.Stdout.Fd())
	var terminalMax int
	if terminal.IsTerminal(stdout) {
		terminalMax, _, _ = terminal.GetSize(stdout)
	}
	if terminalMax < 80 {
		terminalMax = 80
	}

	// Evaluate last column compressed size
	lineSize := 0
	for index := range t.cols {
		lineSize += t.max.ValueOf(index)
	}
	lineSize += colSize * (len(t.cols) - 1)

	compressedIndex := len(t.cols)
	if lineSize > terminalMax {
		t.max.Eval(compressedIndex, terminalMax-(lineSize-t.max.ValueOf(compressedIndex-1)))
		t.max.Eval(compressedIndex, StringCompressMin)
	} else {
		t.max.Eval(compressedIndex, t.max.ValueOf(compressedIndex-1))
	}

	// Define line and separator Format
	colsNum := len(t.cols)
	for index := range t.cols {
		colFormat := "%%-%ds"
		if index == colsNum-1 {
			format := fmt.Sprintf(colFormat, t.max.ValueOf(index+1))
			t.formatLine += format
			t.formatSep += format
		} else {
			format := fmt.Sprintf(colFormat, t.max.ValueOf(index))
			t.formatLine += format + " | "
			t.formatSep += format + "-+-"
		}
	}

	// Define Separator line
	values := make([]interface{}, len(t.cols))
	for index := range t.cols {
		if index == colsNum-1 {
			values[index] = strings.Repeat("-", t.max.ValueOf(index+1))
		} else {
			values[index] = strings.Repeat("-", t.max.ValueOf(index))
		}
	}
	sepLine := fmt.Sprintf(t.formatSep+"\n", values...)

	// Print Header
	for index, value := range t.cols {
		values[index] = value
	}
	fmt.Printf(t.formatLine+"\n", values...)

	// Print Separator
	fmt.Println(sepLine)

	// Print Data
	for _, key := range t.sortedList {
		fmt.Printf(t.formatLine+"\n", getLineData(key, t.max.ValueOf(compressedIndex))...)
	}

	// Print Separator
	fmt.Println(sepLine)
}
