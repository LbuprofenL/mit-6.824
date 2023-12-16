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
	//0:need to be processed
	//1:processing
	//2:processed
	mapState    map[int]int //mapId -> state
	reduceState map[int]int //reduceId ->state
	fileMap     map[int]string
	nReduce     int

	mapFinishedCnt    int
	reduceFinishedCnt int

	phase string
	mu    sync.Mutex
}

func (c *Coordinator) Request(args *RequestArgs, reply *RequestReply) error {
	switch c.phase {
	case "Map":
		err := c.MapRequest(args, reply)
		if err != nil {
			return err
		}
	case "Reduce":
		err := c.ReduceRequest(args, reply)
		if err != nil {
			return err
		}
	case "Done":
		reply.Finished = true
	}
	return nil
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) MapRequest(args *RequestArgs, reply *RequestReply) error {
	for mapId, state := range c.mapState {
		if state == 0 {
			reply.mu.Lock()
			reply.Filename = c.fileMap[mapId]
			reply.NReduce = c.nReduce
			reply.MapId = mapId
			reply.ReduceId = -1
			reply.TaskType = c.phase
			reply.mu.Unlock()

			c.mu.Lock()
			c.mapState[mapId] = 1
			c.mu.Lock()
			log.Printf("mapRequest Changed State to: %v", c.mapState)
			t := time.NewTimer(time.Second * 10)
			go func(mapId int) {
				<-t.C
				if c.mapState[mapId] == 1 {
					c.mu.Lock()
					c.mapState[mapId] = 0
					c.mu.Unlock()

					log.Printf("mapRequest Changed State to: %v", c.mapState)
				}
			}(mapId)
		}
	}
	return nil
}
func (c *Coordinator) ReduceRequest(args *RequestArgs, reply *RequestReply) error {
	for reduceId, state := range c.reduceState {
		if state == 0 {
			reply.mu.Lock()
			reply.Filename = c.fileMap[reduceId]
			reply.NReduce = c.nReduce
			reply.ReduceId = reduceId
			reply.MapId = -1
			reply.TaskType = c.phase
			reply.mu.Unlock()

			c.mu.Lock()
			c.reduceState[reduceId] = 1
			c.mu.Lock()

			log.Printf("mapRequest Changed State to: %v", c.mapState)
			t := time.NewTimer(time.Second * 10)
			go func(reduceId int) {
				<-t.C
				if c.reduceState[reduceId] == 1 {
					c.mu.Lock()
					c.reduceState[reduceId] = 0
					c.mu.Unlock()

					log.Printf("mapRequest Changed State to: %v", c.mapState)
				}
			}(reduceId)
		}
	}
	return nil
}

func (c *Coordinator) TaskDone(args *DoneArgs, reply *DoneReply) error {
	if args.TaskType != c.phase {
		reply.Reset = true
	}
	return nil
}
func (c *Coordinator) MapTaskDone(args *DoneArgs, reply *DoneReply) error {
	c.mu.Lock()
	defer c.mu.Lock()
	c.mapState[args.MapId] = 2
	c.mapFinishedCnt++
	if c.mapFinishedCnt == len(c.mapState) {
		c.phase = "Reduce"
	}

	return nil
}
func (c *Coordinator) ReduceTaskDone(args *DoneArgs, reply *DoneReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reduceState[args.ReduceId] = 2
	c.reduceFinishedCnt++
	if c.reduceFinishedCnt == len(c.reduceState) {
		c.phase = "Done"
	}
	return nil
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

	if c.phase == "Done" {
		ret = true
	}
	return ret
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	//Get all the files
	//每个文件会有三种状态：未被处理 0，处理中 1，处理结束 2
	mapState := make(map[int]int)
	fileName := make(map[int]string)
	for mapId, filename := range files {
		mapState[mapId] = 0
		fileName[mapId] = filename
	}
	c.fileMap = fileName
	c.mapState = mapState
	c.nReduce = nReduce
	reduceState := make(map[int]int)
	for i := 0; i < nReduce; i++ {
		reduceState[i] = 0
	}
	c.reduceState = reduceState
	c.phase = "Map"
	c.server()
	return &c
}
