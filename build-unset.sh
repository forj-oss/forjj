if [ "$BUILD_ENV_PATH" != "" ]
then
   export PATH=$BUILD_ENV_PATH
   unset BUILD_ENV_PATH
   unset PROMPT_ADDONS_BUILD_ENV
   unset BUILD_ENV_LOADED
   unset BUILD_ENV_PROJECT
   unalias build-env-unset
   alias build-env='if [ -f build-env.sh ] ; then source build-env.sh ; else echo "Please move to your project where build-env.sh exists." ; fi'

   if [ "$BUILD_ENV_GOPATH" != "" ]
   then
      GOPATH="$BUILD_ENV_GOPATH"
      unset BUILD_ENV_GOPATH
   fi
fi
