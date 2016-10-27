# Introduction

Forjj is the *New Power DevOps solution* to build and maintain a complete DevOps solution, integrating SCM/CI/ChatOps/CD all with
your preferred toolset.

Forjj has been designed to be widely opened to contribution.

- [Forjj plugins](https://github.com/forj-oss/forjj-contribs)

  Any pieces to build a solution with integration are stored in [forjj-contribs](https://github.com/forj-oss/forjj-contribs) repository as a forjj plugin.
  For example, ci/jenkins is a plugin which contains code to create/manage a jenkins instance and jobs. See [jenkins forjj plugin](https://github.com/forj-oss/forjj-contribs/ci/jenkinsi)
  See [contribution page to get plugins details](CONTRIBUTION.md)

- [Forjj core](https://github.com/forj-oss/forjj) (this project)

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

# Project status:

- **Partially completed** - the infra repository, `myforj-infra`, with the infra code to create/maintain it, pushed to github.
  - **DONE** github: forjj driver to help forjj to interact with github service
  - **CLOSED to DONE** jenkins: forjj driver to help forjj to create a jenkins instance as code and apply it with maintain.
- **NOT STARTED** - a `demo` repository, pushed to github.
- **IN PROGRESS** - a new jenkins instance in your workstation.
- **IN PROGRESS** - integration between github and jenkins (7 jobs)

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
go get https://github.com/forj-oss/forjj.git
cd ~/src/go/src/github.com/forj-oss/forjj
go get
go install
```

Now, `forjj` should work

For forjj github plugin:
```bash
go get https://github.com/forj-oss/forjj-contribs
cd ~/src/go/src/github.com/forj-oss/forjj-contribs/upstream/github
go generate
bin/build.sh
```

Now, a new docker image has been generated, called `docker.hos.hpecorp.net/forjj/github`


For forjj jenkins plugin:
```bash
go get https://github.com/forj-oss/forjj-contribs
cd ~/src/go/src/github.com/forj-oss/forjj-contribs/ci/jenkins
go generate
bin/build.sh
```

Now, a new docker image has been generated, called `docker.hos.hpecorp.net/forjj/jenkins`


So, now you can start playing with forjj.

## Testing forjj

We can create thing, step by step. Like create a basic workspace.

Ex: Create a basic workspace:

```bash
# Create an empty workspace.
forjj create ~/tmp/forjj
```

```bash
# Create a worspace
# create source code for github
# write github source code, add and commit it
# Instantiate github configuration as described by the generated code.
forjj create ~/tmp/forjj --debug --apps ci:jenkins,upstream:github --github-token $GITHUB_TOKEN --github-server github.com --jenkins-service-addr maison.emea.hpqcorp.net --docker-exe-path ~/tmp/docker
```

```bash
# Create a worspace
# create source code for github
# write github source code, add and commit it
# Instantiate github configuration as described by the generated code.
forjj create ~/tmp/forjj --debug --apps ci:jenkins,upstream:github --github-token $GITHUB_TOKEN --github-server github.com --jenkins-service-addr maison.emea.hpqcorp.net --docker-exe-path ~/tmp/docker
```
