
function go_jenkins_context {
    return
}

function go_set_path {
    PATH=$GOPATH/bin:$PATH
}

function go_check_and_set {
    if [[ "$CI_ENABLED" = "TRUE" ]]
    then # The CI will configure the GOPATH automatically at docker call.
        return
    fi

    # Local setup
    if [[ $# -ne 0 ]]
    then
       if [[ ! -d $1 ]]
       then
          echo "Invalid gopath"
          return 1
       fi
       if [[ ! -d $1/src ]]
       then
          echo "At least, your GOPATH must have an src directory. Not found."
          return 1
       fi
       echo "$1" > .be-gopath
       echo "$1 added as GOPATH"
    fi

    if [[ -f .be-gopath ]]
    then
       local gopath="$(cat .be-gopath)"
       if [[ "$gopath" != "" ]] && [[ -d "$gopath" ]]
       then
          BUILD_ENV_GOPATH="$GOPATH"
          export GOPATH="$gopath"
          echo "Local setting of GOPATH to '$GOPATH'"
       else
          echo "GOPATH = '$gopath' is invalid. Please update your .be-gopath"
       fi
    fi

    if [[ "$GOPATH" = "" ]]
    then
       echo "Missing GOPATH. Please set it, or define it in your local personal '.be-gopath' file"
       return
    fi

}

function unset_go {
    if [[ "$BUILD_ENV_GOPATH" != "" ]]
    then
      GOPATH="$BUILD_ENV_GOPATH"
      unset BUILD_ENV_GOPATH
      local fcts="`compgen -A function go_`"
      unset -f $fcts
    fi
}

function be_create_wrapper_go {
    case $1 in
        go)
            cat $BASE_DIR/modules/go/bin/go.sh >> $2
            ;;
        glide)
            cat $BASE_DIR/modules/go/bin/glide.sh >> $2
            ;;
        create-go-build-env.sh)
            cat $BASE_DIR/modules/go/bin/create-be.sh >> $2
            ;;
    esac
}

function be_go_mount_setup {
    MOUNT="$MOUNT -v $GOPATH:/go -w /go/src/$BE_PROJECT"
}

function be_do_go_docker_run {
    if [[ "$CI_WORKSPACE" != "" ]]
    then # Jenkins workspace detected.
        START_DOCKER="$BUILD_ENV_DOCKER run -di $MOUNT $PROXY -e GOPATH=/go/workspace $USER $1"
        echo "Starting container : '$START_DOCKER'"
        local CONT_ID=$($START_DOCKER /bin/cat)
        shift
        if [[ $CONT_ID = "" ]]
        then
            echo "Unable to start the container"
            exit 1
        fi
        set -e
        $BUILD_ENV_DOCKER exec -i $CONT_ID mkdir -p $WORKSPACE/go-workspace/src $WORKSPACE/go-workspace/bin
        $BUILD_ENV_DOCKER exec -i $CONT_ID ln -sf $WORKSPACE/go-workspace /go/workspace
        $BUILD_ENV_DOCKER exec -i $CONT_ID ln -sf $WORKSPACE /go/workspace/src/$BE_PROJECT
        $BUILD_ENV_DOCKER exec -i $CONT_ID bash -c "cd /go/workspace/src/$BE_PROJECT ; eval \"$*\""
        $BUILD_ENV_DOCKER rm -f $CONT_ID
    else
        eval do_docker_run "$@"
    fi
}

function be_create_go_docker_build {
    cp -vrp $BASE_DIR/modules/go/glide build-env-docker/
}

beWrappers["go"]="go glide create-go-build-env.sh"
