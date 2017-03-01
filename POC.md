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

**Requirement:**
- docker 1.10 or higher
- git
- bash

**Optional:**
- go 1.5 or higher - by default, we use docker to build with GO 1.7
- glide 0.12.1 or higher - by default, we use docker to install deps

1. Create (or use) your GO Path environment

```bash
mkdir ~/src/go/src # if you want to create it
cd ~/src/go/src # Or use $GOPATH/src if you already have one.
```

2. Clone your fork (or https://github.com/forj-oss/forjj)

```bash
git clone https://github.com/<MyName>/forjj
cd forjj
```

3. If you want to use docker to build, load the build environment.

```bash
echo ~/src/go > .gopath
source build-env.sh
create-build-env.sh
```

**Note:**
This build environment creates aliases to do `go` and `glide` transparently
from a docker container, exactly like `go` and `glide` would be installed on
your workstation.

So, you will be able to call directly `go` or `glide` from command line.

`build-env.sh` script add 2 aliases to your ~/.bashrc called `build-env` and `build-env-unset`

Use `build-env-unset` to unload the build environment

Use `build-env` to load a project build environment like `forjj`. Note that
this script will simply execute `source build-env.sh` for you.
As soon as you have defined the GOPATH variable, you could build forjj as follow:

You can also use a dynamic prompt to keep you informed about which build
environment is loaded. See #Build_environment.

`create-build-env.sh` (located in bin/) will create a docker image with
 golang 1.7 and glide. See [build env Dockerfile](glide/Dockerfile)

4. Install project dependencies with glide

```bash
glide i
```

5. Install forjj

```bash
go install
```

Now, `forjj` should work:

```bash
$GOPATH/bin/forjj
```


## Build plugins

1. Cloning forjj-contribs

```bash
cd ~/src/go/src
git clone https://github.com/forj-oss/forjj-contribs
```

2. Build github plugin

```bash
cd ~/src/go/src/forjj-contribs/upstream/github
go generate
bin/build.sh
```

Now, a new docker image has been generated, called `docker.io/forjj/github`

3. BUild jenkins plugin

```bash
cd ~/src/go/src/github.com/forj-oss/forjj-contribs/ci/jenkins
go generate
bin/build.sh
```

Now, a new docker image has been generated, called `docker.io/forjj/jenkins`


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
