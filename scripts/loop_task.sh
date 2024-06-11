#!/usr/bin/env bash

# 生成初始loop文件
# 文件名
ROOT_DIR=/home/mit-6.824-lab1
MAIN_DIR=${ROOT_DIR}/src/main
CONF_DIR=${ROOT_DIR}/conf
SHARED_DIR=${ROOT_DIR}/shared

output_file="loop.txt"
previous_weight=1
tolerance=0.01
# 行数
line_count=8397

rm -f $output_file
# 生成文件
seq -f "%g 1" 1 $line_count > "$output_file"
cp "$output_file" "loop-0.txt"

# 执行loop任务 
cd $MAIN_DIR
#循环执行20次
for i in {1..20}
do
    echo "Iteration $i"
    /usr/lib/go/bin/go run start-.go $CONF_DIR/loop.yaml > "loop_run_$i.log" 2>&1

    if [[ -f "${SHARED_DIR}/loop/loop.txt" ]]; then
        cp ${SHARED_DIR}/loop/loop.txt "${SHARED_DIR}/loop/loop-$i.txt"
    else
        echo "File ${SHARED_DIR}/loop/loop.txt does not exist."
        exit 1
    fi
   
    echo "Iteration $i completed"
done

rm -f loop-*.txt

echo "All iterations completed"
