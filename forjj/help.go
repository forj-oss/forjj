package main

// Define FORJJ commands/flags helps
const (
  app_help =
`FORJJ - Software Factory 'solution as code' builder - Is the new successor of Forj. 

His role is to help you building an automated Software factory by integrating several tools together,
implemented with the 'Infrastructure as Code' idea.

'Infrastructure as Code' is part of DevOps. FORJJ can help you create/update/maintain your DevOps solution. 
But FORJJ is not DevOps if you do not help your teams to do DevOps.
`
  app_debug_help         = "debug mode activated"
  app_infra_name_help    = "Infra repository name. By default, the name is '<Organization>-infra'. It is the Infra repo cloned directory name as well."
  app_orga_name_help     = "Organization name. By default, the name is given by the workspace directory name"

  create_action_help     = "Create your Software factory.\n"

  create_orga_help       = "organization workspace used to store repositories locally or in docker volume."
  create_ssh_dir_help    = "PATH to a git ssh keys directory. It will be mounted as local path '/home/devops/.ssh' in the container."
  create_ci_driver_help  = "Defines the Continous Integration system to implement"
  create_us_driver_help  = "Defines the GIT upstream to implement. To get the list of a driver parameter, set this flag in conjonction with --help."
  create_infra_url_help  = "Infra repository url to git clone. By default, the repository will be created."
  create_infra_path_help = "Original infra path containing source code. By default, if the repository is going to be created, it will be empty."
  create_contrib_help    = "Set a local forjj-contribs directory like or a github like url."

  update_action_help     = `Update the infra. Used to create/update/remove projects and infrastructure migration 
(for example from local jenkins to a mesos jenkins solution)
`
  update_orga_help       = "organization workspace used to store repositories locally or in docker volume."
  update_ssh_dir_help    = "PATH to a git ssh keys directory. It will be mounted as local path '/home/devops/.ssh' in the container."
  update_ci_driver_help  = "Defines the Continous Integration system to implement"
  update_us_driver_help  = "Defines the GIT upstream to implement. To get the list of a driver parameter, set this flag in conjonction with --help."
  update_contrib_help    = create_contrib_help

  maintain_action_help   = "Used by your CI to update the infra from the 'infra' repository.\n"
  maintain_orga_help     = "organization workspace used to store repositories locally or in docker volume."
  maintain_infra_url_help= "Infra repository url to git clone. This option is required if the workspace is empty."
)
