#!/bin/bash
# 
# First version of forjj. It could be replaced by a nodejs or go program instead to simplify reading yaml configuration files and share function with the future new ui.

if [ "$1" = bash ]
then
   echo "Entering in the container itself."
   exec bash
fi

if [ "$1"  = help ] || [ "$1" = "" ]
then
   echo "forjj <action> <ORGA> [options]
where :
- <ORGA>: is a string representing the name of the organization space used to store repositories locally or in docker volume.
  if <ORGA> contains a PATH, this PATH will become the organization space.
  ortherwise a new volume alias will be created in docker to host the organization space.
- action can be :
  - create  : Used to create the initial DevOps organization, repos and set of tools installed and configured.
  - update  : Update the infra. Used to create/update/remove projects and infrastructure migration (for example from local jenkins to a mesos jenkins solution)
  - maintain: Used by your CI to update the infra from the 'infra' repository.
"
   exit
fi


if [ "$1" = create ] || [ "$1" = update ] || [ "$1" = maintain ]
then
   ACTION="$1"
   ORGA="$2"
   shift
   if [ "$CONTRIB_REPO" != "" ]
   then
      CONTRIB_REPO_ARG="-v $CONTRIB_REPO:/forjj-contribs"
   else
      CONTRIB_REPO_ARG="-v $(basename $ORGA)-forjj-contribs:/forjj-contribs"
   fi
   DOCKER_OPTS="-it --rm --shm-size=64m"
   DOCKER_VOLS="-v $ORGA:/devops"
   if [ "$SSH_DIR" != "" ]
   then
      DOCKER_VOLS="$DOCKER_VOLS -v $SSH_DIR:/home/devops/.ssh"
   fi
   DOCKER_IMG="docker.hos.hpecorp.net/devops/forjj"
   DOCKER_APP="/usr/local/bin/forjj-$ACTION.sh"
   # Determine docker connection
   if [[ "$DOCKER_HOST" =~ ".*://.*" ]]
   then
      echo "Using Docker Host at '$DOCKER_HOST'"
   else
      DOCKER_API_VERSION=$(sudo docker version -f '{{ .Server.Version }}' 2>&1 | sed 's/.*API version: \([0-9.]*\).*$/\1/g' | tail -n 1)
      if [ "$DOCKER_API_VERSION" != "" ]
      then
         echo "Using old Docker API version : $DOCKER_API_VERSION"
         export DOCKER_API_VERSION
      fi
   fi
   exec sudo -E docker run $DOCKER_OPTS $CONTRIB_REPO_ARG $DOCKER_VOLS $DOCKER_IMG $DOCKER_APP "$@"
fi

if [ "$(dirname $1)" = /usr/local/bin ]
then
   exec $@
fi

echo "Unknown task '$1'. See help."
