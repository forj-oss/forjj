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
  app_infra_name_help    = "Infra repository name. By default, the name is '<Organization>-infra'. It is the Infra repo cloned directory name as well. Warning! You cannot update it on an existing workspace."
  app_orga_name_help     = "Organization name. By default, the name is given by the workspace directory name. Warning! You cannot update it on an existing workspace"
  app_creds_help         = "Credentials file. Used by plugins to collect credentials information. If you set driver credential flag on plugins, your workspace will collect them in your workspace 'forjj-creds.yml'."

  create_action_help     = "Create your Software factory.\n"

  create_orga_help       = "organization workspace used to store repositories locally or in docker volume."
  create_ssh_dir_help    = "PATH to a git ssh keys directory. It will be mounted as local path '/home/devops/.ssh' in the container."
  driver_help            = "Define the application name and type. Formated as 'APP[,APP ...]' where APP is as '<Type>:<DriverName>[:<InstanceName>]'. It will search for <Type>/<DriverName>/<DriverName>.yaml from the forjj-contribs repo."
  repos_help             = "Define the list of repository name and flow attached. Formated as 'REPO[,REPO ...]' where REPO is as '<RepoName>[:<FlowName>]'. It will search for <FlowName>/<FlowName>.yaml from the forjj-flows repo."

  docker_exe_path_help   = "Path to a static docker binary used when a forjj plugin service container requires DooD (Docker out of Docker) capability."
  contribs_repo_help     = "Set a local forjj-contribs directory like or a github like url for FORJJ plugins. You can set CONTRIBS_REPO as env."
  flows_repo_help        = "Set a local forjj-flows directory like or a github like url for FORJJ flows. You can set FLOWS_REPO as env"
  repotemplates_repo_help= "Set a local forjj-repotemplates directory like or a github like url for FORJJ Repository templates. You can set REPOTEMPLATES_REPO as env."
  create_ci_driver_help  = "Defines the Continous Integration system to implement"
  create_us_driver_help  = "Defines the GIT upstream to implement. To get the list of a driver parameter, set this flag in conjonction with --help."
  create_infra_url_help  = "Infra repository url to git clone. By default, the repository will be created."
  create_infra_path_help = "Original infra path containing source code. By default, if the repository is going to be created, it will be empty."
  create_infra_upstream  = "Set infra repository upstream instance"
  create_no_maintain_help= "Do not start maintain at the end of the create task except the infra repository upstream to create it."

  update_infra_upstream  = "Set infra repository upstream instance"
  update_branch_help     = "Set branch name where you want to create your fixes. Avoid using 'master', as 'master' usually is not opened to developers. Even if opened to you, using a branch is a good practice."
  update_action_help     = `Update the infra. Used to create/update/remove projects and infrastructure migration
(for example from local jenkins to a mesos jenkins solution)
`
  update_orga_help       = "organization workspace used to store repositories locally or in docker volume."
  update_ssh_dir_help    = "PATH to a git ssh keys directory. It will be mounted as local path '/home/devops/.ssh' in the container."
  update_ci_driver_help  = "Defines the Continous Integration system to implement"
  update_us_driver_help  = "Defines the GIT upstream to implement. To get the list of a driver parameter, set this flag in conjonction with --help."
  flow_help              = "Define the default flow to apply to new repositories."

  maintain_action_help   = "Used by your CI to update the infra from the 'infra' repository.\n"
  maintain_orga_help     = "organization workspace used to store repositories locally or in docker volume."
  maintain_infra_url_help= "Infra repository url to git clone. This option is required if the workspace is empty."
  maintain_option_file   = "Forjj yaml file for plugins options"
)
