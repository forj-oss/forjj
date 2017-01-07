if [ "$FORJJ_PATH" = "" ]
then
   FORJJ_PATH=$PATH
   export PATH=$(pwd)/bin:$PATH
   echo "Build env loaded. To unload it, use 'source build-unset.sh"
fi
