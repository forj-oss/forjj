package main

// Call docker to create the Solution source code from scratch with validated parameters.
// This container do the real stuff (git/call drivers)
// I would expect to have this go tool to have a do_create to replace the shell script.
// But this would be a next version and needs to be validated before this decision is made.
func (a *Forj) Maintain() {
    // Load Workspace information
    a.w.Load(a)

    // Read forjj infra file and the options file given
    a.LoadForjjPluginsOptions()

    // Loop on each drivers to ask them to maintain the infra, it controls.

    // Loop on upstream repositories to ensure it exists with at least a README.md file.

    println("FORJJ - maintain", a.w.Organization, "DONE") // , cmd.ProcessState.Sys().WaitStatus)
}
