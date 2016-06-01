# Introduction

**The project is under development, and start with a POC to build github/jenkins integration.**
[github + jenkins Solution Builder][See Example below]

Forjj is the *New Power DevOps solution* to build and maintain a complete DevOps solution, integrating SCM/CI/ChatOps/CD all with 
your preferred toolset.

Forjj has been designed to be widely opened to contribution. 

- Solution Builder

  Any pieces to build a solution with integration are stored in [forjj-contribs](forjj-contribs repository)
  For example, jenkins-ci is a solution builder which contains integration builder with github. See [forjj-contribs](jenkins-ci solution builder)
  See [CONTRIBUTION.md](contribution page to get solution builder details)

- Forjj core

  Forjj core scripts are located in this repository. It contains code to:
  
  - automate the solution Builder creation from a solution description data.
  - store source code in git repositories.
  - create a demo project
  - enhance solution, in scaling, or any other infrastructure update.

  Forjj core should be expanded to expose a nodejs web ui to:

  - expose options of solution builders
  - assist in project management (creation/update/delete)
  - assist in solution infrastructure management (scale/changes)


# How to start building your first simple DevOps organization on your workstation?

Really? I can create a DevOps organization on my workstation??? Yes. 

Here is the requirement:

- docker : You need a linux docker system, 1.9.1 or higher.
- github : To simplify, we use github.com for the SCM upstream part. You must have a valid account on github.com and a valid personal token. (`GITHUB_TOKEN`)
  The token must have access to `admin:org, admin:repo_hook, admin:org_hook`
- Linux OS: A jenkins instance will be created as docker container. You workstation should have at least 4G RAM to support properly the jenkins container (Use 1Gb)
- firewall/Nat: Your workstation needs to be opened and accessible from github to port 8080 (jenkins port)
- curl : Used to get a simple script to prepare your forjj image with your local docker version.

# Run it

    $ mkdir ~/myforj
    $ DOCKER_SOCK="-v /var/run/docker.sock:/var/run/docker.sock"
    $ GITHUB_TOKEN=d6603f803805a4e9f51ebf85d5668bf499f58331
    $ docker run -it --rm $DOCKER_SOCK docker.hos.hpecorp.net/devops/forjj create ~/myforj --ci jenkins-ci --scm-us github --github-token $GITHUB_TOKEN

Then forjj will create in $PWD/myforj :
- an `infra` repository, updated with the infra code to create/maintain it, pushed to github.
- a `demo` repository, pushed to github.
- a new jenkins instance in your workstation.
- integration between github and jenkins (7 jobs)

If you want to create a new repository, start the following:

    $ docker run -it --rm $DOCKER_SOCK docker.hos.hpecorp.net/devops/forjj update ~/myforj --repos myrepo --branch new-repo

This will create in $PWD/myforj infra a new commit with:
- New repo definition.

You will see a new branch 'new-repo' in github. So, you need to create your first pull request to master. 
The jenkins `myforj_infra_PR` should be started pretty soon and shown GREEN status in github.
Then merging this PR will start the `myforj_infra_master` job. This job will call forjj-orchestrator to create the repo with github/jenkins integration (3 new jobs)
