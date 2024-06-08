#!/usr/bin/env bash

#
# map-reduce tests
#

# un-comment this to run the tests with the Go race detector.
# RACE=-race

if [[ "$OSTYPE" = "darwin"* ]]
then
  if go version | grep 'go1.17.[012345]'
  then
    # -race with plug-ins on x86 MacOS 12 with
    # go1.17 before 1.17.6 sometimes crash.
    RACE=
    echo '*** Turning off -race since it may not work on a Mac'
    echo '    with ' `go version`
  fi
fi

ISQUIET=$1
maybe_quiet() {
    if [ "$ISQUIET" == "quiet" ]; then
      "$@" > /dev/null 2>&1
    else
      "$@"
    fi
}


TIMEOUT=timeout
TIMEOUT2=""
if timeout 2s sleep 1 > /dev/null 2>&1
then
  :
else
  if gtimeout 2s sleep 1 > /dev/null 2>&1
  then
    TIMEOUT=gtimeout
  else
    # no timeout command
    TIMEOUT=
    echo '*** Cannot find timeout command; proceeding without timeouts.'
  fi
fi
if [ "$TIMEOUT" != "" ]
then
  TIMEOUT2=$TIMEOUT
  TIMEOUT2+=" -k 2s 120s "
  TIMEOUT+=" -k 2s 45s "
fi

# run the test in a fresh sub-directory.
rm -rf mr-tmp
mkdir mr-tmp || exit 1
cd mr-tmp || exit 1
rm -f mr-*

# make sure software is freshly built.
(cd ../../mrapps && go clean)
(cd .. && go clean)
(cd ../../mrapps && go build $RACE -buildmode=plugin pre_pr.go) || exit 1

(cd .. && go build $RACE mrcoordinator.go) || exit 1
(cd .. && go build $RACE mrworker.go) || exit 1
(cd .. && go build $RACE ./sequential/pre_sequential.go) || exit 1

failed_any=0

#########################################################
# first pre

# generate the correct output
echo '***' Starting pre_seq test.
../pre_sequential ../../mrapps/pre_pr.so ../pr.txt || exit 1
sort mr-out-0 > mr-correct-pre.txt
rm -f mr-out*

echo '***' Starting pre test.

maybe_quiet $TIMEOUT ../mrcoordinator ../pr/pr*txt &
pid=$!

# give the coordinator time to create the sockets.
sleep 1

# start multiple workers.
(maybe_quiet $TIMEOUT ../mrworker ../../mrapps/pre_pr.so) &
(maybe_quiet $TIMEOUT ../mrworker ../../mrapps/pre_pr.so) &
(maybe_quiet $TIMEOUT ../mrworker ../../mrapps/pre_pr.so) &

# wait for the coordinator to exit.
wait $pid

# since workers are required to exit when a job is completely finished,
# and not before, that means the job has finished.
# sort mr-out* | grep . > mr-pre-all
# if cmp mr-pre-all mr-correct-pre.txt
# then
#   echo '---' pre test: PASS
# else
#   echo '---' pre output is not the same as mr-correct-pre.txt
#   echo '---' pre test: FAIL
#   failed_any=1
# fi

# wait for remaining workers and coordinator to exit.
wait
