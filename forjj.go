package main

import (
        "os"
        "gopkg.in/alecthomas/kingpin.v2"
        "log"
)

var forj_app Forj

// Define the default Docker image to use for running forjj actions task by drivers.
const Docker_image = "docker.hos.hpecorp.net/devops/forjj"

// ************************** MAIN ******************************
func main() {
 forj_app.init()
 parse, err := forj_app.app.Parse(os.Args[1:])
 forj_app.InitializeDriversFlag()
 switch kingpin.MustParse(parse, err) {
   case "create":
     if err := forj_app.Create() ; err != nil {
         log.Fatalf("Forjj create issue. %s", err)
     }

   case "update":
        if err := forj_app.Update() ; err != nil {
            log.Fatalf("Forjj update issue. %s", err)
        }

   case "maintain":
        if err := forj_app.Maintain() ; err != nil {
            log.Fatalf("Forjj maintain issue. %s", err)
        }
   }
}
