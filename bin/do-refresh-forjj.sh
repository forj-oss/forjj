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


set -e
mkdir -p ~/bin
cd ~/bin

set +e
echo "Refreshing do-refresh-forjj.sh..."
wget -q -O ~/bin/do-refresh-forjj.new https://github.com/forj-oss/forjj/master/bin/do-refresh-forjj.sh
echo "Downloading forjj..."
wget -q -O ~/bin/forjj.new https://github.com/forj-oss/forjj/releases/download/$VERSION/forjj
set -e
if [[ -f forjj ]] 
then
    chmod +x ~/bin/forjj.new
    if [[ "$OLD_FORJJ" != "$NEW_FORJJ" ]]
    then
        if [[ "$DIFF" != "" ]]
        then
            $DIFF --side-by-side <(/home/larsonsh/src/forj/bin/forjj --version 2>/dev/null| sed 's/, /\n/g') <(forjj --version | sed 's/, /\n/g')
        else
            echo "Forjj has been updated:"
            echo "OLD: $(forjj --version 2>/dev/null)"
            echo "NEW: $(forjj.new --version 2>/dev/null)"
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

if [[ -f ~/bin/do-refresh-forjj.new ]] 
then
    chmod +x ~/bin/do-refresh-forjj.new
    mv ~/bin/do-refresh-forjj.new ~/bin/do-refresh-forjj
fi
