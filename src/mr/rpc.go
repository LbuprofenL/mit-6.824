package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

type RequestArgs struct {
	Pid int
}

type RequestReply struct {
	TaskType string
	// Filename string
	// NReduce  int
	// NMap     int
	// MapId    int
	// ReduceId int
	MapReply    MapReplyStruct
	ReduceReply ReduceReplyStruct

	Finished bool //已无任务
}

type MapReplyStruct struct {
	Filename string
	NReduce  int
	MapId    int
}
type ReduceReplyStruct struct {
	NMap     int
	ReduceId int
}
type DoneArgs struct {
	TaskType string
	ReduceId int
	MapId    int
}

type DoneReply struct {
	Reset bool //超时节点
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
