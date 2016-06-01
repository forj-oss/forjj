# How to contribute to this project?

Fork it and submit PR!

# How the project is organized:

Currently forjj is a schell script code. But I believe we should move to GO instead.

forjj is divided in 2 pieces:

- UI: Not yet developped. It defines a Web UI to help build/maintain the infra repo.
  It should be used to 
  - add/remove repos
  - add/update infrastructures services. Like moving jenkins-ci from a local workstation container to UCP or mesos.
- CLI: basic initial core code. It defines instructions to create the infra repo thanks to options
  In the core, the cli do nothing alone. It requires a SolutionBuilder code to do the real work.
  This code is located in [https://github.hpe.com/christophe-larsonneur/forjj-contribs](forjj-contribs repository)

# FORJJ CLI

forjj is a container built from `ci/Dockerfile`

forjj 

It defines following main keywords:

- `create <ORGA> [options]` : Create/update a collection of repositories stored under the `<ORGA>` directory. This directory can be stored in a docker volume or your local filesystem.
- `update <ORGA> [options]` : update the a collection of repositories stored under the `<ORGA>` directory and push.
- `maintain <ORGA>`         : Read the infra repo and update the infrastructure if needed. Usually this call is made from jenkins or if infra is a new repository.


## How create do his work?

* `forjj create` works on the `infra` repository.
  * local: create or clone `infra` repository.
  * local: create/update repos data in `infra` repo
  * contribs: check tools context (scm/ci/...) It calls forjj-contribs SolutionBuilder `checks` and `scm-configure` to do this task.
    * `forjj-contribs` check if services already exist or not.
    * `scm-configure` install, configure and run the SCM service. If no install is required, it will just configure it. (online service)
  * local: if infra repo is new, push it to the SCM under master branch. the remote will be identified in git as 'upstream'. Then `forjj maintain <ORGA>`
    will be executed
  * local: if infra repo is not new, push the code update in a branch (--branch)
    At some point a test and merge is done from the branch to master (like PR in github) and a CI job is started to run the `forjj maintain <ORGA>`

## How update do his work?

Shortly, the update will do the update in the `infra` repo and push it to `origin/<branch>`.

## How maintain do his work?
* Start working on `infra` repository.
  * local: New repo are created. If local directory exists with README.md and CONTRIBUTION.md, they will be added as first initial commit and pushed.
  * contribs: Compare current infra with the infra repo
    * contribs: install, configure and run any new services. Old services should be simply stopped.
    * contribs: Existing services are updated according to what the infra has defined. (mainly projects management)


