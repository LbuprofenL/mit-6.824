package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TaskName string `yaml:"task_name"`
	WorkDir  string `yaml:"work_dir"`
	Input    struct {
		InputDir    string `yaml:"input_dir"`
		InputPrefix string `yaml:"input_prefix"`
	} `yaml:"input"`
	Output struct {
		OutputDir      string `yaml:"output_dir"`
		OutputFilename string `yaml:"output_filename"`
	} `yaml:"output"`
	ReducerNum int `yaml:"reducer_num"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <config.yaml>", os.Args[1])
	}

	configFile := os.Args[1]

	// 读取 YAML 文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// 解析 YAML 文件
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// 生成参数字符串
	args := []string{
		config.TaskName,
		config.WorkDir,
		config.Input.InputDir,
		config.Input.InputPrefix,
		config.Output.OutputDir,
		config.Output.OutputFilename,
		fmt.Sprint(config.ReducerNum),
	}
	// 执行 MapReduce 任务
	fmt.Printf("Running MapReduce task with args: %s\n", args)

	cmd := exec.Command("../../scripts/mr-coord.sh", args...)
	// 设置命令的输出和错误输出
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 开始执行命令
	err = cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}

	// 监控命令的状态
	err = cmd.Wait()
	if err != nil {
		// 检查命令是否被信号终止
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			// 获取命令退出状态
			waitStatus := exitErr.Sys().(syscall.WaitStatus)
			exitCode := waitStatus.ExitStatus()
			fmt.Printf("Command exited with non-zero status: %d\n", exitCode)
		} else {
			// 其他错误
			log.Fatalf("Command failed: %v", err)
		}
	} else {
		fmt.Println("Command completed successfully")
	}
}
