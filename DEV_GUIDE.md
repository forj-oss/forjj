# Development guide

This guide explains how to modify or contribute to the Forjj  project.
Forjj is written in GO language, that needs to be compiled. This is the main difficulty, however some scripts have been written to really simplify this process.


## Repositories

To properly build Forjj you need to get the sources from github.

Forjj is composed of various repositories that are available from the project main page https://github.com/forj-oss .
- forjj --> main program, cli
- forjj-contribs --> plugins (github, jenkins)
- forjj-flows --> ?
- forjj-modules --> Modules used by Forjj cli
- forjj-repotemplates --> Templates for created project by Forjj
- goforjj --> GO templates in order to build plugins

You need to understand, which part of the project your modification will change and download the proper code repository.

Advice : so far do not change the name of the default project directory because it may create troubles building the project.

## Build tree

Using GO, you need to have a tree similar to this inside your home directory.

```
go
├── bin
├── pkg
│   └── linux_amd64
│       ├── github.com
│       ├── golang.org
│       └── gopkg.in
└── src
    ├── forjj
    └── forjj-contribs
```
- **go** is the root of go projects. You need to set the **GOPATH** environment variable to that directory.
- **bin** will host the binary files after a successful build. You should add this directory into your path.
- **pkg** will host the go packages needed as dependencies. There is no need to create this folder it will be created by the **go get** command.
- **src** should contains your go sources. This is the directory in which we need to clone the Forjj repositories.


## Build tools

The build process extensively uses **docker** to create a common build environment on any kind of Linux distribution.
- You need to have docker installed and configured on your host https://docs.docker.com/engine/installation/ .
- You need to configure your sudoers file to run docker as root without password.


## Proxy consideration

Efforts have been done to use and build the project behind a corporate proxy. If you are in this case, you need to set the **http_proxy** environment variable to the url of your corporate proxy. e.g:
```
export http_proxy=http://<proxyname>:<proxyport>
```
You can also define a **no_proxy** environment variable to avoid proxy usage on some hosts or subnets. e.g:
```
export no_proxy=10.3.0.0/24,myhost
```
Note, no_proxy variable should contain a comma separated list of domain extension, and you should not have spaces between the comas.

You need to also configure your docker daemon to use a proxy, some information here https://docs.docker.com/engine/admin/systemd/#httphttps-proxy .


## Build environment

Each repository will have one or several **build-env.sh** scripts and a **bin** directory to properly set the build environment and compile the required part.

By sourcing the build-env.sh, it will build a container **forjj-golang-env** that contains all the GO tools. Then create aliases and set priority in the path for all the scripts in the project bin directory. These scripts are wrappers around the usual GO commands to call the ones in the container instead of your system commands.

To build you need to:
1. Source ./build-env.sh
2. Run build.sh

Example to build Forjj

1. Source the environment.
```
[uggla@ugglalaptop forjj]$ . ./build-env.sh 
Build env loaded. To unload it, call 'build-env-unset'
```

