# Source file to add to your build-env.sh

source lib/build-env.fcts.sh

MODS=(`cat build-env.modules`)
for MOD in ${MOD[@]}
do
    source lib/source-be-$MOD.sh
done

# Core build env setup

set +xe
be_valid
if [[ $? -ne 0 ]]
then
    return
fi

be_ci_detected

be_setup

if [ $# -ne 0 ]
then
   if [ "$1" = "--sudo" ]
   then
      echo "sudo docker" > .be-docker
      shift
   fi

   # TODO: Be able to load from a defined list of build env type. Here it is GO
   go_jenkins_context "$@"
   while [[ $SHIFT -gt 0 ]]
   do
        shift
        let SHIFT--
   done
fi

be_docker_setup

# TODO: Be able to load from a defined list of build env type. Here it is GO
go_check_and_set "$@"

be_common_load

be_ci_run
