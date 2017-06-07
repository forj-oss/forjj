#!/usr/bin/bash

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
   echo 'Usage is $0 latest|<tagName> <ReleaseName>'
   exit 0
fi

set -e
go get -d github.com/itchio/gothub
go build -o bin/gothub github.com/itchio/gothub
if [ "$(git rev-parse --abbrev-ref HEAD)" != master ]
then
   echo "You must be on master branch."
   exit 1
fi
git stash
git reset --hard upstream/master

if [ $# -eq 1 ]
then
   TAG=latest
else
   TAG="$1"
   RELEASE_NAME="$2"
fi

set +e
git tag -d $TAG
set -e

git tag $TAG
git push -f upstream latest

build.sh
export GITHUB_USER=forj-oss
export GITHUB_REPO=forjj
if [ "$TAG" = latest ]
then
   set +e
   gothub info -t latest
   if [ $? -ne 0 ]
   then
      # TODO: Remove hardcoded binary name.
      gothub release --tag $TAG --name forjj --description "Latest version of forjj."
   fi
   # TODO: Need else case
fi

gothub upload --tag $TAG --name forjj --file $GOPATH/bin/forjj --replace
