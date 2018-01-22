#!/usr/bin/env bash

if [[ "$1" != "" ]]
then
    VERSION="$1"
else
    VERSION="latest"
fi

set -e
mkdir -p ~/bin
cd ~/bin
rm -f forjj
wget -O ~/bin/forjj https://github.com/forj-oss/forjj/releases/download/$VERSION/forjj
chmod +x forjj

./forjj --version
