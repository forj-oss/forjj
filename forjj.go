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

var (
	forj_app Forj
	build_branch string
	build_commit string
	build_date string
	build_tag string
)

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
	parse, err := forj_app.cli.Parse(os.Args[1:], nil)

	// Check initial requirement for forjj create
	/*	if parse == "create" {
		if found, _ := forj_app.w.check_exist(); found {
			log.Fatalf("Unable to create the workspace '%s'. Already exist.", forj_app.w.Path())
		}
	}*/
	if err == nil && forj_app.w.Error() != nil {
		kingpin.Fatalf("Unable to go on. %s", forj_app.w.Error())
	}

	//	TODO : Use cli : Re-apply following function
	// forj_app.InitializeDriversAPI()
	defer forj_app.driver_cleanup_all()
	action := kingpin.MustParse(parse, err)
	if f, found := forj_app.actionDispatch[forj_app.contextAction] ; found {
		f(action)
	}
}

func (a *Forj) contextDisplayed() {
	tmpl_file := a.f.GetForjfileTemplateFileLoaded()
	file := a.f.GetForjfileFileLoaded()
	if tmpl_file != "" || file != "" {
		msg := "Forjj has loaded the following context:\n"
		if tmpl_file != "" {
			msg += "Template Forjfile: " + tmpl_file + "\n"
		}
		if file != "" {
			msg += "Repository Forjfile: " + file + "\n"
		}
		msg += "-------------------------------------\n"
		gotrace.Warning(msg)
	}
}
