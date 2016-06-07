#!/bin/bash
#
# This code implements the core forjj create task. Create do :
# - Update the infra repository with services configured (CI/...)
# - Update the infra `repos` with the list of repos upstream
#
# At the end, each repository are updated with at least initial commit.
# 
# At this time, only the UPSTREAM service will be started/configured.
# there is no requirement for the CI to be running.
# 

GITHUB_HOST=github.hpe.com
GITHUB_ORGANIZATION=christophe-larsonneur
CONTRIB_REPO=forjj-contribs

INFRA_REPO=christophe-larsonneur-infra

declare -A ORCH

function parse_options
{ # Parse and identify forjj core options.
 while [ $# -ne 0 ]
 do
   case "$1" in
     --repos)
       REPOS="$2"; shift;shift;;
     --no-proxy)
       export no_proxy=$2
       shift;shift;;
     --forjj-infra)
       INFRA_REPO=$2
       shift;shift;;
     --forjj-organization)
       ORGA_NAME=$2
       shift;shift;;
     --http-proxy)
       export http_proxy=$2
       export https_proxy=$2
       shift;shift;;
     --driver-*)
       TYPE=${1:9}
       ORCH[$TYPE]="$2" 
       if [ ! -x /forjj-contribs/$TYPE/$2/$2 ]
       then
          echo "The '$2' '$TYPE' driver is invalid: Unable to find forjj-contribs/$TYPE/$2/$2 executable"
          exit 1
       fi
       echo "Using '$2' '$TYPE' forjj driver"
       shift;shift;;
     *)
      shift;;
   esac
 done  
}

function json-test
{
 [ "$(echo "$1" | jq -r "$2")" = true ]
 return $?
}

function check_upstream
{
# Ask the SCM upstream orchestrator to check about $INFRA_REPO
# - Ensure the repo exist remotely (Service may be installed and started.)
# - Ensure the local dir is a git repo and upstream/origin is properly configured.
# - Ensure clone is up to date
# - Goodies added (README.md/CONTRIBUTION.md)
# All parameters are passed to the orchestrator for its task

ORCH_SCM=${ORCH[upstream]}
REPO="$1"
shift
# ### Checking the remote service
check_repo="$(/forjj-contribs/upstream/$ORCH_SCM/$ORCH_SCM check "$@")"
# If all is in place, return must be 0 and state_code = 200

# Check that the SCM upstream orchestrator as set origin and upstream data.
RET=$?
if [ $RET -ne 0 ]
then
   echo "SCM Upstream orchestrator '$ORCH_SCM' fails with return code $RET. forjj aborted."
   exit 1
fi

# Check return that service is up and found a repo. But it do not test the infra config file existence.
# But it gives the config name if i
if json-test "$check_repo" '.state_code == 200'
then
   return 1
fi

}

function ensure_local_repo_exist
{
 # ### Checking local repositories
 if [ ! -d /devops/$1/.git ]
 then
    git init /devops/$1
 fi
}

####################### MAIN

cd /devops/

# If we do not found the ci subdir, then we consider that this path is not having a valid 
# forjj-contribs.
# this path can be git repo or a simple dir. No git task is done on this repo anyway.
if [ ! -d /forjj-contribs/ci ]
then
   git clone https://$GITHUB_HOST/$GITHUB_ORGANIZATION/$CONTRIB_REPO /forjj-contribs
fi

parse_options "$@"

if [ "${ORCH[upstream]}" = "" ]
then
   echo "SCM-Upstream plugin missing. Please use --git-us to set one."
   exit 1
fi

# This check the infra repo, ensure infra is a cloned git copy, and add initial commits if needed.
if ! check_upstream $INFRA_REPO --infra "$@"
then
   echo "DevOps Infra solution is already implemented. You cannot use 'create' anymore. To build your workspace, use 'update'."
   exit 1
fi

ensure_local_repo_exist $INFRA_REPO
ensure_local_repo_exist ${INFRA_REPO}-state

RES="$(/forjj-contribs/upstream/${ORCH[upstream]}/${ORCH[upstream]} create "$@")"

if ! json-test "$RES" '.state-code == 200'
then
   echo "DevOps solution failure. Exiting."
   exit 1
fi

# Check about CI service configuration in infra thanks to the CI orchestrator.
# Ask the CI orchestrator to check if the service is properly configured in `infra`.
# - Ensure CI code exist and up to date.
#if [ "${ORCH[ci]}" ! = "" ]
#then 
#   /forjj-contribs/ci/${ORCH[ci]}/${ORCH[ci]} update "$@"
# fi

# Process done!
# Infra is updated

