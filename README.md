# Introduction

**This project is under development, and start with a POC to build a DevOps Solution based on github and jenkins.
So some examples/use cases described in this README.md are not fully functional.**

For details about the POC, read [this](POC.md).

# Note about Devops and Forjj

Forjj is NOT a DevOps tool, mainly because DevOps is mainly a Culture minset.
So, you can't consider that adopting Forjj will ensure you are DevOps!

But as soon as you want to adopt a DevOps Culture for any reasons, Forjj
 can help you very much to automate your process to move to a more
 compliant DevOps environment and support your new DevOps Culture.

The core design of Forjj implements an automated way to start from
scratch and maintain a combination of tools completely integrated
(github/jenkins/...) with some **DevOps paradigm** applied like:

- Everything as Code, declined by Forjj as *Infrastructure as code*.
- Control everything from GIT as Source control system.
- Use a DevOps Flow which helps to eliminates walls between teams.
  Ie, Code open, shared, and accessible to newbies or experts.

We called this **a DevOps Solution**

But do not forget that Your DevOps solution is not only what Forjj can
do for you, but how you are going to transform your process, organization and teams, with the help of Forjj

# How to start your first simple DevOps solution on your workstation?

Really? I can create a DevOps organization on my workstation??? Yes.

## Requirement:

- A linux 64 bits system with docker 1.9 or higher. It can be a VM.
- A static docker binary. You can get it from https://get.docker.com
- wget & git 1.9 or higher installed
- Internet access (through proxy or not)
- A github organization
- A github Token to manage repositories and organization

## Getting started

How easy is it to create and maintain a DevOps Solution? Follow this
3 steps to get a Minimal DevOps Solution based on github and jenkins

1. Download our sample Forjfile and edit it. Replace any <...> by appropriate value.

```bash
cd /tmp
wget -O Forjfile https://github.com/forj-oss/forjj/raw/master/samples/minimal/Forjfile
vim Forjfile
```

2. Download forjj binary:

```bash
mkdir -p ~/bin
wget -O ~/bin/forjj https://github.com/forj-oss/forjj/releases/download/latest/forjj
chmod +x ~/bin/forjj
```

3. Create your infra with:

```bash
forjj create --infra-path ~/devops/myforj-infra
```

Access your github organization to see what has been created.
Open http://localhost:8080 to see your Jenkins master.
(You may need to replace localhost by your VM DNS name instead)

## What is behind the scene?


Forjj has created 3 things:

1. Your infra repository
2. Configure a Github Entreprise organization with your inputs
3. A jenkins container on your linux system.

Nothing else?
No! There is a lot of things that forjj did:

For details about the POC, and how to build it, see [this](POC.md)
For contribution, see [this](CONTRIBUTION.md)
Well... There is a lot of things that forjj did:

### Infrastructure as Code (IaC)

Forjj has :
- generated a lot of sources which were used to instantiate
and configure, everything.
- created a new `infra` repository which contains generated source code
- Defined the Github Pull Request Flow connected with jenkins.
  (ie configured github & jenkins to build at different flow steps)

### Dev & Ops as One Team (DOaOT)

Usually, in your team, you have Dev & Ops and ... walls.

![devops](imgs/devops.png)

If you introduce Infrastructure as Code, are you sure you will contribute
to remove walls?

If you automate the infrastructure,

- How Dev could use it?
- How Dev can request new resources? like a new repository? like a new job?
- Do you need some of your Dev team to become Ops automation expert?
- How could help Dev & Ops teams to collaborate and build an efficient
DevOps model ?

In short, how do you enhance Dev & Ops to become DevOps?
There is several ideas to push here:

- Move team in a common chat room where they can ask/contribute/suggest.
- Be proud to make your infrastructure code visible and accessible.
- Give easy access to contribution at several layers.
  If your team do not need to contribute to the infrastructure by default,
  this doesn't mean some of your team member can't do it. So, let anyone
  contribute where they want to.

Where forjj helps?

- IaC : Forjj generates source code and expose it to a Git repo.
- IaC : Source code are generated from templates, stored in the Git Repo
 as well
- IaC : Each repository can be configured to follow a Change flow, like
 Github Pull Request flow.
- DOaOT : Source code are pushed to Social coding infrastructure like github.
- DOaOT : Each application define and expose useful application data to
 forjj cli. Any updates are pushed to the repository and change flow.

# Use the Minimal DevOps Solution
OK! And concretely what next could we do?

**!! Not yet implemented !!**

Well. Let's say you are a Team leader or project leader and you want to
start a new project. And so you need a Repository... but not just that!
With everything pre-configured in background, ie, jenkins and Pull Request

OK! This is done in 2 steps.

- Create a Pull request
- Merge the pull request

Now, let's create that repository called `my-project`:

```bash
forjj add repo my-project -m "My new repository" --pr-submit
```

Connect to github and click on Pull request.
A new Pull request has been created from a new fork done at your account.

So, if you are happy with that first contribution, merge it!

Then shortly, Jenkins will call forjj maintain to create your new
Repository

DONE!
Go to github and a new repository has been created.
But jenkins also has been updated with new pipeline jobs!

And so now your team can start developing!!!

In real life, you may introduce more step before creating a Pull request
, like:

- Create a git branch
- Update source with forjj (add repo, applications, update any existing apps ...)
- Update any other source, like templates
- create one or more commits
- Push
- Create your Pull request.

In the above command, forjj did it all for you.


## What is behind the scene?

Forjj has :

- created a new empty repository with a simple README and a
basic Jenkinsfile.
- Pushed the repository to github
- Configured a seedjob for your project to automatically build your
 Jenkins pipeline
- Configure github flow to call those jobs when a new Pull request is made

# What next?

As everything with forjj is managed through code, so you can do everything!!!
But everything controlled, secured and contributed.

The contribution can be made at several layers:

Newbie or Dev team:
- Call Forjj to update source code for you. Good for team that do not want
to go deeply in infrastructure automation Code.
- End user Documentation.

Expert/active contributor or Ops team
- Update Application data exposure to enhance forjj cli
- Update application templates
- Share your DevOps model (Forjfile stored as source in your infra repo)

Interesting? Then go to the [User Guide](USER_GUIDE.md) or
[Dev Guide](DEV_GUIDE.md)

Thank you
Forj team