2. Build Forjj, first step is to create the forjj-golang-env containeri because it does not exist.
```
[uggla@ugglalaptop forjj]$ build.sh
+ sudo docker build -t forjj-golang-env --build-arg UID=1000 --build-arg GID=1000 glide
Sending build context to Docker daemon  3.584kB
Step 1/10 : FROM golang:1.7.4
1.7.4: Pulling from library/golang
5040bd298390: Pull complete
fce5728aad85: Pull complete
76610ec20bf5: Pull complete
86b681f75ff6: Pull complete
8553b52886d8: Pull complete
63c25ee63bd6: Pull complete
4268eec6f44b: Pull complete
Digest: sha256:0b3787ac21ffb4edbd6710e0e60f991d5ded8d8a4f558209ef5987f73db4211a
Status: Downloaded newer image for golang:1.7.4
 ---> f3bdc5e851ce
Step 2/10 : MAINTAINER christophe.larsonneur@hpe.com
 ---> Running in c35b00fb01ba
 ---> 21ebd2e17337
Removing intermediate container c35b00fb01ba
Step 3/10 : ARG UID
 ---> Running in ab1ea1fef8ec
 ---> 36c3d998ed63
Removing intermediate container ab1ea1fef8ec
Step 4/10 : ARG GID
 ---> Running in bde06c92ec92
 ---> 0f7eb87f0af7
Removing intermediate container bde06c92ec92
Step 5/10 : ENV GLIDE_VERSION 0.12.3
 ---> Running in 83c49c992969
 ---> 43299c8dc037
Removing intermediate container 83c49c992969
Step 6/10 : ENV GLIDE_DOWNLOAD_URL https://github.com/Masterminds/glide/releases/download/v${GLIDE_VERSION}/glide-v${GLIDE_VERSION}-linux-amd64.tar.gz
 ---> Running in c57e3b38ab71
 ---> 7ced2406f3b6
Removing intermediate container c57e3b38ab71
Step 7/10 : RUN curl -fsSL "$GLIDE_DOWNLOAD_URL" -o glide.tar.gz     && tar -xzf glide.tar.gz     && mv linux-amd64/glide /usr/bin/     && rm -r linux-amd64     && rm glide.tar.gz
 ---> Running in 9ab1fb18e2ed
 ---> 490af6fae5f9
Removing intermediate container 9ab1fb18e2ed
Step 8/10 : RUN groupadd -g $GID developer     && useradd -u $UID -g $GID developer     && install -d -o developer -g developer -m 755 /go/src
 ---> Running in cc2c2745409e
 ---> 53390c9109f0
Removing intermediate container cc2c2745409e
Step 9/10 : COPY go-static /usr/local/go/bin
 ---> acad371a724a
Removing intermediate container 685c317df4a6
Step 10/10 : WORKDIR /go/src
 ---> 3bcbf1749d11
Removing intermediate container f894381b30c5
Successfully built 3bcbf1749d11
Successfully tagged forjj-golang-env:latest
+ set +x
'forjj-golang-env' image built.
```

