package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

		pid := os.Getpid()
		var terminate = make(chan bool)
		go func() {
			CallRequest(pid)
		}



}

func CallRequest(mapf func(string, string) []KeyValue, id int) {
	args := MapArgs{}
	args.Id = id
	reply := MapReply{}

	ok := call("Coordinator.MapRequest", &args, &reply)
	if ok {
		execMapTask(mapf, reply.Filename, id, reply.NReduce)
	}
}

func execMapTask(mapf func(string, string) []KeyValue, filename string, mapId int, nReduce int) bool {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot open %v", filename)
		return false
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", filename)
		return false
	}
	file.Close()
	kva := mapf(filename, string(content))
	for _, kv := range kva {
		reduceId := ihash(kv.Key) % nReduce
		jsonName := fmt.Sprintf("mr-%d-%d", mapId, reduceId)
		saveJSON(jsonName, kv)
	}
	return true
}

func saveJSON(filename string, kv KeyValue) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("cannot create %v", filename)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(kv); err != nil {
		return err
	}
	return nil
}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
