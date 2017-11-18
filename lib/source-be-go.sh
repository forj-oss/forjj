
function go_jenkins_context {
    return
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
       gopath="$(cat .be-gopath)"
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

function go_unset {
    if [[ "$BUILD_ENV_GOPATH" != "" ]]
    then
      GOPATH="$BUILD_ENV_GOPATH"
      unset BUILD_ENV_GOPATH
    fi
}
