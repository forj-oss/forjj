#
# File source to provide common build functions
#

if [ "$BUILD_ENV_LOADED" != "true" ]
then
   echo "Please go to your project and load your build environment. 'source build-env.sh'"
   exit 1
fi

cd $BUILD_ENV_PROJECT

source lib/build-env.fcts.sh
MODS=(`cat build-env.modules`)
for mod in ${MODS[@]}
do
    if [[ $mod != core ]]
    then
        source lib/source-be-${mod}.sh
    fi
done

if [ "$http_proxy" != "" ]
then
   PROXY="-e http_proxy=$http_proxy -e https_proxy=$http_proxy -e no_proxy=$no_proxy"
fi

USER="-u $(id -u)"

if [ -t 1 ]
then
   TTY="-t"
fi

function docker_run {
    MOUNT=""
    if [[ "$DOCKER_JENKINS_MOUNT" != "" ]]
    then # Set if jenkins requires a different mount point
        MOUNT="-v $DOCKER_JENKINS_MOUNT"
    else
        if [[ "$MOD" != "core" ]]
        then
            be_${MOD}_mount_setup
        fi
    fi
    if [[ "$MOD" != "core" ]]
    then
        be_do_${MOD}_docker_run "$@"
    else
        do_docker_run "$@"
    fi
}

function do_docker_run {
    _be_set_debug
    $BUILD_ENV_DOCKER run --rm -i $TTY $MOUNT $PROXY $USER "$@"
    _be_restore_debug
}
