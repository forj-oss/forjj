package git

import (
	"errors"
	"fmt"
)

// Status contains a representation of GIT status in porcelain mode.
type Status struct {
	Ready    gitFiles
	NotReady gitFiles
	files    gitFilesStatus
	Err      error
}

// Files return all files updated identified by git status
func (gs *Status) Files() (files []string) {
	files = make([]string, gs.CountFiles())

	files = append(files, gs.Ready.Files()...)
	files = append(files, gs.NotReady.Files()...)
	return
}

// CountFiles returns the number of files updated tracked or not.
func (gs *Status) CountFiles() int {
	return gs.Ready.CountFiles() + gs.NotReady.CountFiles()
}

// Tracked return Tracked files
func (gs *Status) Tracked() (files []string) {
	files = make([]string, gs.CountTracked())

	files = append(files, gs.Ready.Tracked()...)
	files = append(files, gs.NotReady.Tracked()...)
	return
}

// CountTracked returns the number of tracked files updated in ready or not ready area.
func (gs *Status) CountTracked() int {
	return gs.Ready.CountTracked() + gs.NotReady.CountTracked()
}

// Untracked return Tracked files
func (gs *Status) Untracked() (files []string) {
	files = make([]string, gs.CountUntracked())

	files = append(files, gs.Ready.Untracked()...)
	files = append(files, gs.NotReady.Untracked()...)
	return
}

// CountUntracked returns the number of tracked files updated in ready or not ready area.
func (gs *Status) CountUntracked() int {
	return gs.NotReady.CountUntracked()
}

// GetFile return a status file object
func (gs *Status) GetFile(file string) (statusFile FileStatus, err error) {
	if gs == nil {
		err = errors.New("Status object is nil")
		return
	}
	if s, found := gs.files[file]; found {
		statusFile = s
	} else {
		err = fmt.Errorf("%s not found in GIT status", file)
	}
	return
}
