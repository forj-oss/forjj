#!/usr/bin/env bash

if [[ "$1" != "" ]]
then
    VERSION="$1"
else
    VERSION="latest"
fi

DIFF=$(which colordiff 2>/dev/null)
if [[ "$DIFF" = "" ]]
then
   DIFF=$(which diff 2>/dev/null)
fi

DOWNLOAD_PROG="$(which wget)"
DOWNLOAD_PROG_ARGS=" -q -O "
if [[ "$DOWNLOAD_PROG" = "" ]]
then
    DOWNLOAD_PROG="$(which curl)"
    DOWNLOAD_PROG_ARGS=" -s -o "
fi

if [[ "$DOWNLOAD_PROG" = "" ]]
then    
    echo "Unable to refresh forjj. Missing wget or curl. Please install one of them and retry."
    exit 1
fi

set -e
mkdir -p ~/bin
cd ~/bin

set +e
echo "Refreshing do-refresh-forjj.sh..."
$DOWNLOAD_PROG $DOWNLOAD_PROG_ARGS ~/bin/do-refresh-forjj.new https://github.com/forj-oss/forjj/raw/master/bin/do-refresh-forjj.sh
DO_REFRESH_STATUS=$?
echo "Downloading forjj..."
$DOWNLOAD_PROG $DOWNLOAD_PROG_ARGS ~/bin/forjj.new https://github.com/forj-oss/forjj/releases/download/$VERSION/forjj
set -e
if [[ -f forjj ]] 
then
    chmod +x ~/bin/forjj.new
    OLD_FORJJ="$(forjj --version 2>/dev/null)"
    NEW_FORJJ="$(forjj.new --version 2>/dev/null)"
    if [[ "$OLD_FORJJ" != "$NEW_FORJJ" ]]
    then
        if [[ "$DIFF" != "" ]]
        then
            $DIFF --side-by-side <(/home/larsonsh/src/forj/bin/forjj --version 2>/dev/null| sed 's/, /\n/g') <(forjj --version | sed 's/, /\n/g')
        else
            echo "Forjj has been updated:"
            echo "OLD: $OLD_FORJJ"
            echo "NEW: $NEW_FORJJ"
        fi
        rm -f forjj
        mv forjj.new forjj
    else
        echo "You already have the $VERSION version."
    fi
else
   mv forjj.new forjj 
fi
chmod +x forjj

if [[ $DO_REFRESH_STATUS -eq 0 ]] 
then
    if [[ -f ~/bin/do-refresh-forjj.new ]] 
    then
        chmod +x ~/bin/do-refresh-forjj.new
        mv ~/bin/do-refresh-forjj.new ~/bin/do-refresh-forjj.sh
    fi
else
    rm -f ~/bin/do-refresh-forjj.new
    echo "Unable to refresh the refresher script... wget https://github.com/forj-oss/forjj/raw/master/bin/do-refresh-forjj.sh fails."
fi
