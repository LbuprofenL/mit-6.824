package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	//读取../loop/loop.txt文件
	file, err := os.Open("../loop/loop.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	//读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	//将文件内容转换为字符串
	lines := strings.Split(string(content), "\n")
	pagerank := make(map[int]float64, 10000)
	var total_rank float64
	for _, line := range lines {
		num_str := strings.Split(line, " ")
		if len(num_str) == 2 {
			page, err := strconv.ParseInt(num_str[0], 10, 16)
			if err != nil {
				fmt.Println(err)
				return
			}
			rank, err := strconv.ParseFloat(num_str[1], 64)
			if err != nil {
				fmt.Println(err)
				return
			}
			pagerank[int(page)] = rank
			total_rank += rank
		}
	}

	//对rank规范化
	for page, rank := range pagerank {
		pagerank[page] = rank / total_rank
	}

	//输出结果到./pagerank.txt 并按rank值降序排列
	file, err = os.OpenFile("./pagerank.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	for page, rank := range pagerank {
		_, err = file.WriteString(fmt.Sprintf("%d %f\n", page, rank))
		if err != nil {
			fmt.Println(err)
			return
		}
	}

}
