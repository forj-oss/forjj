#
# File source to provide common build functions
#

if [ "$BUILD_ENV_LOADED" != "true" ]
then
   echo "Please go to your project and load your build environment. 'source build-env.sh'"
   exit 1
fi

cd $BUILD_ENV_PROJECT

if [ "$http_proxy" != "" ]
then
   PROXY="-e http_proxy=$http_proxy -e https_proxy=$http_proxy -e no_proxy=$no_proxy"
fi

USER="-u $(id -u)"
echo "INFO! Run from docker container."

if [[ "$DOCKER_JENKINS_MOUNT" != "" ]]
then # Set if jenkins requires a different mount point
    MOUNT="-v $DOCKER_JENKINS_MOUNT"
else
    # TODO: To move out of generic BE script
    MOUNT="-v $GOPATH:/go -w /go/src/$BE_PROJECT"
fi

if [ -t 1 ]
then
   TTY="-t"
fi

function do_docker_run {
    if [[ "$CI_WORKSPACE" != "" ]]
    then # Jenkins workspace detected.
        START_DOCKER="$BUILD_ENV_DOCKER run -di $MOUNT $PROXY -e GOPATH=/go/workspace $USER $1"
        echo "Starting container : '$START_DOCKER'"
        CONT_ID=$($START_DOCKER /bin/cat)
        shift
        if [[ $CONT_ID = "" ]]
        then
            echo "Unable to start the container"
            exit 1
        fi
        set -xe
        $BUILD_ENV_DOCKER exec -i $CONT_ID mkdir -p $WORKSPACE/go-workspace/src $WORKSPACE/go-workspace/bin
        $BUILD_ENV_DOCKER exec -i $CONT_ID ln -sf $WORKSPACE/go-workspace /go/workspace
        $BUILD_ENV_DOCKER exec -i $CONT_ID ln -sf $WORKSPACE /go/workspace/src/$BE_PROJECT
        $BUILD_ENV_DOCKER exec -i $CONT_ID bash -c "cd /go/workspace/src/$BE_PROJECT ; $*"
        $BUILD_ENV_DOCKER rm -f $CONT_ID
        set +x
    else
        eval $BUILD_ENV_DOCKER run --rm -i $TTY $MOUNT $PROXY $USER "$@"
    fi
}
