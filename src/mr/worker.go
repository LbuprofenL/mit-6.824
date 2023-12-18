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
	logname := fmt.Sprintf("mr-worker-%v.log", pid)
	logfile, err := os.OpenFile(logname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(logfile)
	fmt.Printf("Worker %v start\n", pid)
	failedCallCnt := 0
	for {
		requestArgs := RequestArgs{}
		requestArgs.Pid = pid
		requestReply := RequestReply{MapReply: MapReplyStruct{}, ReduceReply: ReduceReplyStruct{}}
		ok := call("Coordinator.Request", &requestArgs, &requestReply)
		if failedCallCnt > 30 {
			log.Printf("Worker %v failed to call Coordinator.Request %d times\n", pid, failedCallCnt)
			break
		}
		if !ok {
			failedCallCnt++
			continue
		}

		if requestReply.Finished {
			log.Printf("Worker %v finish\n", pid)
			break
		}
		doneArgs := DoneArgs{}
		doneReply := DoneReply{}
		switch requestReply.TaskType {
		case "Map":
			log.Printf("Worker %v start map task %d\n", pid, requestReply.MapReply.MapId)
			err := execMapTask(mapf, requestReply.MapReply.Filename, requestReply.MapReply.MapId, requestReply.MapReply.NReduce)
			if err != nil {
				log.Printf("execMapTask failed, filename: %s, mapId: %d\n", requestReply.MapReply.Filename, requestReply.MapReply.MapId)
				log.Fatalf("execMapTask failed: %v", err)
				continue
			}
			log.Printf("Worker %v finish map task %d\n", pid, requestReply.MapReply.MapId)
			doneArgs =
				DoneArgs{MapId: requestReply.MapReply.MapId, TaskType: "Map"}
		case "Reduce":
			log.Printf("Worker %v start reduce task %d\n", pid, requestReply.ReduceReply.ReduceId)
			err := execReduceTask(reducef, requestReply.ReduceReply.NMap, requestReply.ReduceReply.ReduceId)
			if err != nil {
				log.Printf("execReduceTask failed, filename: %d, reduceId: %d\n", requestReply.ReduceReply.NMap, requestReply.ReduceReply.ReduceId)
				log.Fatalf("execReduceTask failed: %v", err)
				continue
			}
			log.Printf("Worker %v finish reduce task %d\n", pid, requestReply.ReduceReply.ReduceId)
			doneArgs =
				DoneArgs{ReduceId: requestReply.ReduceReply.ReduceId, TaskType: "Reduce"}
		default:
			log.Printf("unknown task type: %s\n", requestReply.TaskType)
			continue
		}
		done := call("Coordinator.TaskDone", &doneArgs, &doneReply)
		if !done {
			log.Printf("call Coordinator.Done failed\n")
			continue
		}
		if doneReply.Reset {
			continue
		}
	}

}

func execMapTask(mapf func(string, string) []KeyValue, filename string, mapId int, nReduce int) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot open %v", filename)
		return err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", filename)
		return err
	}
	file.Close()
	kvAll := mapf(filename, string(content))
	kvBucket := make([][]KeyValue, nReduce)
	for _, kv := range kvAll {
		reduceId := ihash(kv.Key) % nReduce
		kvBucket[reduceId] = append(kvBucket[reduceId], kv)
	}
	for i, bu := range kvBucket {
		jsonName := fmt.Sprintf("mr-%d-%d.json", mapId, i)
		err := saveJSON(jsonName, bu)
		if err != nil {
			log.Fatalf("cannot save %v", jsonName)
			return err
		}
	}
	return nil
}

func execReduceTask(reducef func(string, []string) string, nMap int, reduceId int) error {
	kva := make([]KeyValue, 0)
	for i := 0; i < nMap; i++ {
		jsonName := fmt.Sprintf("./out/mr-%d-%d.json", i, reduceId)
		file, err := os.Open(jsonName)
		if err != nil {
			log.Fatalf("cannot open %v", jsonName)
			return err
		}
		dec := json.NewDecoder(file)
		for {
			var kv []KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			kva = append(kva, kv...)
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
		return err
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
	return nil
}

func saveJSON(filename string, kv []KeyValue) error {
	// 定义文件夹路径
	dirPath := "./out"

	// 尝试创建文件夹，包括所有必要的父文件夹
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return err
	}

	filename = fmt.Sprintf("%v/%v", dirPath, filename)
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
