package git

// Status contains a representation of GIT status in porcelain mode.
type Status struct {
	Ready    gitFiles
	NotReady gitFiles
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

type gitFiles map[string][]string

// Files returns the list of files identified for the GIT area choosen.
func (gs gitFiles) Files() (files []string) {
	files = make([]string, gs.CountFiles())

	count := 0
	for _, list := range gs {
		for _, file := range list {
			files[count] = file
			count++
		}
		count += len(list)
	}
	return
}

// CountTracked returns the list of files identified and tracked for the GIT area choosen.
func (gs gitFiles) CountFiles() (count int) {
	count = 0
	for _, list := range gs {
		count += len(list)
	}
	return
}

// Tracked returns the list of files identified and tracked for the GIT area choosen.
func (gs gitFiles) Tracked() (files []string) {
	files = make([]string, gs.CountTracked())

	count := 0
	for fileStatus, list := range gs {
		if fileStatus == "?" {
			continue
		}
		for _, file := range list {
			files[count] = file
			count++
		}
		count += len(list)
	}
	return
}

// Tracked returns the list of files identified and tracked for the GIT area choosen.
func (gs gitFiles) Untracked() (files []string) {
	files = make([]string, gs.CountUntracked())

	count := 0
	for fileStatus, list := range gs {
		if fileStatus != "?" {
			continue
		}
		for _, file := range list {
			files[count] = file
			count++
		}
		count += len(list)
	}
	return
}

// CountTracked returns the list of files identified and tracked for the GIT area choosen.
func (gs gitFiles) CountTracked() (count int) {
	count = 0
	for fileStatus, list := range gs {
		if fileStatus == "?" {
			continue
		}
		count += len(list)
	}
	return
}

// CountTracked returns the list of files identified and tracked for the GIT area choosen.
func (gs gitFiles) CountUntracked() (count int) {
	count = 0
	for fileStatus, list := range gs {
		if fileStatus != "?" {
			continue
		}
		count += len(list)
	}
	return
}

// Internal functions

func (gs gitFiles) init(untracked bool) {
	gs["A"] = make([]string, 0, 2)
	gs["M"] = make([]string, 0, 2)
	gs["D"] = make([]string, 0, 2)
	if untracked {
		gs["?"] = make([]string, 0, 2)
	}
}

func (gs gitFiles) add(fileStatus, file string) {
	list := gs[fileStatus]
	list = append(list, file)
	gs[fileStatus] = list
}
