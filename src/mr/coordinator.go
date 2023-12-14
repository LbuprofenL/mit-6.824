package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Coordinator struct {
	fileStateMap map[string]int
	taskMap      map[int]string
	nReduce      int
	mu           sync.Mutex
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) MapRequest(args *MapArgs, reply *MapReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for file, state := range c.fileStateMap {
		if state == 0 {
			reply.filename = file
			reply.nReduce = c.nReduce

			c.fileStateMap[file] = 1
			c.taskMap[args.Id] = file

			timer := time.NewTimer(time.Second * 10)
			go func() {
				<-timer.C
				if c.fileStateMap[file] != 2 {
					c.fileStateMap[file] = 0
				}
			}()
			break
		}
	}
	return nil
}

func (c *Coordinator) MapTaskDone(args *MapArgs, reply *MapReply) (error, bool) {

	c.mu.Lock()
	task := c.taskMap[args.Id]
	c.fileStateMap[task] = 2
	c.mu.Unlock()

	cnt := 0
	for _, state := range c.fileStateMap {
		if state == 2 {
			cnt++
		}
	}
	if cnt == len(c.fileStateMap) {
		return nil, true
	}
	return nil, false
}

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.

	return ret
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	//Get all the files
	//每个文件会有三种状态：未被处理 0，处理中 1，处理结束 2
	m := make(map[string]int)
	for _, filename := range files {
		m[filename] = 0
	}
	taskMap := make(map[int]string)
	c.fileStateMap = m
	c.taskMap = taskMap
	c.nReduce = nReduce
	c.server()
	return &c
}
