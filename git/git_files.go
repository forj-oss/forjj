package git

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
	gs["R"] = make([]string, 0, 2)
	if untracked {
		gs["?"] = make([]string, 0, 2)
	}
}

func (gs gitFiles) add(fileStatus, file string) {
	list := gs[fileStatus]
	list = append(list, file)
	gs[fileStatus] = list
}
