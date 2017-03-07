package utils

import (
	"strings"
	"os"
	"fmt"
)

func DefColor(colorNum int) (color, reset string) {
	if strings.HasPrefix(os.Getenv("TERM"), "xterm") {
		color = fmt.Sprintf("\x1b[%dm", colorNum)
		reset = "\x1b[0m"
	}
	return
}

