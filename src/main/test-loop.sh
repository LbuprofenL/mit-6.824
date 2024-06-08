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
(cd ../../mrapps && go build $RACE -buildmode=plugin loop_pr.go) || exit 1

(cd .. && go build $RACE mrcoordinator.go) || exit 1
(cd .. && go build $RACE mrworker.go) || exit 1

failed_any=0

#########################################################
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
sort mr-out* | grep . > mr-pre-all

echo '***' Starting loop test.

mkdir mr-tmp-loop || exit 1
cd mr-tmp-loop || exit 1

maybe_quiet $TIMEOUT ../../mrcoordinator ../mr-out-* &
pid=$!
# 文件名
output_file="pagerank.txt"

# 行数
line_count=8397

# 生成文件
yes 1 | head -n "$line_count" > "$output_file"

# give the coordinator time to create the sockets.
sleep 1
# start multiple workers.
(maybe_quiet $TIMEOUT ../../mrworker ../../../mrapps/loop_pr.so) &
(maybe_quiet $TIMEOUT ../../mrworker ../../../mrapps/loop_pr.so) &
(maybe_quiet $TIMEOUT ../../mrworker ../../../mrapps/loop_pr.so) &

# wait for the coordinator to exit.
wait $pid
sort mr-out* | grep . > mr-loop-all
wait
