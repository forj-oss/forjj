package main

import (
	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
	"log"
	"os"
)

// TODO: Implement RepoTemplates
// TODO: Implement Flow
// TODO: Call maintain to start the plugin provision container command.

var forj_app Forj

// Define the default Docker image to use for running forjj actions task by drivers.
const Docker_image = "docker.hos.hpecorp.net/devops/forjj"

// ************************** MAIN ******************************
func main() {

	debug := os.Getenv("FORJJ_DEBUG")
	if debug == "true" {
		log.Printf("Debug set to '%s'.\n", debug)
		gotrace.SetDebug()
	}

	forj_app.init()
	parse, err := forj_app.cli.Parse(os.Args[1:])

	// Check initial requirement for forjj create
	if parse == "create" {
		if found, _ := forj_app.w.check_exist(); found {
			log.Fatalf("Unable to create the workspace '%s'. Already exist.", forj_app.w.Path())
		}
	}

	forj_app.InitializeDriversFlag()
	defer forj_app.driver_cleanup_all()
	switch kingpin.MustParse(parse, err) {
	case "create":
		if err := forj_app.Create(); err != nil {
			log.Fatalf("Forjj create issue. %s", err)
		}
		log.Print("===========================================")
		if !*forj_app.no_maintain {
			log.Print("Source codes are in place. Now, starting instantiating your DevOps Environment services...")
			forj_app.do_maintain() // This will implement the flow for the infra-repo as well.
		} else {
			log.Print("Source codes are in place. Now, Please review commits, push and start instantiating your DevOps Environment services with 'forjj maintain' ...")
		}
		println("FORJJ - create ", forj_app.w.Organization, " DONE") // , cmd.ProcessState.Sys().WaitStatus)

	case "update":
		if err := forj_app.Update(); err != nil {
			log.Fatalf("Forjj update issue. %s", err)
		}
		println("FORJJ - update ", forj_app.w.Organization, " DONE") // , cmd.ProcessState.Sys().WaitStatus)

	case "maintain":
		if err := forj_app.Maintain(); err != nil {
			log.Fatalf("Forjj maintain issue. %s", err)
		}
		println("FORJJ - maintain ", forj_app.w.Organization, " DONE") // , cmd.ProcessState.Sys().WaitStatus)
	}
}
