if [ "$FORJJ_PATH" != "" ]
then
   export PATH=$FORJJ_PATH
   unset FORJJ_PATH
   unset PROMPT_ADDONS_FORJJ
   unset BUILD_ENV_LOADED
   unset BUILD_ENV_PROJECT
   unalias build-env-unset
   alias build-env='if [ -f build-env.sh ] ; then source build-env.sh ; else echo "Please move to your project where build-env.sh exists." ; fi'
fi
