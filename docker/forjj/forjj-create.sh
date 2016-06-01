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

INFRA_REPO=infra

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
     --http-proxy)
       export http_proxy=$2
       export https_proxy=$2
       shift;shift;;
     --driver-*)
       TYPE=${1:10}
       ORCH[$TYPE]="$2" 
       if [ ! -x /forjj-contribs/$TYPE/$2/$2 ]
       then
          echo "The '$2' '$TYPE' orchestrator is invalid: Unable to find forjj-contribs/$TYPE/$2/$2 executable"
          exit 1
       fi
       shift;shift;;
   esac
 done  
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

REPO=$1 ; shift

/forjj-contribs/upstream/$ORCH_SCM/$ORCH_SCM upstream_configure $REPO "$@"

# Check that the SCM upstream orchestrator as set origin and upstream data.
if [ $? -ne 0 ]
then
   echo "SCM Upstream orchestrator '$ORCH_SCM' fails with return code $?. forjj aborted."
   exit 1
fi

if [ ! -d /devops/$REPO/.git ]
then
   echo "SCM Upstream orchestrator '$ORCH_SCM' fails to ensure '$REPO' is a git repository. forjj aborted."
   exit 1
fi

if [ $(git remote | grep -e "^origin$" -e '^upstream$' | wc -l) -ne 2 ]
then
   echo "SCM Upstream orchestrator '$ORCH_SCM' fails to ensure origin and upstream remote properly set in '$REPO' git repository. forjj aborted."
   exit 1
fi
}

echo "Analyzing '$1' repositories..."

ORGA_NAME="$(basename "$1")"

shift

cd /devops/

# If we do not found the ci subdir, then we consider that this path is not having a valid 
# forjj-contribs.
# this path can be git repo or a simple dir. No git task is done on this repo anyway.
if [ ! -d /forjj-contrib/ci ]
then
   git clone https://$GITHUB_HOST/$GITHUB_ORGANIZATION/$CONTRIB_REPO /forjj-contrib
fi

parse_options


# TODO: Bug: ORCH not updated

if [ "${ORCH[upstream]}" = "" ]
then
   echo "Using 'github' as default forjj SCM-Upstream orchestrator"
   ORCH[upstream]=github
fi

if [ "${ORCH[ci]}" = "" ]
then
   echo "Using 'jenkins-ci' as default forjj CI orchestrator"
   ORCH_SCM=jenkins-ci
fi

# This will update the infra repo, ensure infra is a cloned git copy, and add initial commits if needed.
check_upstream $INFRA_REPO --infra "$@"

# Check about CI service configuration in infra thanks to the CI orchestrator.
# Ask the CI orchestrator to check if the service is properly configured in `infra`.
# - Ensure CI code exist and up to date.
/forjj-contribs/ci/${ORCH[ci]}/${ORCH[ci]} ci_configure "$@"

if [ "$REPOS" != "" ]
then
   echo "Ensuring '$REPOS' exists and connected to the upstream..."
   
   for REPO in ${REPOS/;/ }
   do # This will update only infra repository (commit auto-created for each repo.)
      check_upstream $REPO "$@"
   done
fi

# Process done!
# Infra is updated

# The next step should be to maintain the service as described by the infra repository.
# ie : repos exists, services exists up and running and properly configured.
