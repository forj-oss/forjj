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
PROJECT="$(basename $BUILD_ENV_PROJECT)"
echo "INFO! Run from docker container."

if [ "$GOPATH" = "" ]
then
   echo "GOPATH not set. Please set it."
   exit 1
fi

if [[ "$DOCKER_JENKINS_HOME" != "" ]]
then # Set if jenkins requires a different mount point
   MOUNT="-v $DOCKER_JENKINS_HOME -w $GOPATH/src/$PROJECT"
else
   MOUNT="-v $GOPATH:/go -w /go/src/$PROJECT"
fi

if [ -t 1 ]
then
   TTY="-t"
fi

function do_docker_run {
    eval $BUILD_ENV_DOCKER run --rm -i $TTY $MOUNT $PROXY $USER "$@"
}
