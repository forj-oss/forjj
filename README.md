# Introduction

**The project is under development, and start with a POC to build github/jenkins integration.**

For details about the POC, read [this](POC.md).

If you are interesting on what forjj will do in future. Have a look [here](FUTURE.md)

# How to start your first simple DevOps solution on your workstation?

Really? I can create a DevOps organization on my workstation??? Yes.

## Requirement:

- docker : You need a linux docker system, 1.9.1 or higher.
- github : To simplify, we use github.com for the SCM upstream part. You must have a valid account on github.com and a valid personal token. (`GITHUB_TOKEN`)

  The token must have access to `admin:org, admin:repo_hook, admin:org_hook`
  **NOTE**: The latest version of github entreprise do not authorize to create an organization from the API if you are not `siteadmin`. But You can usually create an organization from the web interface. Forjj github will take care of that.
- Linux OS: A jenkins instance will be created as docker container. You workstation should have at least 4G RAM to support properly the jenkins container (Use 1Gb)
- firewall/Nat: Your workstation needs to be opened and accessible from github to port 8080 (jenkins port)
- curl/wget to get forjj binary.

If you do not want to build your own forjj version from source, you can get the version I'm publishing daily. It gets published to my workstation at `http://lacws.emea.hpqcorp.net/Install_private_pdeit/larsonsh/forjj`. But be aware: this version is a DEV version and is totally unstable.

If you want a more stable version (source not changing too much...), you should build it yourself, as I'm pushing my code frequently when I consider it to be acceptable version. See [how to build it section](forjj#how-to-build-it-how-to-test-it).

NOTE: I'm using forjj to build forjj on github.com.

# Run it

NOTE: Following steps are not ready to be used as is. So, you certainly needs to build several things around to make it work. So, steps listed below are given to show how easy it will be and what will be done.

1. Download forjj binary:

```bash
mkdir -p ~/bin && wget -O ~/bin/forjj http://lacws.emea.hpqcorp.net/Install_private_pdeit/larsonsh/forjj/forjj && chmod +x ~/bin/forjj
```

2. Create your infra with:

```bash
GITHUB_TOKEN=d6603f803805a4e9f51ebf85d5668bf499f58331
forjj create ~/devops/myforj --apps ci:jenkins,upstream:github
                             --github-token $GITHUB_TOKEN
                             --jenkins-addr myworkstation.emea.hpqcorp.net
```

Then forjj will create in ~/devops/myforj :
- The infra repository, `myforj-infra`, with the infra code to create/maintain it, pushed to github.
- Source code to implement a jenkins instance to your workstation with docker
- Source code to configure your github organization.

So, now, you should have your organization created with one repo called `myforj-infra` and one jenkins server started from docker.
At this point you will start to use a source update flow, like 'Pull Request flow'. As at the current time the flow is not fully functionnal, the flow is quite basic.

When you want to update something, you do it, yourself on the myforj-infra code or with forjj update.
Today forjj update has capability to add a new repo. That's what I proposed you next.
Forjj will update source code and create commit. But no push will be done by forj. This will be the end user resposability to verify and then push.

The next step will really instanciate your change, by creating the required repo. Explained at step 4. Let's see step 3 first.

3. Add new repo to github:

Creating a new repository in 'myforj' organization is simple: Execute the following:

```bash
forjj update ~/devops/myforj --add-repos "myrepo:::My new software Repository"
```

Normally, you will be able to apply a flow and a repotemplate to the new repo, it gets pre-populated and ready to used with the flow and connected between jenkins and github to support this flow.
Those information could be explicitely set with the `--add-repos` `<myrepo>:<Flow>:<Repotemplate>:<Repo title>`


So, `forjj update` has created in $PWD/devops/myforj infra a new commit in 'master' branch with:
- "FORJJ: Create a new repository called 'myrepo'"
  We should see a new yaml file created in myrepo-infra/repos/myrepo/myrepo.yaml

4. Instantiate the change to the infra with

```bash
forjj maintain ~/devops/myforj
```

In the short future, following should occur:

The flow should create a new local branch to work on. You do any kind of update, through source code or `forjj update`.
The jenkins `myforj_infra_PR` should be started pretty soon and shown GREEN status in github.
The PR job could be there to validate that forjj could maintain this repository with this code. (TBD)
Then merging this PR will start the `myforj_infra_master` job. This job will call forjj maintain to create the repo with github/jenkins integration (3 new jobs)


For details about the POC, and how to build it, see [this](POC.md)
For contribution, see [this](CONTRIBUTION.md)

Thank you
Forj team
