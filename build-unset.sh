# Function loaded by build-env
unset_build_env

if [ "$BUILD_ENV_GOPATH" != "" ]
then
  GOPATH="$BUILD_ENV_GOPATH"
  unset BUILD_ENV_GOPATH
fi

unset CGO_ENABLED

