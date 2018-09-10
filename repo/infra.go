package repository

import (
	"fmt"
	"forjj/git"
	"os"
	"path"
)

type GitRepoStruct struct {
	path string
	err  error
}

// Create the infra repository. Used at forjj create time.
func (i *GitRepoStruct) Create(repo_path string, initial_commit func() ([]string, error), force_create bool) error {
	i.path = path.Clean(repo_path)

	if creatable := i.isCreatable(); !creatable {
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

	if !i.git1stCommitExist("master") {
		return i.git1stCommit(initial_commit)
	}
	return nil
}

// Use re-use an existing repository (in cur dir) and check if this repo contains at least one Forjfile. Forjj will detect later if this Forjfile is not a Template one to exit.
func (i *GitRepoStruct) Use(repo_path string) error {
	i.path = path.Clean(repo_path)

	return i.use()
}

func (i *GitRepoStruct) use() error {
	if !i.isValid() {
		return i.err
	}
	if !i.masterForjfileControlled() {
		return fmt.Errorf("%s do not have the initial commit. You need to use create to create it first.", i.path)
	}
	return nil
}
