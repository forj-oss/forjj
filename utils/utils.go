package utils

import (
	"os"
	"path"
	"os/user"
	"path/filepath"
)

// Abs returns an absolute path
//
//	It expands ~ to $HOME
//	If path is already absolute, it cleans up (shortest path).
//	If the path is relative, it adds current directory to create it as absolute path
func Abs(name string) (string, error) {
	// Check in case of paths like "/something/~/something/"
	if len(name) >2 && name[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		dir := usr.HomeDir
		name = filepath.Join(dir, name[2:])
	}
	if path.IsAbs(name) {
		return path.Clean(name), nil
	}
	wd, err := os.Getwd()
	return path.Clean(path.Join(wd, name)), err
}

func Touch(file string) error {
	if fd, err := os.Create(file); err == nil {
		return fd.Close()
	} else {
		return err
	}
}

