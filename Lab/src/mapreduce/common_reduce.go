package mapreduce

func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTask int, // which reduce task this is
	outFile string, // write the output here
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	//
	// doReduce manages one reduce task: it should read the intermediate
	// files for the task, sort the intermediate key/value pairs by key,
	// call the user-defined reduce function (reduceF) for each key, and
	// write reduceF's output to disk.
	//
	// You'll need to read one intermediate file from each map task;
	// reduceName(jobName, m, reduceTask) yields the file
	// name from map task m.
	//
	// Your doMap() encoded the key/value pairs in the intermediate
	// files, so you will need to decode them. If you used JSON, you can
	// read and decode by creating a decoder and repeatedly calling
	// .Decode(&kv) on it until it returns an error.
	//
	// You may find the first example in the golang sort package
	// documentation useful.
	//
	// reduceF() is the application's reduce function. You should
	// call it once per distinct key, with a slice of all the values
	// for that key. reduceF() returns the reduced value for that key.
	//
	// You should write the reduce output as JSON encoded KeyValue
	// objects to the file named outFile. We require you to use JSON
	// because that is what the merger than combines the output
	// from all the reduce tasks expects. There is nothing special about
	// JSON -- it is just the marshalling format we chose to use. Your
	// output code will look something like this:
	//
	// enc := json.NewEncoder(file)
	// for key := ... {
	// 	enc.Encode(KeyValue{key, reduceF(...)})
	// }
	// file.Close()
	//
	// Your code here (Part I).
	//
	// comments 翻译
	// doReduce 管理一个 reduce 任务：它应该为任务读入中间文件，通过 key
	// 对中间的 key/value 键值对排序，为每个 key 调用用户定义的 reduce 方法(reduceF)，
	// 并将 reduceF 的结果写入磁盘。
	//
	// 你需要从每个 map 任务读入一个中间文件，reduceName(jobName, m, reduceTask)
	// 生成来自 map 任务 m 的文件名。
	//
	// 你的 doMap() 将 key/value 进行编码并写入了中间文件，所以你需要进行解码。如果你
	// 使用 JSON，你可以通过新建一个 decoder 并且重复调用 .Decode(&kv) 直到放回一个错误，
	// 来读取并解码。
	//
	// 你可能发现 GO 的 sort 包文档中的第一个例子很有用。
	//
	// reduceF() 是本应用的 reduce 方法，你应该对每个不同的 key 调用一次，使用这个 key 对应
	// 的所有的 value 的切片。reduceF() 返回 key 已经做过 reduce 的value。
	//
	// 你应该将 reduce 输出写为 JSON 编码的 KeyValue 对象，名为 outFile 的文件。我们要求使用
	// JSON, 因为这个是合并而不是组合所有来自 reduce 期望输出。 JSON 没有什么特变的，它只是我们
	// 选择使用的编码格式。你的输出代码可以像下面这样：
	//
	// enc := json.NewEncoder(file)
	// for key := ... {
	// 	enc.Encode(KeyValue{key, reduceF(...)})
	// }
	// file.Close()
	//
	// 你的代码(Part I)
}
