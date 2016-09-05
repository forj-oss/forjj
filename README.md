# Introduction

**The project is under development, and start with a POC to build github/jenkins integration.**
[github + jenkins Solution Builder][See Example below]

Forjj is the *New Power DevOps solution* to build and maintain a complete DevOps solution, integrating SCM/CI/ChatOps/CD all with
your preferred toolset.

Forjj has been designed to be widely opened to contribution.

- Forjj plugins

  Any pieces to build a solution with integration are stored in [forjj-contribs](https://github.hpe.com/christophe-larsonneur/forjj-contribs) repository as a forjj plugin.
  For example, ci/jenkins is a plugin which contains code to create/manage a jenkins instance and jobs. See [jenkins forjj plugin](https://github.hpe.com/christophe-larsonneur/forjj-contribs/ci/jenkinsi)
  See [contribution page to get plugins details](CONTRIBUTION.md)

- Forjj core

  Forjj core is a GO binary. It contains code to:

  - automate the DevOs solution creation from a solution description data.
  - store source code in git repositories.
  - create a demo project
  - enhance solution, in scaling, or any other infrastructure update.
  - repositories management

  Forjj core could be expanded with a web ui to:

  - expose options of plugins
  - assist in project management (creation/update/delete)
  - assist in solution infrastructure management (scale/changes)


# How to start your first simple DevOps organization on your workstation?

Really? I can create a DevOps organization on my workstation??? Yes.

Here is the requirement:

- docker : You need a linux docker system, 1.9.1 or higher.
- github : To simplify, we use github.com for the SCM upstream part. You must have a valid account on github.com and a valid personal token. (`GITHUB_TOKEN`)
  The token must have access to `admin:org, admin:repo_hook, admin:org_hook`
- Linux OS: A jenkins instance will be created as docker container. You workstation should have at least 4G RAM to support properly the jenkins container (Use 1Gb)
- firewall/Nat: Your workstation needs to be opened and accessible from github to port 8080 (jenkins port)
- curl/wget to get forjj binary.

As I'm at the beginning of this POC, I'm generating the forjj binary everytime I'm building it. Then it gets published to my workstation at `http://lacws.emea.hpqcorp.net/Install_private_pdeit/larsonsh/forjj`. But be aware: this version is a DEV version and is totally unstable.

If you want a more stable version, you should build it yourself, as I'm pushing my code frequently when I consider it to be acceptable version. See how to build it section.

NOTE: I'm using forjj to build forjj on github.hpe.com. So, the organization will get the infrastructure source and description thanks to forjj.
At this stage, a more official stable forjj binary will get published.

# Run it

NOTE: Following steps are not ready to be used as is. So, you certainly needs to build several things around to make it work. So, steps listed below are given to show how easy it will be and what will be done.


Create your infra with:

    $ mkdir ~/bin && wget -O ~/bin/forjj http://lacws.emea.hpqcorp.net/Install_private_pdeit/larsonsh/forjj/forjj && chmod +x ~/bin/forjj
    $ forjj create ~/devops/myforj --apps ci:jenkins,ustream:github --github-server github.hpe.com --github-token d6603f803805a4e9f51ebf85d5668bf499f58331 --jenkins-addr myworkstation.emea.hpqcorp.net

Then forjj will create in ~/devops/myforj :
- **IN PROGRESS** - the infra repository, `myforj-infra`, with the infra code to create/maintain it, pushed to github.
  - github: forjj driver to help forjj to interact with github service
  - jenkins: forjj driver to help forjj to create a jenkins instance as code and apply it with maintain.
- **NOT STARTED** - a `demo` repository, pushed to github.
- **IN PROGRESS** - a new jenkins instance in your workstation.
- **IN PROGRESS** - integration between github and jenkins (7 jobs)

If you want to create a new repository in 'myforj' organization, execute the following:

    $ forjj update ~/devops/myforj --repo "myrepo:My new software Repository"

This will create in $PWD/devops/myforj infra a new commit in 'create-repo-myrepo' branch with:
- "FORJJ: Create a new repository called 'myrepo'"
  We should see a new yaml file created in myrepo-infra/repos/myrepo/myrepo.yaml

You will see a new branch 'create-repo-myrepo' in github. So, you need to create your first pull request to master.
The jenkins `myforj_infra_PR` should be started pretty soon and shown GREEN status in github.
The PR job could be there to validate that forjj could maintain this repository with this code. (TBD)
Then merging this PR will start the `myforj_infra_master` job. This job will call forjj maintain to create the repo with github/jenkins integration (3 new jobs)

# How to build it? How to test it?

Today, there is 3 components to build:
- forjj (this project)
- forjj github (plugin)
- forjj jenkins (plugin)

## Build forjj

As soon as you have defined the GOPATH variable, you could build forjj as follow:


```bash
mkdir ~/src/go
export GOPATH=~/src/go
export PATH=$PATH:$GOPATH/bin
go get https://github.hpe.com/christophe-larsonneur/forjj.git
cd ~/src/go/src/github.hpe.com/christophe-larsonneur/forjj
go get
go install
```

Now, `forjj` should work

For forjj github plugin:
```bash
go get https://github.hpe.com/christophe-larsonneur/forjj-contribs
cd ~/src/go/src/github.hpe.com/christophe-larsonneur/forjj-contribs/upstream/github
go generate
bin/build.sh
```

Now, a new docker image has been generated, called `docker.hos.hpecorp.net/forjj/github`


For forjj jenkins plugin:
```bash
go get https://github.hpe.com/christophe-larsonneur/forjj-contribs
cd ~/src/go/src/github.hpe.com/christophe-larsonneur/forjj-contribs/ci/jenkins
go generate
bin/build.sh
```

Now, a new docker image has been generated, called `docker.hos.hpecorp.net/forjj/jenkins`


So, now you can start playing with forjj.

## Testing forjj

To create a basic workspace:

```bash
# Create an empty workspace.
forjj create ~/tmp/forjj
```

```bash
# Create a worspace
# create source code for github
# write github source code, add and commit it
# Instantiate github configuration as described by the generated code.
forjj create ~/tmp/forjj --debug --apps ci:jenkins,upstream:github --github-token $GITHUB_TOKEN --github-server github.hpe.com --jenkins-service-addr maison.emea.hpqcorp.net --docker-exe-path ~/tmp/docker
```

```bash
# Create a worspace
# create source code for github
# write github source code, add and commit it
# Instantiate github configuration as described by the generated code.
forjj create ~/tmp/forjj --debug --apps ci:jenkins,upstream:github --github-token $GITHUB_TOKEN --github-server github.hpe.com --jenkins-service-addr maison.emea.hpqcorp.net --docker-exe-path ~/tmp/docker
```

Thank you
