package utils

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
)

// Abs returns an absolute path
//
//	It expands ~ to $HOME
//	If path is already absolute, it cleans up (shortest path).
//	If the path is relative, it adds current directory to create it as absolute path
func Abs(name string) (string, error) {
	// Check in case of paths like "/something/~/something/"
	if len(name) > 2 && name[:2] == "~/" {
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

func InStringList(element string, elements ...string) string {
	for _, value := range elements {
		if element == value {
			return value
		}
	}
	return ""
}

func ArrayStringDelete(a []string, element string) []string {
	for index, value := range a {
		if value == element {
			return append(a[:index], a[index+1:]...)
		}
	}
	return a
}

const StringCompressMin = 7

// StringCompress compress a string to display to a maximum given
// if the str is higher, it will be cut in the middle with "..."
// min must be higher than 7 bytes.
// If max is lower than min, max will be set to min to ensure the size returned is never
// under min or 7 bytes
func StringCompress(str string, min, max int) (ret string) {
	if min < StringCompressMin {
		min = StringCompressMin
	}
	if max < min {
		max = min
	}

	size := len(str)
	if size <= max {
		return str
	}
	midMin := max/2 - 2
	midMax := max/2 - 1
	ret = str[0:midMin] + "..." + str[size-midMax:]
	return
}

// EnsureDir ensure the given dir exist and recreate it if needed.
func EnsureDir(dir, mess string) (_ error) {
	if info, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("Unable to create the %s %s. %s", mess, dir, err)
		}
	} else if !info.Mode().IsDir() {
		return fmt.Errorf("Unable to use the %s %s. It exists but not as a directory", mess, dir)
	}
	return
}
