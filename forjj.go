package main

import (
        "os"
        "gopkg.in/alecthomas/kingpin.v2"
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
        forj_app.Create()
        //forj_app.Maintain()
        //fmt.Printf("Contrib-repo: %#v\n", *cr_contrib_repo)


   case "update":
        //forj_app := App{}
        //forj_app.Load()
        forj_app.Update()
        println("update")
        //fmt.Printf("Contrib-repo: %#v\n", *up_contrib_repo)

   case "maintain":
        //forj_app := App{}
        //forj_app.Load()
        forj_app.Maintain()
        println("maintain")
   }
}
