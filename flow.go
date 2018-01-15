package main

import "github.com/forj-oss/forjj-modules/trace"

// FlowStart load the flow in memory,
// configure it with Forjfile information
// and update Forjfile inMemory object data.
func (a *Forj)FlowInit() error {
	var flows []string
	if f := a.f.GetDeclaredFlows() ; len(f) < 0 {
		gotrace.Trace("No flows detected. Flow implementation bypassed.")
		return nil
	} else {
		flows = f
	}
	if err := a.flows.Load(flows ...) ; err != nil {
		return err
	}
	return nil
}

