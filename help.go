package main

// Define FORJJ commands/flags helps
const (
	forjj_help = `FORJJ - Software Factory 'solution as code' builder - Is the new successor of Forj.

His role is to help you building and maintaining an automated Software factory by integrating several tools together,
implemented with the 'Infrastructure as Code' idea.

For details on Forjj, visit https://github.hpe.com/forjj/forjj

'Infrastructure as Code' is part of DevOps (automation). FORJJ can help you create/update/maintain your DevOps solution.
But FORJJ is not DevOps if you do not help your teams to do DevOps (DevOps Culture).
`
	forjj_debug_help      = "debug mode activated"
	forjj_infra_name_help = "Infra repository name. By default, the name is '<Organization>-infra'. It is the Infra repo cloned directory name as well. Warning! You cannot update it on an existing workspace."
	forjj_orga_name_help  = "Organization name. By default, the name is given by the workspace directory name. Warning! You cannot update it on an existing workspace"
	forjj_creds_help      = "Credentials file. Used by plugins to collect credentials information. If you set driver credential flag on plugins, your workspace will collect them in your workspace 'forjj-creds.yml'."

	create_action_help = "Create your Software factory.\n"

	create_orga_help        = "organization workspace used to store repositories locally or in docker volume."
	create_ssh_dir_help     = "PATH to a git ssh keys directory. It will be mounted as local path '/home/devops/.ssh' in the container."
	create_no_maintain_help = "Do not instantiate at create time. (except infra upstream)"

	workspace_path_help     = "Path to your forjj workspace."
	docker_exe_path_help    = "Path to a static docker binary used when a forjj plugin service container requires DooD (Docker out of Docker) capability."
	contribs_repo_help      = "Set a local forjj-contribs directory like or a github like url for FORJJ plugins. You can set CONTRIBS_REPO as env."
	flows_repo_help         = "Set a local forjj-flows directory like or a github like url for FORJJ flows. You can set FLOWS_REPO as env"
	repotemplates_repo_help = "Set a local forjj-repotemplates directory like or a github like url for FORJJ Repository templates. You can set REPOTEMPLATES_REPO as env."

	update_action_help = `Update the infra. Used to create/update/remove projects and infrastructure migration
(for example from local jenkins to a mesos jenkins solution)
`
	update_orga_help = "organization workspace used to store repositories locally or in docker volume."
	flow_help        = "Define the default flow to apply to new repositories."

	add_action_help    = "Add a component to your Software factory."
	remove_action_help = "Remove a component from your Software factory."
	rename_action_help = "Rename a component in your Software factory."
	list_action_help   = "List components of your Software factory."

	maintain_action_help = "Used by your CI to update the infra from the 'infra' repository.\n"
	maintain_option_file = "Forjj yaml file for plugins options"

	repo_instance_name_help = "Instance of the repository."
	repo_name_help          = "Name of the repository."
	new_repo_name_help      = "New name of the repository."
	repo_flow_help          = "Flow to apply to a repository."
	repo_template_help      = "Repository template to apply."
	repo_title_help         = "Short description of the repository."
	repo_list_help          = "List of repositories separated by comma. Syntax : [instance/]repo[:flow[:repotemplate[:title]]]"

	default_flow_help = "Default flow to apply to repositories."

	app_type_help   = "Driver category."
	app_driver_help = "Driver name."
	app_name_help   = "Application instance name. If not set, forjj will use the driver name."
	app_list_help   = "List of application separated by comma. Syntax : category:driver[:instance]"
)
