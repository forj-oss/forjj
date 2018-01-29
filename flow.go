package main

// FlowStart load the flow in memory,
// configure it with Forjfile information
// and update Forjfile inMemory object data.
func (a *Forj)FlowInit() error {
	if err := a.flows.Load(a.f.GetDeclaredFlows() ...) ; err != nil {
		return err
	}
	for _, repo := range a.f.Forjfile().Repos {
		flow_to_apply := "default"
		if v, found := a.f.Get("settings", "default", "flow") ; found {
			flow_to_apply = v.GetString()
		}
		if repo.Flow.Name != "" {
			flow_to_apply = repo.Flow.Name
		}

		if err := a.flows.Apply(flow_to_apply, repo, &a.f) ; err != nil {// Applying Flow to Forjfile
			return err
		}
	}

	return nil
}

