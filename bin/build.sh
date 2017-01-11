#/bin/bash -e

#CGO_ENABLED=0 go install

if [ "$BUILD_ENV_LOADED" != "true" ] 
then
   echo "Please go to your project and load your build environment. 'source build-env.sh'"
   exit 1
fi

cd $BUILD_ENV_PROJECT

BUILD_ENV=forjj-golang-env

if [ "$http_proxy" != "" ]
then
   PROXY="--build-arg http_proxy=$http_proxy --build-arg https_proxy=$http_proxy --build-arg no_proxy=$no_proxy"
fi

USER="--build-arg UID=$(id -u) --build-arg GID=$(id -g)"

set -x
sudo docker build -t $BUILD_ENV $PROXY $USER glide

go install

scp  -P5001 $GOPATH/bin/forjj lacws.emea.hpqcorp.net:/storage/install/published/larsonsh/forjj
