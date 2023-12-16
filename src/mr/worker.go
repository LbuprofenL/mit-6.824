package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

type ByKey []KeyValue

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

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
	fmt.Printf("Worker %v start\n", pid)
	failedCallCnt := 0
	for {
		requestArgs := RequestArgs{}
		requestArgs.Pid = pid
		requestReply := RequestReply{}
		ok := call("Coordinator.Request", &requestArgs, &requestReply)
		if failedCallCnt > 30 {
			break
		}
		if !ok {
			failedCallCnt++
			continue
		}

		if requestReply.Finished {
			break
		}
		doneArgs := DoneArgs{}
		doneReply := DoneReply{}
		switch requestReply.TaskType {
		case "Map":
			ok := execMapTask(mapf, requestReply.Filename, requestReply.MapId, requestReply.NReduce)
			if !ok {
				log.Printf("execMapTask failed, filename: %s, mapId: %d\n", requestReply.Filename, requestReply.MapId)
				continue
			}
			doneArgs.mu.Lock()
			doneArgs.MapId = requestReply.MapId
			doneArgs.TaskType = "Map"
			doneArgs.mu.Unlock()
		case "Reduce":
			ok := execReduceTask(reducef, requestReply.NMap, requestReply.ReduceId)
			if !ok {
				log.Printf("execReduceTask failed, filename: %d, reduceId: %d\n", requestReply.NMap, requestReply.ReduceId)
				continue
			}
			doneArgs.mu.Lock()
			doneArgs.ReduceId = requestReply.ReduceId
			doneArgs.TaskType = "Reduce"
			doneArgs.mu.Unlock()
		default:
			log.Printf("unknown task type: %s\n", requestReply.TaskType)
			continue
		}
		done := call("Coordinator.Done", &doneArgs, &doneReply)
		if !done {
			log.Printf("call Coordinator.Done failed\n")
			continue
		}
		if doneReply.Reset {
			continue
		}
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

func execReduceTask(reducef func(string, []string) string, nMap int, reduceId int) bool {
	kva := make([]KeyValue, 0)
	for i := 0; i < nMap; i++ {
		jsonName := fmt.Sprintf("mr-%d-%d", i, reduceId)
		file, err := os.Open(jsonName)
		if err != nil {
			log.Fatalf("cannot open %v", jsonName)
			return false
		}
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			kva = append(kva, kv)
		}
		file.Close()
	}
	//sort
	sort.Sort(ByKey(kva))
	//shuffle and reduce
	oname := fmt.Sprintf("mr-out-%d", reduceId)
	ofile, err := os.Create(oname)
	if err != nil {
		log.Fatalf("cannot create %v", oname)
		return false
	}
	defer ofile.Close()
	i := 0
	for i < len(kva) {
		j := i + 1
		for j < len(kva) && kva[j].Key == kva[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, kva[k].Value)
		}
		output := reducef(kva[i].Key, values)

		// this is the correct format for each line of Reduce output.
		fmt.Fprintf(ofile, "%v %v\n", kva[i].Key, output)
		i = j
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
