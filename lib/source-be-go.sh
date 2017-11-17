
function go_jenkins_context {
    set +x
    shift
    mkdir -p go-workspace/src
    cd go-workspace/src
    CONTEXT="$1"
    PROJECT=forjj
    ln -sf $CONTEXT $PROJECT
    echo "Moved to Go workspace environment : go-workspace/src/$PROJECT"
    cd $PROJECT
    echo "$CONTEXT/go-workspace" > .be-gopath
    echo "Defining '$CONTEXT/go-workspace' for GOPATH"
    echo "Running from PWD = $(pwd)"
    SHIFT=2
}

function go_check_and_set {
    if [ $# -ne 0 ]
    then
       if [ ! -d $1 ]
       then
          echo "Invalid gopath"
          return 1
       fi
       if [ ! -d $1/src ]
       then
          echo "At least, your GOPATH must have an src directory. Not found."
          return 1
       fi
       echo "$1" > .be-gopath
       echo "$1 added as GOPATH"
    fi

    if [ -f .be-gopath ]
    then
       gopath="$(cat .be-gopath)"
       if [ "$gopath" != "" ] && [ -d "$gopath" ]
       then
          BUILD_ENV_GOPATH="$GOPATH"
          export GOPATH="$gopath"
          echo "Local setting of GOPATH to '$GOPATH'"
       else
          echo "GOPATH = '$gopath' is invalid. Please update your .be-gopath"
       fi
    fi

    if [ "$GOPATH" = "" ]
    then
       echo "Missing GOPATH. Please set it, or define it in your local personal '.be-gopath' file"
       return
    fi

}

function go_unset {
    if [ "$BUILD_ENV_GOPATH" != "" ]
    then
      GOPATH="$BUILD_ENV_GOPATH"
      unset BUILD_ENV_GOPATH
    fi
}
