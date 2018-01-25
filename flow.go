package main

// FlowStart load the flow in memory,
// configure it with Forjfile information
// and update Forjfile inMemory object data.
func (a *Forj)FlowInit() error {
	if err := a.flows.Load(a.f.GetDeclaredFlows() ...) ; err != nil {
		return err
	}
	a.flows.Start() // Applying Flow to Forjfile
	return nil
}

