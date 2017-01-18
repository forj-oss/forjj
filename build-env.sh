if [ -f ~/.bashrc ] && [ "$(grep 'alias build-env=' ~/.bashrc)" = "" ]
then
   echo "alias build-env='if [ -f build-env.sh ] ; then source build-env.sh ; else echo "Please move to your project where build-env.sh exists." ; fi'" >> ~/.bashrc
   echo "Alias build-env added to your existing .bashrc. Next time your could simply move to the project dir and call 'build-env'. The source task will done for you."
fi

if [ -f .gopath ]
then
   gopath="$(cat .gopath)"
   if [ "$gopath" != "" ] && [ -d "$gopath" ]
   then
      BUILD_ENV_GOPATH="$GOPATH"
      export GOPATH="$gopath"
      echo "Local setting of GOPATH to '$GOPATH'"
   else
      echo "GOPATH = '$gopath' is invalid. Please update your .gopath"
   fi
fi

if [ "$GOPATH" = "" ]
then
   echo "Missing GOPATH. Please set it, or define it in your local personal '.gopath' file"
   return
fi

if [ "$BUILD_ENV_PATH" = "" ]
then
   export BUILD_ENV_LOADED=true
   export BUILD_ENV_PROJECT=$(pwd)
   BUILD_ENV_PATH=$PATH
   export PATH=$(pwd)/bin:$PATH:$GOPATH/bin
   PROMPT_ADDONS_BUILD_ENV="BE: $(basename ${BUILD_ENV_PROJECT})"
   echo "Build env loaded. To unload it, call 'build-env-unset'"
   alias build-env-unset='cd $BUILD_ENV_PROJECT && source build-unset.sh'
fi

