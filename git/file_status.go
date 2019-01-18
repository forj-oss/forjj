package git

// FileStatus define the GIT file status
type FileStatus struct {
	index    rune
	workTree rune
}

// Index returns the Index status of the file
func (s FileStatus) Index() rune {
	return s.index
}

// WorkTree returns the work tree status of the file
func (s FileStatus) WorkTree() rune {
	return s.workTree
}
