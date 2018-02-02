package forjfile

import "github.com/forj-oss/forjj-modules/trace"

// Model used by template to secure data.

type RepoModel struct {
	Apps map[string]RepoAppModel
	repo *RepoStruct
}

func (r *RepoModel)Get(field string) (_ string) {
	if r.repo == nil {
		return
	}
	return r.repo.GetString(field)
}

func (r *RepoModel)HasApps(rules ...string) (_ bool) {
		if r.repo == nil {
		return
	}
	if v, err := r.repo.HasApps(rules...) ; err != nil {
		gotrace.Error("%s", err)
	} else {
		return v
	}
	return
}

type RepoAppModel struct {
	Default bool
	AppName string
}

