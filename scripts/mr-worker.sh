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


# make sure software is freshly built.
TARGET_APP="${TASK_NAME}.go"

# check if the work directory exists.
# rm -rf ${WORK_DIR}
# mkdir ${WORK_DIR} || exit 1

#
# check if the input file exists.
#
echo '***' Waiting for file split.
split_count=$(ls -1 ${WORK_DIR}/${INPUT_PREFIX}-split | wc -l)
while [ $split_count -lt $NUM_REDUCER ]
do
  sleep 1
  echo $split_count
  split_count=$(ls -1 ${WORK_DIR}/${INPUT_PREFIX}-split | wc -l)
done


#########################################################

echo '***' Starting ${TASK_NAME} app.

cd ${WORK_DIR} || exit 1
# start multiple workers.
${MAIN_DIR}/mrworker ${APP_DIR}/${TASK_NAME}.so &
pid=$!

wait $pid
