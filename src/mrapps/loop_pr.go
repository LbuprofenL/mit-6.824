package main

//
// a pagerank application "plugin" for MapReduce.
//
// go build -buildmode=plugin pre_pr.go
//

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"6.5840/mr"
)

const pageNum = 8398 // number of pages

// The map function is called once for each file of input. The first
// argument is the name of the input file, and the second is the
// file's complete contents. You should ignore the input file name,
// and look only at the contents argument. The return value is a slice
// of key/value pairs.
func Map(filename string, contents string) []mr.KeyValue {

	// split contents into an array of lines.
	lines := strings.Split(contents, "\n")

	kva := []mr.KeyValue{}
	for _, line := range lines {

		// split line into an array of number pairs by space.
		num_pairs := strings.Split(line, " ")

		for index, nums := range num_pairs {
			if index == 0 {
				continue
			}
			// split number pair into an array of numbers.
			num := strings.Split(nums, ":")
			if len(num) != 2 {
				continue
			}
			new_pair := fmt.Sprintf("%s:%s", num_pairs[0], num[1])

			kv := mr.KeyValue{Key: num[0], Value: new_pair}
			kva = append(kva, kv)
		}
	}
	return kva
}

// The reduce function is called once for each key generated by the
// map tasks, with a list of all the values created for that key by
// any map task.
func Reduce(key string, values []string) string {

	// process the pagerank file.
	// read the current pagerank file
	pr_file, err := os.Open("/home/mit-6.824-lab1/shared/loop/loop.txt")
	if err != nil {
		return err.Error()
	}
	defer pr_file.Close()
	pr_contents, err := io.ReadAll(pr_file)
	if err != nil {
		return err.Error()
	}

	// split the pagerank file into an array of lines.
	pr_lines := strings.Split(string(pr_contents), "\n")

	var pr_arr [pageNum]float64
	for _, v := range pr_lines {
		if v == "" {
			continue
		}
		nums := strings.Split(v, " ")
		if len(nums) != 2 {
			continue
		}

		col, err := strconv.ParseInt(nums[0], 10, 64)
		if err != nil {
			return err.Error()
		}

		// convert the number to float64.
		vv, err := strconv.ParseFloat(nums[1], 64)
		if err != nil {
			return err.Error()
		}
		pr_arr[col] = vv
	}

	// process weights
	var cur_row [pageNum]float64
	for _, v := range values {
		// split number pair into an array of numbers.
		nums := strings.Split(v, ":")

		// convert the number to float64.
		col, err := strconv.ParseInt(nums[0], 10, 64)
		if err != nil {
			return err.Error()
		}

		vv, err := strconv.ParseFloat(nums[1], 64)
		if err != nil {
			return err.Error()
		}

		cur_row[col] = vv
	}

	var ret float64
	//calculate the pagerank of the every visited page.
	for i := 1; i < pageNum; i++ {
		ret += float64(0.85) * pr_arr[i] * cur_row[i]
	}
	ret += float64(0.15) / (float64(pageNum - 1)) //每个页面都有可能跳转到这里
	ret_str := fmt.Sprintf("%.10f", ret)
	return ret_str
}
