#!/usr/bin/env bash

ROOT_DIR=/home/mit-6.824-lab1
MAIN_DIR=${ROOT_DIR}/src/main
CONF_DIR=${ROOT_DIR}/conf
SHARED_DIR=${ROOT_DIR}/shared

# 确定运行位置
cd $MAIN_DIR

# 执行pre任务
go run mr.go ${CONF_DIR}/pre.yaml

# 执行loop任务
$SHARED_DIR/loop/loop_task.sh

# 执行normal任务
go run mr.go ${CONF_DIR}/normal.yaml

