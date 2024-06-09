#!/usr/bin/env bash
 
# 确定运行位置
cd /home/mit-6.824-lab1/src/main

# 执行pre任务
go run mr.go ./pre/pre.yaml

# 执行loop任务
cd /home/mit-6.824-lab1/src/main
./loop/loop_task.sh

# 执行normal任务
cd  /home/mit-6.824-lab1/src/main
go run mr.go ./normal/normal.yaml

