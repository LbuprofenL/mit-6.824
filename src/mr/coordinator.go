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
	log.Printf("received request: %v", args)
	switch c.phase {
	case "Map":
		log.Printf("received map request: %v", args)
		err := c.mapRequest(args, reply)
		if err != nil {
			return err
		}
		log.Printf("send map reply: %v", reply)
	case "Reduce":
		log.Printf("received reduce request: %v", args)
		err := c.reduceRequest(args, reply)
		if err != nil {
			return err
		}
		log.Printf("send reduce reply: %v", reply)
	case "Done":
		log.Printf("received done request: %v", args)
		reply.Finished = true
	}
	return nil
}

func (c *Coordinator) mapRequest(args *RequestArgs, reply *RequestReply) error {
	log.Printf("mapRequest: %v", args)
	for mapId, state := range c.mapState {
		if state == 0 {
			reply.Filename = c.fileMap[mapId]
			reply.NReduce = c.nReduce
			reply.MapId = mapId
			reply.TaskType = c.phase

			c.mu.Lock()
			c.mapState[mapId] = 1
			c.mu.Unlock()
			log.Printf("mapRequest Changed State to: %v", c.mapState)
			t := time.NewTimer(time.Second * 10)
			go func(id int) {
				<-t.C
				if c.mapState[id] == 1 {
					c.mu.Lock()
					c.mapState[id] = 0
					c.mu.Unlock()

					log.Printf("mapRequest has changed State to: %v,becasuse of timeout.", c.mapState)
				}
			}(mapId)
			break
		}
	}
	return nil
}
func (c *Coordinator) reduceRequest(args *RequestArgs, reply *RequestReply) error {
	log.Printf("reduceRequest: %v", args)
	for reduceId, state := range c.reduceState {
		if state == 0 {
			reply.NMap = len(c.mapState)
			reply.ReduceId = reduceId
			reply.TaskType = c.phase

			c.mu.Lock()
			c.reduceState[reduceId] = 1
			c.mu.Unlock()

			log.Printf("reduceRequest Changed State to: %v", c.reduceState)
			t := time.NewTimer(time.Second * 10)
			go func(id int) {
				<-t.C
				if c.reduceState[id] == 1 {
					c.mu.Lock()
					c.reduceState[id] = 0
					c.mu.Unlock()

					log.Printf("ruduceRequest has changed State to: %v,becasuse of timeout.", c.reduceState)
				}
			}(reduceId)
			break
		}
	}
	return nil
}

func (c *Coordinator) TaskDone(args *DoneArgs, reply *DoneReply) error {
	if args.TaskType != c.phase {
		reply.Reset = true
		return nil
	}
	if args.TaskType == "Map" {
		err := c.mapTaskDone(args, reply)
		if err != nil {
			return err
		}
	}

	if args.TaskType == "Reduce" {
		err := c.reduceTaskDone(args, reply)
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *Coordinator) mapTaskDone(args *DoneArgs, reply *DoneReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mapState[args.MapId] = 2
	log.Printf("mapTaskDone changed state to: %v", c.mapState)
	c.mapFinishedCnt++
	if c.mapFinishedCnt == len(c.mapState) {
		log.Printf("mapTaskDone all map task finished,change phase to reduce")
		c.phase = "Reduce"
	}
	return nil
}
func (c *Coordinator) reduceTaskDone(args *DoneArgs, reply *DoneReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reduceState[args.ReduceId] = 2
	log.Printf("reduceTaskDone changed state to: %v", c.reduceState)
	c.reduceFinishedCnt++
	if c.reduceFinishedCnt == len(c.reduceState) {
		log.Printf("reduceTaskDone all reduce task finished,change phase to Done")
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
	logname := "mrcoordinator.log"
	logfile, err := os.OpenFile(logname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(logfile)
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
	c.mapFinishedCnt = 0
	c.reduceFinishedCnt = 0
	c.server()
	return &c
}
