package main

import (
        "fmt"
        "os"
        "gopkg.in/alecthomas/kingpin.v2"
//        "github.com/alecthomas/kingpin"
)

// Define generic FORJJ commands/flags

var (
     app      = kingpin.New(os.Args[0], app_help).UsageTemplate(DefaultUsageTemplate)
     debug    = app.Flag("debug", "Enable debug mode.").Bool()

// ****** CREATE Action ********** //
     create = app.Command("create", create_action_help)

     cr_orga  = create.Arg( "organization", create_orga_help   ).Required().String()
     cr_ssh_dir  = create.Flag("ssh_dir",      create_ssh_dir_help).  Default(fmt.Sprintf("%s/.ssh", os.Getenv("HOME"))).String()
     cr_contrib_repo_flag = create.Flag("contrib-repo", create_contrib_help)
     cr_contrib_repo = cr_contrib_repo_flag.String()

     // CI Drivers type parameters
     cr_ci_driver_flag = create.Flag("ci", create_ci_driver_help).Envar("FORJJ_CI")
     cr_ci_driver      = cr_ci_driver_flag.String()
     // Additional options will be loaded from the selected driver itself.

     // GIT Upstream type Drivers parameters
     cr_gitus_driver_flag = create.Flag("git-us", create_us_driver).Envar("FORJJ_SCM_UPSTREAM")
     cr_gitus_driver      = cr_gitus_driver_flag.String()
     // Additional options will be loaded from the selected driver itself.

// ****** UPDATE Action ************ //
     update   = app.Command("update", update_action_help)

     up_orga  = update.Arg("organization", update_orga_help).Required().String()
     up_ssh_dir  = update.Flag("ssh_dir",      update_ssh_dir_help).  Default(fmt.Sprintf("%s/.ssh", os.Getenv("HOME"))).String()
     up_contrib_repo_flag = update.Flag("contrib-repo", update_contrib_help)
     up_contrib_repo = up_contrib_repo_flag.String()

     // CI Drivers type parameters
     up_ci_driver_flag = update.Flag("ci", update_ci_driver_help).Envar("FORJJ_CI")
     up_ci_driver      = up_ci_driver_flag.String()
     // Additional options will be loaded from the selected driver itself.

     // GIT Upstream type Drivers parameters
     up_gitus_driver_flag = update.Flag("git-us", update_us_driver).Envar("FORJJ_SCM_UPSTREAM")
     up_gitus_driver      = up_gitus_driver_flag.String()
     // Additional options will be loaded from the selected driver itself.

// ****** MAINTAIN Action ************ //
     maintain = app.Command("maintain", maintain_action_help)
     ma_orga  = maintain.Arg("organization", maintain_orga_help).Required().String()
)

// Define FORJJ commands/flags helps
const (
  app_help =
`FORJJ - Software Factory 'solution as code' builder - Is the new successor of Forj. 

His role is to help you building an automated Software factory by integrating several tools together,
implemented with the 'Infrastructure as Code' idea.

'Infrastructure as Code' is part of DevOps. FORJJ can help you create/update/maintain your DevOps solution. 
But FORJJ is not DevOps if you do not help your teams to do DevOps.
`
  create_action_help     = "Create your Software factory.\n"

  create_orga_help       = "organization workspace used to store repositories locally or in docker volume."
  create_ssh_dir_help    = "PATH to a git ssh keys directory. It will be mounted as local path '/home/devops/.ssh' in the container."
  create_contrib_help    = "Mount a local forjj-contribs directory like"
  create_ci_driver_help  = "Defines the Continous Integration system to implement"
  create_us_driver       = "Defines the GIT upstream to implement. To get the list of a driver parameter, set this flag in conjonction with --help."

  update_action_help     = `Update the infra. Used to create/update/remove projects and infrastructure migration 
(for example from local jenkins to a mesos jenkins solution)
`
  update_orga_help       = "organization workspace used to store repositories locally or in docker volume."
  update_ssh_dir_help    = "PATH to a git ssh keys directory. It will be mounted as local path '/home/devops/.ssh' in the container."
  update_contrib_help    = "Mount a local forjj-contribs directory like"
  update_ci_driver_help  = "Defines the Continous Integration system to implement"
  update_us_driver       = "Defines the GIT upstream to implement. To get the list of a driver parameter, set this flag in conjonction with --help."

  maintain_action_help   = "Used by your CI to update the infra from the 'infra' repository.\n"
  maintain_orga_help     = "organization workspace used to store repositories locally or in docker volume."
)

