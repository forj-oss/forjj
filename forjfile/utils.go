package forjfile

import (
	"os"
	"path"
)


func Abs(name string) (string, error) {
	if path.IsAbs(name) {
		return name, nil
	}
	wd, err := os.Getwd()
	return path.Clean(path.Join(wd, name)), err
}
