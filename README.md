# Forjj - Factory as Code

[![Build Status](https://jenkins-forjj.eastus.cloudapp.azure.com/buildStatus/icon?job=forjj/master)](https://jenkins-forjj.eastus.cloudapp.azure.com/job/forjj/master)

## Introduction

What is `Forjj`?
This is a tool to help building a Factory as Code.

What is Factory as Code?
In short, it is a way to deploy and maintain a collection of tools and process, Described/Developed from code perspective.

In a little more details, `Forjj` describes in a `Forjfile` (yaml file) a collection of tools, like github, jenkins, vault, etc... and
attach one of more flow process to it to deliver an immediate service to Development and Operators teams.

This means that as soon as you started your Factory with `Forjj`, your next action is ... ask your development teams to start their 
development work with a minimum viable software factory to run CI/CD, tests, ...

From the operator point of view, with `Forjj`, you deliver services through code as well. So that you do not need to send
a mail to your preferred Operator to create a repo and configure everything around related to the flow, you just
explain how to update some configuration files, request a pull request to approve, suggest, discuss and then merge.

Then the `Forjj pipeline` (created by `Forjj`) will deliver the change to your factory (ie new repo (github), webhook (github), pipeline (jenkins), secrets path (vault), etc ...)

As everything is Code, the core of your operator team is to deliver the code in any kind of language or tools to deliver the final service to the end users.

## Note about Devops and `Forjj`

It is recurrent that `DevOps` coaches remind that `DevOps` is not methodology or a tools or a team.

So, `Forjj` is NOT a DevOps tool, mainly because `DevOps` is mainly a Culture mindset.
So, you can't consider that adopting `Forjj` will ensure you are DevOps!

But as soon as you want to adopt a `DevOps` Culture for any reasons, `Forjj`
 can help you very much to automate your process to move to a more
 compliant DevOps environment and support your new `DevOps` Culture.

The core design of Forjj implements an automated way to start from
scratch and maintain a combination of tools completely integrated
(github/jenkins/...) with some **DevOps paradigm** applied like:

- Everything as Code, declined by Forjj as *Factory as code*.
- Control everything from GIT as Source control system.
- Use a DevOps Flow which helps to eliminates walls between teams.
  Ie, Code open, shared, and accessible to newbies or experts.

We called this **a DevOps Solution**

But do not forget that Your DevOps solution is not only what Forjj can
do for you, but how you are going to transform your process, organization and teams, with the help of Forjj

Note also, that Forjj requires you to help others.
What does it mean? Shortly, we need your contribution!

In details, Forjj is designed to share the way we do to start a new factory.

If you use Forjj to deliver your Factory, share this if you can, so that others can rely on your work and our work.
Which finally will build the Forjj community to help anyone and improve Forjj and your factory.

## How to start your first simple DevOps solution on your workstation

Really? I can create a DevOps organization on my workstation??? Yes.

### Requirement

- A linux 64 bits system with docker 1.9 or higher. It can be a VM.
- A *static* docker binary. You can get it from https://get.docker.com
- wget & git 1.9 or higher installed
- Internet access (through proxy or not)
- A github organization
- A github Token to manage repositories and organization

### Getting started

How easy is it to create and maintain a DevOps Solution? Follow this
3 steps to get a Minimal DevOps Solution based on github and jenkins

1. Download our Forjfile model template and edit it. Replace any <...> by appropriate value.

    ```bash
    cd /tmp
    wget -O Forjfile https://github.com/forj-oss/forjj/raw/master/samples/minimal/Forjfile
    vim Forjfile
    ```

> **NOTE**: A Forjfile model is a unique Forjfile describing a Factory model and used at create time only to create a new factory.
> This file usually defines deployments specification (`deployments/<deployName>/define/...`) and can contain credentials.
> When your forge will be created:
> - deployment specfications will be dispatched in several deployment files in the infra repository (`deployments/<deployName>/Forfile`)
> - Credentials data will be moved to Forjj vault.
>
> Later, you can create your own Forjfile model to help other teams to create their Factory from your Factory model. So it became easy to deploy a `Factory On Demand`.

1. Download `forjj` binary with the `do-refresh-forjj.sh` helper

    ```bash
    mkdir -p ~/bin
    wget -O ~/bin/do-refresh-forjj.sh https://raw.githubusercontent.com/forj-oss/forjj/master/bin/do-refresh-forjj.sh
    bash ~/bin/do-refresh-forjj.sh
    ```

    **NOTE**: Anytime you need to get latest forjj version, call `do-refresh-forjj.ss`.

2. Create your infra with:

    ```bash
    forjj create --infra-path ~/devops/myforj-infra
    ```

    Access your github organization to see what has been created.
    Open http://localhost:8080 to see your Jenkins master.
    (You may need to replace localhost by your VM DNS name instead)

### What is behind the scene?

`Forjj` has created 4 things:

1. Your source infra repository
2. A production deployment repository
3. Configured a Github public/entreprise organization with your inputs (repos/teams/webhooks)
4. Started a jenkins container on your linux system.

This was done by 2 phases described in details below. This is given to understand how it works...

#### Create phase

This is the details of what `forjj` did at `create phase`:

- `forjj` creates local git repositories, infra and production to host respectively source code (source repository) and deployment code (deployment repository).
- It starts creating source code for github as follow
  - It starts a `forjj plugin` called `forjj-github` as container
  - Then creates 1 file. `github.yaml` under `apps/upstream/github` directory.
- It starts creating source code for jenkins as follow
  - It starts a `forjj plugin` called `forjj-jenkins` as container
  - Then creates few files. like `jenkins.yaml` under `apps/ci/jenkins` directory. forjj-jenkins use GO templates to generate `deploy code`.
    By default, this templates files are stored in forjj-jenkins plugin under [`templates`](https://github.com/forj-oss/forjj-jenkins/tree/master/templates)

- Now `forjj` starts creating your deploy code.
- It starts generating deploy code for github as follow
  - It starts/restarts the `forjj plugin` called `forjj-github` as container
  - Then copy 1 file. `github.yaml` under `github` directory for the `deployment repository`.
- It starts creating source code for jenkins as follow
  - It starts/restarts the `forjj plugin` called `forjj-jenkins` as container
  - It reads the `templates.yaml` to identify how to copy/generate source code. Except if you set `appplications/jenkins/source-templates` in your Forjfile, the templates.yaml file is located under [`templates`](https://github.com/forj-oss/forjj-jenkins/tree/master/templates)
  - It copies or generates code to the deployment repository.
  - If the templates describes a `run_build` command, this command is executed.
  - Any files generated at build time are collected and reported back to forjj.

- Now `forjj` do a git add of those new files. Ready to to be `git commited`.

Except if you used `--no-maintain`, `forjj create` automatically run your initial deployment code to deploy your factory. This is done through the call to `forjj maintain` internally.

NOTE: `forjj` communicates with `forjj plugins` using REST API.

#### Maintain phase

Now, `forjj` is going to deploy, based on the code generated, ie from `deployment repository`. This is an automatic maintain step executed from `forjj create`.

- `forjj` started the github deployment.
  - forjj starts (or restart) at this time the `forjj plugin` called `forjj-github` as container
  - forjj-github created the repos and configure them accordingly to what Forjfile describes.
- `forjj` push source codes created at create phase
- It started the jenkins deployment
  - `forjj` started (or restarts) the `forjj plugin` called `forjj-jenkins` as container.
  - `forjj-jenkins` read the `maintain-cmd.yaml` from the deployment repository and run the command identified as `deploy_run`
  - `forjj-jenkins` then wait for the end of this deployment command.

## What do we do next

In short, you start using your factory!

**Warning! What is described here is currently not fully implemented !!**

Well. Let's say you are a Team leader or project leader and you want to
start a new project. And so you need a Repository... but not just that!
With everything pre-configured in background, ie, jenkins and Pull Request

OK! This is done in 3 steps.

- Fork the infra repository
- Create a Pull request and update the `deployments/production/Forjfile`
- Merge the pull request

Now, let's dot it. We want to create a repository called `my-project`:

So, you read in the documentation that you have to update the production Forjfile with

```bash
repositories:
  my-project:
    title: This is my new repo to host my project
    issue_tracker: false # By default it is true... your choice for the example...
```

- Connect to github, forjj the INFRA repo and update the `deployments/production/Forjfile`.
- Update the appropriate section and submit as pull request.

So, if you are happy with that first contribution, merge it!

Then shortly, Jenkins will call `forjj maintain` to create your new repository.

> **NOTE:**  This Jenkins job is not yet implemented. It will be shortly with a Jenkinsfile in your infra repository.
> This task is currently made manually in 2 steps:
> `forjj update production && forjj maintain production`

DONE!

Go to github and a new repository has been created.
But jenkins also has been updated with new pipeline jobs!

And so now your team can start developing!!!

In real life, you may introduce more step before creating a Pull request
, like:

- Create a git branch
- Update source with `Forjj` (add repo, applications, update any existing apps ...)
- Update any other source, like templates
- create one or more commits
- Push
- Create your Pull request.

In the above command, `Forjj` did it all for you.


### What is behind the scene?

`Forjj` has :

- created a new empty repository with a simple README and a
basic Jenkinsfile.
- Pushed the repository to github
- Configured a seedjob for your project to automatically build your
 Jenkins pipeline
- Configure github flow to call those jobs when a new Pull request is made

## What next?

As everything with `Forjj` is managed through code, so you can do everything!!!
But everything controlled, secured and contributed.

The contribution can be made at several levels:

Newbie or Dev team:

- Call `forjj` to update source code for you. Good for team that do not want
to go deeply in infrastructure automation Code.
- End user Documentation.

Expert/active contributor or Ops team

- Update Application data exposure to enhance `Forjj` cli
- Update application templates
- Share your DevOps model (Forjfile stored as source in your infra repo)

Interesting? Then go to the [User Guide](USER_GUIDE.md) or
[Dev Guide](DEV_GUIDE.md)

Thank you
Forj team
