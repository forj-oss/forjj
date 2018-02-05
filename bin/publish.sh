#!/usr/bin/bash

# This script is executed in a special context:
# - When a new PR has been merged to the master branch. It generates the latest version
# - when a decision is made to officially published a new release

# In both case, it assumes forjj git clone is configured as follow:
# - remote origin is pointing out to a fork
# - remote upstream is connected to forj-oss/forjj, with read/write access
# - The local branch must be master and should be connected to origin/master. Not upstream/master.

# If one of this condition is not respected, the script will exit.

if [ "$BUILD_ENV_LOADED" != "true" ]
then
   echo "Please go to your project and load your build environment. 'source build-env.sh'"
   exit 1
fi

if [ "$GITHUB_TOKEN" = '' ]
then
   echo "GITHUB_TOKEN is missing. You need it to publish."
   exit 1
fi

if [ $# -eq 0 ]
then
   echo 'Usage is $0 latest|version'
   exit 0
fi

set -e
if [ ! -f bin/gothub ]
then
   go get -d github.com/itchio/gothub
   go build -o bin/gothub github.com/itchio/gothub
fi

if [[ "$CI_ENABLED" = "FALSE" ]]
then
    echo  "You are going to publish manually. This is not recommended. Do it only if
    Jenkins fails to do it automatically.

    Press Enter to continue."
    read

    if [ "$(git rev-parse --abbrev-ref HEAD)" != master ]
    then
       echo "You must be on master branch. You are on '$(git rev-parse --abbrev-ref HEAD)'"
       exit 1
    fi

    REMOTE="$(git remote -v | grep "^upstream")"
    if [ "$REMOTE" = "" ]
    then
        echo "upstream is missing. You must have it configured (git@github.com:forj-oss/forjj.git) and rights given to push"
        exit 1
    fi
    if [[ ! "$REMOTE" =~ git@github\.com:forj-oss/forjj\.git ]]
    then
        echo "upstream is wrongly configured. It must be set with git@github.com:forj-oss/forjj.git"
        exit 1
    fi
    git stash # Just in case
    git fetch upstream
    git reset --hard upstream/master
fi

set +e
if [ "$1" = "latest" ]
then
   TAG=latest
   git fetch --tags
   git tag -d $TAG
else
   TAG="$(grep VERSION version.go | sed 's/const VERSION="\(.*\)"/\1/g')"
   PRE_RELEASE="$(grep VERSION version.go | sed 's/const PRERELEASE="\(.*\)"/\1/g')"
   if [ "$(git tag | grep "^$TAG$")" != "" ]
   then
      echo "Unable to publish $TAG. Already published and released."
      exit 1
   fi
   if [[ "$1" != "--auto" ]] && [[ "$CI_ENABLED" = "FALSE" ]]
   then
      echo "You are going to publish version $TAG. Ctrl-C to interrupt or press Enter to go on"
      read
   else
      echo "Publishing version $TAG..."
   fi
fi

set -e

echo "Tagging to $TAG..."
git tag $TAG

echo "Pushing it ..."
if [[ "$CI_ENABLED" = "TRUE" ]]
then
    git config --local credential.helper 'store --file /tmp/.git.store'
    echo "https://${GITHUB_USER}:${GITHUB_TOKEN}@github.com" > /tmp/.git.store
    git push -f origin $TAG
    rm -f /tmp/.git.store
    GOPATH=go-workspace
else
    git push -f upstream $TAG

    build.sh
fi
COMMIT_ID=$(git log --format=format:%H -1)
if [[ "$($GOPATH/bin/$BE_PROJECT --version | grep $COMMIT_ID)" = "" ]]
then
   echo "forjj binary is not in sync with current commit $COMMIT_ID"
   $GOPATH/bin/$BE_PROJECT --version
   exit 1
fi

echo "Deploying $BE_PROJECT to github..."
export GITHUB_REPO=$BE_PROJECT
# Removing the release everytime and ignore error.
set +e
gothub delete -t $TAG
set -ex
if [ "$TAG" = latest ]
then
   gothub release --tag $TAG --name $BE_PROJECT --description "Latest version of $BE_PROJECT." -p
else
    GOTHUB_PARS=""
   if [ "$PRE_RELEASE" = true ]
   then
      GOTHUB_PARS="-p"
   fi
   gothub release --tag $TAG --name $BE_PROJECT --description "$BE_PROJECT version $TAG." $GOTHUB_PARS
fi

# Checking binary build info
if [[ "$($GOPATH/bin/$BE_PROJECT --version | grep $COMMIT_ID)" = "" ]]
then
   echo "forjj binary is not in sync with current commit $COMMIT_ID"
   tmp/$BE_PROJECT --version
   exit 1
fi

gothub upload --tag $TAG --name $BE_PROJECT --file $GOPATH/bin/$BE_PROJECT --replace

mkdir -p tmp
rm -f tmp/$BE_PROJECT
curl -Lo tmp/$BE_PROJECT https://github.com/forj-oss/forjj/releases/download/$TAG/$BE_PROJECT
chmod +x tmp/$BE_PROJECT

if [[ "$(tmp/$BE_PROJECT --version | grep $COMMIT_ID)" = "" ]]
then
   echo "forjj binary is not in sync with current commit $COMMIT_ID"
   tmp/$BE_PROJECT --version
   exit 1
fi

rm tmp/$BE_PROJECT
