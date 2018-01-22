# Introduction

Thank you for your interest in Forjj.

This user guide is currently really basic and will need to be strongly
reworked, as soon as Forjj has all wanted features and stable.

## About Forjj

Forjj is a simple static binary.

You can easily download it from github.

The latest version is always available. But be aware that you can find
out some instability, regression as Forjj is still in development.

At this time of writing, Forjj can build and maintain a forge
composed by :
- 'github' (public or private)
- Jenkins started from Docker.

The Forjj flow which integrates `Jenkins` and `Github` to automatically
build any repository containing a Jenkinsfile is still under heavy
development and should be ready in couple of weeks.

## Usage of Forjj

Forjj is used by Forjj to build Forjj.
This is our 1st use case. We still need to enhance the automation part
Most of pending issues are in each repository issues section.

Forjj is also used internally in my company at several places.

In short, today, there is about 4 Forges managed by Forjj.

## Forjj Plugins

Forjj itself can't do any work alone.

It requires plugins to work.
Today, there is only `Jenkins` and `github`.

Those plugins are stored in https://github.com/forj-oss/forjj-contribs

To use them, we declare them in a Forjfile.
The available options in Forjfile depends on the plugin.

As of now, there is no real good documentation on those options.
They are just defined respectively in their `github.yaml` or `jenkins.yaml`
files in the Source.

A user story has been written to create a dynamic plugin help.
Everything information are in those <plugin>.yaml file. We 'just' need
to present it in the Forjj help...

# Forjj User guide

## Installing / updating Forjj

To install or update Forjj, you can run the following code:

```bash
curl https://raw.githubusercontent.com/forj-oss/forjj/master/bin/do-refresh-forjj.sh | bash
```

If you want to simplify the update of latest Forjj version, you can
install this small script and run it anytime you want:

```bash
curl -o ~/bin/do-refresh-forjj.sh https://raw.githubusercontent.com/forj-oss/forjj/master/bin/do-refresh-forjj.sh
chmod +x ~/bin/do-refresh-forjj.sh
```

## Your Forjfile

To start your Forge with Forjj, you need to create a Forjfile
with the list of applications you want and the organization name:

```yaml
forjj-settings:
  organization: myOrg
applications:
  github:
    type: upstream
  jenkins:
    type: ci
  <...>
```

Depending where you have your running services or where you want to start
them, you will need to add some extra parameters like `server`

Look in https://github.com/forj-oss/forjj-contribs/blob/master/upstream/github/github.yaml#L27
to see which parameters can be set under github.

Look in https://github.com/forj-oss/forjj-contribs/blob/master/ci/jenkins/jenkins.yaml#L32
to see which parameters can be set under jenkins.

Please note that group flags are parameters prefixed by the group name

Ex:
```yaml
objects:
  app:
    groups:
      dockerfile:
        flags:
          # Information we can define for the Dockerfile.
          from-image:
            help: "Base Docker image tag name to use in Dockerfile. Must respect [server/repo/]name."
default: forjdevops/jenkins
```

Following this jenkins plugin definition, you can add a `<group>-<flag>`
. In this example it will be `dockerfile-from-image`

All other objects defined by the plugin can be set in the Forjfile as well:

Ex:
```yaml
objects:
  projects:
    default-actions: ["add", "change", "remove"]
    identified_by_flag: name
    flags:
      name:
        help: "Project name"
        required: true
      remote-type:
        default: "{{ $Project := .Current.Name }}{{ (index .Forjfile.Repos $Project).RemoteType }}"
        help: "Define remote source  type. 'github' is used by default. Support 'git', 'github'."
    groups:
      github:
        flags:
          api-url:
            default: "{{ $Project := .Current.Name }}{{ (index .Forjfile.Repos $Project).UpstreamAPIUrl }}"
            help: "with remote-type = 'github', Github API Url. By default, it uses public github API."
          repo-owner:
            default: "{{ $Project := .Current.Name }}{{ (index .Forjfile.Repos $Project).Owner }}"
            help: "with remote-type = 'github', Repository owner. Can be a user or an organization."
          repo:
            default: "{{ .Current.Name }}"
            help: "with remote-type = 'github', Repository name."
```

So in your Forjfile, you can add:
```yaml
projects:
  <projectName>:
    github-api-url: https://...
```

In this example, `<projectName>` is your project name, identified as `name`
and you set a group flag called github and a flag called `api-url`

# More to come.

The documentation is in progress.
