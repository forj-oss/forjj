package repository

import (
	"path"
	"forjj/git"
	"fmt"
	"os"
)

type GitRepoStruct struct {
	path string
	err error
}

func (i *GitRepoStruct)Create(repo_path string, initial_commit func() ([]string, error), force_create bool) error {
	i.path = path.Clean(repo_path)

	if creatable := i.is_creatable() ; !creatable {
		if force_create {
			return i.use()
		}
		return i.err
	}

	if git.Do("init", i.path) > 0 {
		return fmt.Errorf("Unable to initialize %s", i.path)
	}

	if err := os.Chdir(i.path); err != nil {
		return fmt.Errorf("Unable to move repository at %s. %s", i.path, err)
	}

	if ! i.git_1st_commit_exist("master") {
		i.git_1st_commit(initial_commit)
	}
	return nil
}

func (i *GitRepoStruct)Use(repo_path string) error {
	i.path = path.Clean(repo_path)

	return i.use()
}

func (i *GitRepoStruct)use() error {
	if ! i.is_valid() {
		return i.err
	}
	if ! i.git_1st_commit_exist("master") {
		return fmt.Errorf("%s do not have the initial commit. You need to use create to create it first.", i.path)
	}
	return nil
}
