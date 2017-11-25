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
for MOD in ${MOD[@]}
do
    source lib/source-be-$MOD.sh
done

if [ "$http_proxy" != "" ]
then
   PROXY="-e http_proxy=$http_proxy -e https_proxy=$http_proxy -e no_proxy=$no_proxy"
fi

USER="-u $(id -u)"
echo "INFO! Run from docker container."

MOUNT=""
if [[ "$DOCKER_JENKINS_MOUNT" != "" ]]
then # Set if jenkins requires a different mount point
    MOUNT="-v $DOCKER_JENKINS_MOUNT"
else
    if [[ "$MOD" != "" ]]
    then
        for MOD in ${MOD[@]}
        do
            be_${MOD}_mount_setup
        done
    fi
fi

if [ -t 1 ]
then
   TTY="-t"
fi

function docker_run {
    if [[ "$MOD" != "" ]]
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
