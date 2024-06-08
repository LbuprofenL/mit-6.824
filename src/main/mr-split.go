package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func splitFileIntoParts(filePath string, parts int, outputDir string) error {
	// 打开输入文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建输出目录
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return err
	}

	// 计算每个文件大约应该包含的行数
	totalLines := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		totalLines++
	}
	linesPerPart := (totalLines + parts - 1) / parts // 确保最后一部分也包含行

	// 重新打开文件，准备开始分割
	file.Seek(0, 0)
	scanner = bufio.NewScanner(file)
	partNum := 1
	var partContent strings.Builder // 用于保存当前部分的内容
	linesRead := 0

	for scanner.Scan() {
		line := scanner.Text()

		// 写入当前行到当前部分内容
		partContent.WriteString(line + "\n")
		linesRead++

		// 如果已读取的行数达到了每个文件应该包含的行数，进行文件分割
		if linesRead >= linesPerPart {
			partFilePath := filepath.Join(outputDir, fmt.Sprintf("%s-%d.txt", filepath.Base(filePath)[:len(filepath.Base(filePath))-len(filepath.Ext(filePath))], partNum))
			err := os.WriteFile(partFilePath, []byte(partContent.String()), 0644)
			if err != nil {
				return err
			}
			fmt.Printf("Part %d written to %s\n", partNum, partFilePath)

			// 清空当前部分内容，重置已读取的行数，并递增分割号
			partContent.Reset()
			linesRead = 0
			partNum++
		}
	}

	// 处理可能剩余的内容，将最后部分内容写入到文件
	if partContent.Len() > 0 {
		partFilePath := filepath.Join(outputDir, fmt.Sprintf("%s-%d.txt", filepath.Base(filePath)[:len(filepath.Base(filePath))-len(filepath.Ext(filePath))], partNum))
		err := os.WriteFile(partFilePath, []byte(partContent.String()), 0644)
		if err != nil {
			return err
		}
		fmt.Printf("Part %d written to %s\n", partNum, partFilePath)
	}

	fmt.Println("File split completed.")
	return nil
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run mr-split.go <input_file> <number_of_parts> <out_file>")
	}
	filePath := os.Args[1]                     // 替换为实际的文件路径
	num_parts, err := strconv.Atoi(os.Args[2]) // 替换为实际的分割数量
	if err != nil {
		fmt.Println("Error:", err)
	}
	outputPath := os.Args[3] + "-split"
	err = splitFileIntoParts(filePath, num_parts, outputPath)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
