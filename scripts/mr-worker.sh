#!/usr/bin/env bash

#
# map-reduce 
#

ROOT_DIR=/home/mit-6.824-lab1
MAIN_DIR=${ROOT_DIR}/src/main
APP_DIR=${ROOT_DIR}/src/mrapps
MR_DIR=${ROOT_DIR}/src/mr
CONF_DIR=${ROOT_DIR}/conf
SHARED_DIR=${ROOT_DIR}/shared

TASK_NAME=$1
WORK_DIR=$2
INPUT_DIR=$3
INPUT_PREFIX=$4
OUTPUT_DIR=$5
OUTPUT_FILE=$6
NUM_REDUCER=$7

WORK_DIR=${ROOT_DIR}/${WORK_DIR}
INPUT_DIR=${ROOT_DIR}/${INPUT_DIR}
OUTPUT_DIR=${ROOT_DIR}/${OUTPUT_DIR}

# check if the args are correct.
if [ -z $TASK_NAME ] || [ -z $WORK_DIR ] || [ -z $INPUT_DIR ] || [ -z $INPUT_PREFIX ] || [ -z $OUTPUT_DIR ] || [ -z $OUTPUT_FILE ] || [ -z $NUM_REDUCER ]
then
  echo "Usage: mr-task.sh <task_name> <work_dir> <input_dir> <INPUT_PREFIX> <output_dir> <output_file> <num_reducer>"
  exit 1
fi

ISQUIET=$1
maybe_quiet() {
    if [ "$ISQUIET" == "quiet" ]; then
      "$@" > /dev/null 2>&1
    else
      "$@"
    fi
}

# make sure software is freshly built.
TARGET_APP="${TASK_NAME}.go"

# check if the work directory exists.
rm -rf ${WORK_DIR}
mkdir ${WORK_DIR} || exit 1
cd ${WORK_DIR} || exit 1

# echo `pwd` # echo $(pwd)

(cd $APP_DIR && go clean)
(cd ${MAIN_DIR} && go clean)
(cd ${APP_DIR} && go build  -buildmode=plugin $TARGET_APP) || exit 1

# (cd ${MAIN_DIR} && go build  mrcoordinator.go) || exit 1
(cd ${MAIN_DIR} && go build  mrworker.go) || exit 1
# (cd ${MAIN_DIR} && go build mr-split.go)

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

failed_any=0

echo '***' Starting split input file

#
# check if the input file exists.
#

if [ ! -f ${INPUT_DIR}/${INPUT_PREFIX}.txt ]
then
  echo '*** Error: input file missing'
  exit 1
fi

# maybe_quiet $TIMEOUT ${MAIN_DIR}/mr-split ${INPUT_DIR}/${INPUT_PREFIX}.txt $NUM_REDUCER ./${INPUT_PREFIX}

#########################################################

echo '***' Starting ${TASK_NAME} app.

# maybe_quiet $TIMEOUT ${MAIN_DIR}/mrcoordinator $NUM_REDUCER ${WORK_DIR}/${INPUT_PREFIX}-split/${INPUT_PREFIX}*.txt &
pid=$!

# give the coordinator time to create the sockets.
sleep 1

# start multiple workers.
(maybe_quiet $TIMEOUT ${MAIN_DIR}/mrworker ${APP_DIR}/${TASK_NAME}.so) &

#
# check if the output directory is missing.
#

# if [ ! -d ${OUTPUT_DIR} ]
# then
#   mkdir ${OUTPUT_DIR}
# fi

# wait

# cat mr-out-* > ${OUTPUT_DIR}/${OUTPUT_FILE}
# rm -rf out
# rm -rf *split
# rm -f mr-out*