#!/usr/bin/env bash
ROOT_DIR="/home/mit-6.824-lab1"
MAIN_DIR=${ROOT_DIR}/src/main
CONF_DIR=${ROOT_DIR}/conf
SHARED_DIR=${ROOT_DIR}/shared
# 确定运行位置
cd "$MAIN_DIR"

# 执行pre任务
go run start_worker.go ${CONF_DIR}/pre.yaml

# 等待pre任务完成
while true; do
    if [ -f "${SHARED_DIR}/pre/pre.txt" ]; then
        break
    fi
    sleep 1
done

