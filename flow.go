package main

import (
	"fmt"

	"github.com/forj-oss/forjj-modules/trace"
)

// FlowInit load the flow in memory,
func (a *Forj) FlowInit() error {
	// Load flows from Forjfile sources.
	return a.flows.Load(a.f.GetDeclaredFlows()...)
}

// FlowApply apply flows to Forjfile
// it updates Forjfile inMemory object data.
func (a *Forj) FlowApply() error {
	ffd := a.f.InMemForjfile()
	bInError := false
	defaultFlowToApply := "default"
	if v, found := a.f.Get("settings", "default", "flow"); found {
		defaultFlowToApply = v.GetString()
	}

	if err := a.flows.Apply(defaultFlowToApply, nil, ffd); err != nil { // Applying Flow to Forjfile
		gotrace.Error("Forjfile: %s", err)
		bInError = true
	}

	for _, repo := range ffd.Repos {
		flowToApply := defaultFlowToApply
		if repo.Flow.Name != "" {
			flowToApply = repo.Flow.Name
		}

		if err := a.flows.Apply(flowToApply, repo, ffd); err != nil { // Applying Flow to Forjfile repo
			gotrace.Error("Repo '%s': %s", repo.GetString("name"), err)
			bInError = true
		}
	}

	if bInError {
		return fmt.Errorf("Several errors has been detected when trying to apply flows on Repositories. %s", "Please review and fix them.")
	}

	return nil
}
