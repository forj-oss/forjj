#/bin/bash

#CGO_ENABLED=0 go install
go install

scp  -P5001 $GOPATH/bin/forjj lacws.emea.hpqcorp.net:/storage/install/published/larsonsh/forjj
