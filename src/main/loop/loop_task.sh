#!/usr/bin/env bash

# 生成初始loop文件
# 文件名
output_file="loop.txt"
previous_weight=1
tolerance=0.01
# 行数
line_count=8397

# 生成文件
seq -f "%g 1" 1 $line_count > "$output_file"
cp "$output_file" "loop-0.txt"

# 执行loop任务 
cd ..
#循环执行20次
for i in {1..20}
do
    echo "Iteration $i"
    go run mr.go ./loop/loop.yaml > "loop_run_$i.log" 2>&1

    if [[ -f "./loop/loop.txt" ]]; then
        cp ./loop/loop.txt "./loop/loop-$i.txt"
    else
        echo "File ./loop/loop.txt does not exist."
        exit 1
    fi
   
    echo "Iteration $i completed"
done

echo "All iterations completed"
