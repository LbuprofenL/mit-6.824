#!/usr/bin/env bash

ROOT_DIR=/home/mit-6.824-lab1
MAIN_DIR=${ROOT_DIR}/src/main
CONF_DIR=${ROOT_DIR}/conf
SHARED_DIR=${ROOT_DIR}/shared
SCRIPTS_DIR=${ROOT_DIR}/scripts

# 确定运行位置
cd $MAIN_DIR

# 执行loop任务
$SCRIPTS_DIR/loop_task.sh

