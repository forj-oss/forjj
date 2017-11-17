# Source file to add to your build-en

function be_setup {
    if [ -f ~/.bashrc ] && [ "$(grep 'alias build-env=' ~/.bashrc)" = "" ]
    then
       echo "alias build-env='if [ -f build-env.sh ] ; then source build-env.sh ; else echo "Please move to your project where build-env.sh exists." ; fi'" >> ~/.bashrc
       echo "Alias build-env added to your existing .bashrc. Next time your could simply move to the project dir and call 'build-env'. The source task will done for you."
    fi
}

function be_docker_setup {
    if [ -f .be-docker ]
    then
       export BUILD_ENV_DOCKER="$(cat .be-docker)"
    else
       echo "Using docker directly. (no sudo)"
       export BUILD_ENV_DOCKER="docker"
    fi

    $BUILD_ENV_DOCKER version > /dev/null
    if [ $? -ne 0 ]
    then
       echo "$BUILD_ENV_DOCKER version fails. Check docker before going further. If you configured docker through sudo, please add --sudo:
    source build-env.sh --sudo ..."
       return 1
    fi

    $BUILD_ENV_DOCKER inspect forjj-golang-env > /dev/null
    if [ $? -ne 0 ]
    then
       bin/create-build-env.sh
    fi
}

function be_common_load {
    if [ "$BUILD_ENV_PATH" = "" ]
    then
       export BUILD_ENV_LOADED=true
       export BUILD_ENV_PROJECT=$(pwd)
       BUILD_ENV_PATH=$PATH
       export PATH=$(pwd)/bin:$PATH:$GOPATH/bin
       PROMPT_ADDONS_BUILD_ENV="BE: $(basename ${BUILD_ENV_PROJECT})"
       echo "Build env loaded. To unload it, call 'build-env-unset'"
       alias build-env-unset='cd $BUILD_ENV_PROJECT && source build-unset.sh'
    fi
}

function unset_build_env {
    if [ "$BUILD_ENV_PATH" != "" ]
    then
        export PATH=$BUILD_ENV_PATH
        unset BUILD_ENV_PATH
        unset PROMPT_ADDONS_BUILD_ENV
        unset BUILD_ENV_LOADED
        unset BUILD_ENV_PROJECT
        unalias build-env-unset
        alias build-env='if [ -f build-env.sh ] ; then source build-env.sh ; else echo "Please move to your project where build-env.sh exists." ; fi'

        # TODO: Be able to load from a defined list of jenkins context
        go_unset
    fi
}

# Core build env setup

be_setup

if [ $# -ne 0 ]
then
   if [ "$1" = "--sudo" ]
   then
      echo "sudo docker" > .be-docker
      shift
   fi

   # TODO: Be able to load from a defined list of jenkins context
   go_jenkins_context "$@"
   while [[ $SHIFT -gt 0 ]]
   do
        shift
        let SHIFT--
   done
fi

be_docker_setup

# TODO: Be able to load from a defined list of jenkins context
go_check_and_set "$@"

be_common_load

