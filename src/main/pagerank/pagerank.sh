#!/usr/bin/env bash

# 执行pre任务
go run ../mr.go ../pre/pre.yaml

# 执行loop任务
go run ../mr.go ../loop/loop.yaml

# 执行normal任务
go run ../mr.go ../normal/normal.yaml

