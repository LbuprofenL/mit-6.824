#!/bin/bash

ROOT_DIR=/home/mit-6.824-lab1
MAIN_DIR=${ROOT_DIR}/src/main
APP_DIR=${ROOT_DIR}/src/mrapps
MR_DIR=${ROOT_DIR}/src/mr
CONF_DIR=${ROOT_DIR}/conf
SHARED_DIR=${ROOT_DIR}/shared

(cd ${MAIN_DIR} && go build  mrcoordinator.go) || exit 1
# (cd ${MAIN_DIR} && go build  mrworker.go) || exit 1
(cd ${MAIN_DIR} && go build mr-split.go)
(cd ${APP_DIR} && go build  -buildmode=plugin pre_pr.go) || exit 1
(cd ${APP_DIR} && go build  -buildmode=plugin loop_pr.go) || exit 1
(cd ${APP_DIR} && go build  -buildmode=plugin normal_pr.go) || exit 1
(cd ${MAIN_DIR} && go build  mrworker.go) || exit  1