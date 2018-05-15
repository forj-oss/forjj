#!/bin/bash

if [[ "$1" =~ glide ]] && [[ "$GLIDE_HOME" != "" ]] && [[ ! -d $GLIDE_HOME ]]
then
   mkdir -vp $GLIDE_HOME
fi

exec "$@"