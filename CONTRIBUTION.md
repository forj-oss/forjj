# How to contribute to this project?

Fork it and submit PR!

# How the project is organized:

Currently forjj is a schell script code. But I believe we should move to GO instead.

forjj is divided in 2 pieces:

- forjj cli: forjj core code.
  Shortly, forjj manage an infra GIT repository and communicate with several plugins to do the real task.
  This code is located in this repository
- forjj plugins: Application drivers for forjj.
  This code is located in [https://github.hpe.com/christophe-larsonneur/forjj-contribs](forjj-contribs repository)

  Shortly, a plugin is a container started by forjj as a service. The plugin implements a REST API used by Forjj to communicate with.
  The plugin itself implements all the logic to control/configure and install the application that forjj will use.

# FORJJ CLI

forjj is a simple binary that you can download from http://lacws.emea.hpqcorp.net/Install_private_pdeit/larsonsh/forjj

It defines following main keywords:

- `create`   : Create a DevOps Solution with collection of application and repositories under a defined flow. In a single command line, we can start a complete ready to work DevOps solution.
- `update`   : Update do a change proposal to a running DevOps Solution, by updating it in the infra repository source code.
- `maintain` : Read the infra repo and update the infrastructure if needed. Usually this call is made from jenkins or if infra is a new repository.

Those keywords are implemented as well to each plugins in his REST API.
But usually, when forjj implements a flow on a repo, it will contact several plugins to implement the DevOps integration. And forjj will use different plugins REST API functions to complete it. I will provide some few example to illustrate this.


To give an example, let's see
## How create do his work?

`forjj create` shortly do the following:
* create a local `infra` repository.
* Call the upstream to ensure the remote to exist.
* Configure the local upstream with the remote link.
* Generate source code for each application requested and store them in the infra repo.
* Generate source code for each Repository requested and store them in the infra repo. Each repo defines a Flow to apply.
* start maintain to instantiate the generated source, ie application, then repos


## How update do his work?

Shortly, the update follow the infra-repo flow to implement any kind of change. Globally it does:
* create a git branch to store the change
* re-generate any source code to reflect the new change requested.
* push to `git origin`

From this point, The team can change anything to the source code and submit it to the team.
As soon as the change gets accepted/integrated by the team, forjj could be used to clean up the branch.

## How maintain do his work?
Start working on `infra` repository.
* local: New repo are created. The can be created with several predefined files as templates.
* plugins will compare current infra with the infra repo
* plugins will install, configure and run any new services. Old services should be simply stopped.
* Existing services are updated by same plugins according to what the infra has defined. (mainly projects management)