3. Now the real build phase of Forjj, it uses glide to get all the project dependencies packages. The packages are placed into vendor directory. In this case, the dependencies are already found locally so not downloaded.
```
PROJECT = forjj
[INFO]	Downloading dependencies. Please wait...
[INFO]	--> Found desired version locally github.com/alecthomas/kingpin a328427ab7d619fe3c8d16a0da66899d03d5afae!
[INFO]	--> Found desired version locally github.com/alecthomas/template a0175ee3bccc567396460bf5acd36800cb10c49c!
[INFO]	--> Found desired version locally github.com/alecthomas/units 2efee857e7cfd4f3d0138cc3cbb1b4966962b93a!
[INFO]	--> Found desired version locally github.com/fatih/color 62e9147c64a1ed519147b62a56a14e83e2be02c1!
[INFO]	--> Found desired version locally github.com/forj-oss/forjj-modules d304ad5c6a969246676bbe9ebb248bc68232af5e!
[INFO]	--> Found desired version locally github.com/forj-oss/goforjj 88350d1636c503336a679e6d34a5872294968ae7!
[INFO]	--> Found desired version locally github.com/kr/text 7cafcd837844e784b526369c9bce262804aebc60!
[INFO]	--> Found desired version locally github.com/kvz/logstreamer a635b98146f0f465f3413a0b8ec481f3698d3fc2!
[INFO]	--> Found desired version locally github.com/mattn/go-colorable 5411d3eea5978e6cdc258b30de592b60df6aba96!
[INFO]	--> Found desired version locally github.com/mattn/go-isatty 57fdcb988a5c543893cc61bce354a6e24ab70022!
[INFO]	--> Found desired version locally github.com/moul/http2curl 4e24498b31dba4683efb9d35c1c8a91e2eda28c8!
[INFO]	--> Found desired version locally github.com/parnurzeal/gorequest a578a48e8d6ca8b01a3b18314c43c6716bb5f5a3!
[INFO]	--> Found desired version locally github.com/pkg/errors c605e284fe17294bda444b34710735b29d1a9d90!
[INFO]	--> Found desired version locally golang.org/x/net ddf80d0970594e2e4cccf5a98861cad3d9eaa4cd!
[INFO]	--> Found desired version locally golang.org/x/sys e24f485414aeafb646f6fca458b0bf869c0880a1!
[INFO]	--> Found desired version locally gopkg.in/alecthomas/kingpin.v2 7f0871f2e17818990e4eed73f9b5c2f429501228!
[INFO]	--> Found desired version locally gopkg.in/yaml.v2 cd8b52f8269e0feb286dfeef29f8fe4d5b397e0b!
[INFO]	Setting references.
[INFO]	--> Setting version for github.com/fatih/color to 62e9147c64a1ed519147b62a56a14e83e2be02c1.
[INFO]	--> Setting version for github.com/mattn/go-isatty to 57fdcb988a5c543893cc61bce354a6e24ab70022.
[INFO]	--> Setting version for github.com/kr/text to 7cafcd837844e784b526369c9bce262804aebc60.
[INFO]	--> Setting version for github.com/mattn/go-colorable to 5411d3eea5978e6cdc258b30de592b60df6aba96.
[INFO]	--> Setting version for golang.org/x/sys to e24f485414aeafb646f6fca458b0bf869c0880a1.
[INFO]	--> Setting version for github.com/moul/http2curl to 4e24498b31dba4683efb9d35c1c8a91e2eda28c8.
[INFO]	--> Setting version for github.com/parnurzeal/gorequest to a578a48e8d6ca8b01a3b18314c43c6716bb5f5a3.
[INFO]	--> Setting version for github.com/pkg/errors to c605e284fe17294bda444b34710735b29d1a9d90.
[INFO]	--> Setting version for github.com/kvz/logstreamer to a635b98146f0f465f3413a0b8ec481f3698d3fc2.
[INFO]	--> Setting version for github.com/alecthomas/template to a0175ee3bccc567396460bf5acd36800cb10c49c.
[INFO]	--> Setting version for github.com/alecthomas/units to 2efee857e7cfd4f3d0138cc3cbb1b4966962b93a.
[INFO]	--> Setting version for github.com/forj-oss/goforjj to 88350d1636c503336a679e6d34a5872294968ae7.
[INFO]	--> Setting version for github.com/alecthomas/kingpin to a328427ab7d619fe3c8d16a0da66899d03d5afae.
[INFO]	--> Setting version for github.com/forj-oss/forjj-modules to d304ad5c6a969246676bbe9ebb248bc68232af5e.
[INFO]	--> Setting version for golang.org/x/net to ddf80d0970594e2e4cccf5a98861cad3d9eaa4cd.
[INFO]	--> Setting version for gopkg.in/yaml.v2 to cd8b52f8269e0feb286dfeef29f8fe4d5b397e0b.
[INFO]	--> Setting version for gopkg.in/alecthomas/kingpin.v2 to 7f0871f2e17818990e4eed73f9b5c2f429501228.
[INFO]	Exporting resolved dependencies...
[INFO]	--> Exporting github.com/forj-oss/goforjj
[INFO]	--> Exporting github.com/alecthomas/kingpin
[INFO]	--> Exporting github.com/kr/text
[INFO]	--> Exporting github.com/mattn/go-colorable
[INFO]	--> Exporting github.com/mattn/go-isatty
[INFO]	--> Exporting github.com/alecthomas/template
[INFO]	--> Exporting github.com/pkg/errors
[INFO]	--> Exporting golang.org/x/sys
[INFO]	--> Exporting github.com/fatih/color
[INFO]	--> Exporting github.com/moul/http2curl
[INFO]	--> Exporting github.com/forj-oss/forjj-modules
[INFO]	--> Exporting github.com/alecthomas/units
[INFO]	--> Exporting github.com/kvz/logstreamer
[INFO]	--> Exporting github.com/parnurzeal/gorequest
[INFO]	--> Exporting golang.org/x/net
[INFO]	--> Exporting gopkg.in/yaml.v2
[INFO]	--> Exporting gopkg.in/alecthomas/kingpin.v2
[INFO]	Replacing existing vendor dependencies
```
4. Finally the build process that uses **go build** behind the scene.
```
PROJECT = forjj
'forjj' -> '/go/bin/forjj'

[uggla@ugglalaptop forjj]$ ll bin/forjj
-rwxr-xr-x 1 uggla uggla 10552499 Jul 24 13:14 bin/forjj
```

Note, in this case Forjj is a local binary. The plugin part build is a bit different as the result is a docker images with the required binaries.


Thank you
Forj team
