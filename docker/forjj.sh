#!/bin/bash -e
# 
# First version of forjj. It could be replaced by a nodejs or go program instead to simplify reading yaml configuration files and share function with the future new ui.

if [ "$1" = bash ]
then
   echo "Entering in the container itself."
   exec bash
fi

#if [ "$1"  = help ] || [ "$1" = "" ]
#then
#   echo "forjj <action> <ORGA> [options]
#where :
#- <ORGA>: is a string representing the name of the organization space used to store repositories locally or in docker volume.
#  if <ORGA> contains a PATH, this PATH will become the organization space.
#  ortherwise a new volume alias will be created in docker to host the organization space.
#- action can be :
#  - create  : Used to create the initial DevOps organization, repos and set of tools installed and configured.
#  - update  : Update the infra. Used to create/update/remove projects and infrastructure migration (for example from local jenkins to a mesos jenkins solution)
#  - maintain: Used by your CI to update the infra from the 'infra' repository.
#"
#   exit
#fi


if [ "$1" = create ] || [ "$1" = update ] || [ "$1" = maintain ]
then
   sudo chown -R devops:devops /devops
   ACTION="$1"
   shift
   exec /usr/local/bin/forjj-$ACTION.sh $@
fi

echo "Unknown task '$1'."
