6.824 Lab 1: MapReduce  
***

### 介绍
* 在本实验中你会构建一个 MapReduce 库作为 GO 编程入门学习，并构建一个分布式容错系统。第一部分，你需要写一个简单的 MapReduce 程序。第二部分，你要写一个 Master 用来为 MapReduce workers 分发任务，解决 workers 的故障。这个库的接口和容错的解决方法类似于 MapReduce 论文中介绍的那样。
### 合作规则
* 你必须自己完成6.824课程的所有代码，除了那些我们作为课程一部分分发的代码。你不允许看别人的解决方法，也不能看前几年的。你可以与其他同学讨论，但是不能拷贝或者看其他人的代码。这么规定的原因是我们相信你们可以从亲身设计实现实验解决方案的过程中学到最多。
* 请不要将你的代码推送给现在或者未来6.824的学生，Github 的仓库默认是公有的，所以不要将代码放 Github 上除非是私有仓库。你可能会发现 MIT's GitHub 很好用，但是请确保所建仓库为私有。
### 软件
* 我们使用 [GO](https://golang.org/) 实现所有的实验。GO 的网站已经包含了许多教程，你们可能想去看看。我们使用 GO 1.9来为你们的实验打分，所以你也该用1.9，因为我们不知道其他版本有没有问题。
* 这些实验室设计用于在 x86 或 x86_64 架构的 Athena Linux 机器上运行； uname -a应该提到i386 GNU / Linux或i686 GNU / Linux或x86_64 GNU / Linux。 你可以使用 ssh athena.dialup.mit.edu登录公共 Athena主机。你可能会很幸运地发现实验室可以在其他环境中工作，例如在某些 Linux 或 OSX 的笔记本电脑。
* 我们提供了 MapReduce 的部分实现，支持分布式或者非分布式(有点无聊)的操作。你需要使用 [git](https://git-scm.com/) (一个版本控制系统)来抓取实验的初始版本。可以通过查看 [Pro Git book](https://git-scm.com/book/en/v2) 或者 [git user's manual](https://mirrors.edge.kernel.org/pub/software/scm/git/docs/user-manual.html) 来学习git，又或者你已经很熟悉其他版本控制系统，你也许会发现 [CS-oriented overview of git](http://eagain.net/articles/git-for-computer-scientists/)很有用。
* 这些 Athena 命令会让你获得 git 和 GO。
    > athena$ add git   
    > athena$ setup ggo_v1.9

    本课程的 git 仓库地址是：git://g.csail.mit.edu/6.824-golabs-2018，你可以通过运行如下命令来 clone 课程仓库到你本地的目录中。
    > $ git clone git://g.csail.mit.edu/6.824-golabs-2018 6.824     
      $ cd 6.824     
      $ ls    
    Makefile src    
    
    Git 可以允许我们跟踪代码的变化轨迹。例如，你想保存下工作进度，可以通过下面命令提交你的修改：
    > $ git commit -am 'partial solution to lab 1'

* 我们提供的 Map/Reduce 实现提供了两种模式的操作，顺序和分布式。顺序操作中，map 和 reduce 一次只能执行一个，第一个 map 任务执行完才能执行第二个，然后第三个。当所有 map 任务完成了，才开始执行第一个 reduce 任务。这种模式，虽然慢点，但是利于调试。分布式模式就是运行很多线程，这些线程并发的执行 map 任务，然后是 reduce 任务。这种模式运行更快，但是也更难实现，更难调试。
### 前言：熟悉源代码
* mapreduce 包提供了一个简单的 Map/Reduce 库（在 mapreduce 目录下）。应用程序一般应该调用 Distributed() 方法(在 master.go 文件中)来启动一个任务，但也可以调用 Sequential() 方法(同样在 master.go 文件)来获得顺序的执行从而用来调试程序。
* 代码执行任务如下：
    * 1 应用提供一些输入文件，一个 map 方法，一个 reduce 方法，以及 reduce 任务的数量(nReduce)
    * 2 一个 master 基于这些信息创建。它启动一个 PRC 服务器(详见 master_rpc.go )，并且等待 workers 来进行注册( 使用 master.go 文件中定义的 RPC 调用 Register())。伴随着任务变为可用(在第4,第5步)，schedule() 方法(schedule.go)决定如何分发这些任务给 workers，以及如何处理 worker 的错误。
    * 3 master 把每个输入文件当成一个 map 任务，然后至少为每个 map 任务调用一次 doMap()方法(common_map.go)。任务要么直接触发(当使用 Sequential()方法)，要么通过给 worker 发出 DoTask() 的 RPC 调用(worker.go)。每个 doMap() 方法读取正确的文件，并对文件内容使用 map 方法，将 key/value 键值对结果写入到 nReduce (注：步骤1中定义) 个中间文件中。doMap()方法对每个 key 做哈希来决定中间文件，reduce 任务也会使用这些 key。当所有 map 任务完成后会生成 nMap * nReduce 个文件。每个文件名会包含 map 任务编号和 reduce 任务编号的前缀(注：看下面列的文件名是不是应该是后缀？)。如果有2个 map 任务和3个 reduce 任务，map 任务会创建6个中间文件：

         > mrtmp.xxx-0-0         
         > mrtmp.xxx-0-1     
         > mrtmp.xxx-0-2     
         > mrtmp.xxx-1-0     
         > mrtmp.xxx-1-1     
         > mrtmp.xxx-1-2 
    
         每个 worker 必须有权限读取另一个 worker 写的文件作为自己的输入文件。真正的部署使用分布式存储系统例如 GFS 来实现这些读取即使 workers 在不同的机器上运行。在本次实验中，所以的 worker 在同一机器上运行，并使用本地文件系统。
    * 4 master 接下来至少为每个 reduce 任务运行一次 doReduce()(common_reduce.go) 。就跟 doMap()方法一样，任务要么直接运行，要么通过 worker 触发。reduce 任务 r 的doReduce() 方法从 map 任务中收集第 r 个中间文件，然后调用 reduce 方法来处理每一个在文件中出现的 key。reduce 任务生成 nReduce 个结果文件。
    * 5 master 调用 mr.merge()(master_splitmerge.go)来合并上一步生成的所有 nReduce 个文件使它们输出为单个文件。

    > 注意：在本课下面的练习中，你需要分别实现/改写 common_map.go,  common_reduce.go 和 schedule.go 文件中的doMap, doReduce 以及 schedule 方法。另外也需要实现 ../main/wc.go 中的 map 和 reduce 方法。
* 你不需要修改其他任何文件，但是认真阅读它们有助于理解其他代码是如何嵌入系统的整个架构中的。

### Part I： Map/Reduce 输入和输出
* 我们给出的 Map/Reduce 实现缺少了一些部分。在你开始写你的第一个 Map/Reduce 功能对之前，你需要修复顺序操作的实现。特别说明的是，给你们的代码缺少了两个关键部分：分割 map 任务输出的功能和为一个 reduce 任务收集所有输入的功能。这些任务分别是由 common_map.go 中的 doMap()，common_reduce.go 中的 doReduce()执行。这些文件中的注释可以指引你正确的方向。  
为了帮你决定你是否正确地实现了 doMap() 和 doReduce()，我们提供了一个 GO 测试套件可以检测你实现的正确性。这些测试都在 test_test.go 文件中实现了。可以通过运行下面的命令来测试你关于顺序操作的修复操作：
  > $ cd 6.824  
   $ export "GOPATH=$PWD"  # go needs $GOPATH to be set to the project's working directory   
   $ cd "$GOPATH/src/mapreduce"  
   $ go test -run Sequential   
   ok  	mapreduce	2.694s
  
  ***任务： 当我们在我们的机器上运行你的代码，如果通过顺序操作的测试(运行上述命令)，你将会得到本部分的满分。***  

  如果测试输出没有显示OK，那你的实现就有bug。在 common.go 中设置 debugEnabled = true，并且在上面的测试命令中加入 -v 命令，会给出更纤细的输出。你会得到更多输出根据如下命令行：
  > $ env "GOPATH=$PWD/../../" go test -v -run Sequential   
    === RUN   TestSequentialSingle  
    master: Starting Map/Reduce task test   
    Merge: read mrtmp.test-res-0    
    master: Map/Reduce task completed   
    --- PASS: TestSequentialSingle (1.34s)  
    === RUN   TestSequentialMany  
    master: Starting Map/Reduce task test   
    Merge: read mrtmp.test-res-0  
    Merge: read mrtmp.test-res-1  
    Merge: read mrtmp.test-res-2  
    master: Map/Reduce task completed   
    --- PASS: TestSequentialMany (1.33s)  
    PASS  
    ok  	mapreduce	2.672s  

### Part II：单 worker 单词统计
* 现在你要实现单词统计 -- 一个简单的 Map/Reduce 例子。查看 main/wc.go 文件，你会发现空方法，mapF() 和 reduceF()。你的工作就是在 wc.go 文件中加入报告它的输入中每个单词的出现次数的代码。就像 unicode.IsLetter 中确定的一样，任何连续的字符序列算一个单词。   
在 ~/6.824/src/main 目录下有一些类似于 pg-*.txt 格式的输入文件，都是从 [Project Gutenberg](https://www.gutenberg.org/ebooks/search/%3Fsort_order%3Ddownloads) 下载的。下面是如何使用这些输入文件运行 wc ：
  > $ cd 6.824    
    $ export "GOPATH=$PWD"    
    $ cd "$GOPATH/src/main"   
    $ go run wc.go master sequential pg-*.txt   
      \# command-line-arguments    
    ./wc.go:14: missing return at end of function   
    ./wc.go:21: missing return at end of function   

  编译会失败因为 mapF() 和 reduceF()未完成。  
  回顾 MapReduce 论文的第二章，你的 mapF() 和 reduceF() 要和原论文章节 2.1 有点区别。你的 mapF() 要传递文件名，文件内容，应该将内容分割为单词，并返回一个 mapreduce.KeyValue 的 GO 切片。当你为mapF选择输出 key 和 value的时候，对于单词统计来说只有将单词设置为 key才有意义。你的 reduceF() 对每个 key 都将被调用一次，并且 mapF() 方法会针对这个 key 生成所有 values 的切片。它必须返回一个包含 key 的出现次数的字符串。
  > 提示：
  * 一个很好的阅读 Go string 的地方是[Go Blog on strings](http://blog.golang.org/strings)
  * 你可以使用 [strings.FieldsFunc](http://golang.org/pkg/strings/#FieldsFunc) 将 string 拆分为 components
  * [strconv 包](http://golang.org/pkg/strconv/) 可以用来将string 转换为 int等  

  你可以通过以下方法来测试你的解决方法：
    >  $ cd "$GOPATH/src/main"   
    $ time go run wc.go master sequential pg-*.txt    
    master: Starting Map/Reduce task wcseq    
    Merge: read mrtmp.wcseq-res-0   
    Merge: read mrtmp.wcseq-res-1   
    Merge: read mrtmp.wcseq-res-2   
    master: Map/Reduce task completed   
    2.59user 1.08system 0:02.81elapsed  

  输出在"mrtmp.wcseq"文件中，如果下面的命令生成下面显示的输出，那你的实现就是正确的:
  > $ sort -n -k2 mrtmp.wcseq | tail -10    
        that: 7871    
        it: 7987    
        in: 8415    
        was: 8578   
        a: 13382    
        of: 13536   
        I: 14296    
        to: 16079   
        and: 23612    
        the: 29748    

  你也可以删除输出文件和所有的中间文件：
    > $ rm mrtmp.*
  
  为了让你们的测试更加简单，运行下面的命令，它会报告你的方法正确与否：
    > $ bash ./test-wc.sh

  ***任务：当我们在我们机器上运行你的软件，如果你的 Map/Reduce 单词统计输出跟顺序方式执行的正确输出一致，你就得到本部分满分。***

### Part III： 分布式的 MapReduce 任务
  * 现在的实现一次只执行一个 map 和 reduce 任务。Map/Reduce 一个最大的卖点就是它可以自动并行化普通的串行代码而开发人员无需额外工作。在实验的本部分，你需要完成 MapReduce 的一个版本，可以分割在多个处理器上运行的一系列 worker 线程。虽然不像真实部署的 Map/Reduce 运行在多台机器上，你的实现要用 RPC 来模拟分布式计算。   
  mapreduce/master.go 中的代码实现了管理 MapReduce job 的大部分工作。我们也在 mapreduce/worker.go 代码中提供了一个 worker 线程的完整代码，还有一些 mapreduce/common_rpc.go 中处理 RPC 的代码。  
  你的工作是实现 mapreduce/schedule.go 中的 schedule() 方法。master 在一个 MapReduce job 中调用 schedule() 方法两次，一次是为 Map 调用，一次为 Reduce。schedule() 的工作是为可用的 worker 分配任务。通常任务数比线程数多，所以 schedule() 必须给每个 worker 一个任务序号，一次一个。schedule() 要等到所以任务结束之后才能返回。   
  schedule() 通过读取 registerChan 参数来了解 worker 的工作集。channel 为每个 worker 生成一个字符串，其中包括 worker 的 RPC 地址。有些 worker 可能在 schedule() 调用前已经存在，还有的在 schedule() 运行的时候才开始，这些所有的都存在于 registerChan 中。schedule() 要使用所有的 worker，包括那些它启动之后才出现的。  
  schedule() 给 worker 发送 Worker.DoTask 的 PRC 调用来通知一个 worker 去执行任务。RPC调用的参数都在 mapreduce/common_rpc.go 文件中的 DoTaskArgs 参数定义了。File 元素只被 Map 任务使用，并读取文件名，schedule() 可以在 mapFiles 中找到所有的文件名。  
  使用 mapreduce/common_rpc.go 中的 call() 来给 worker 发送一个 RPC。第一个参数是从 registerChan 读到的 worker 的地址。第二个参数应该是 "Worker.DoTask"。第三个是 DoTaskArgs 的结构体，最后一个是 nil。 
  在 Part III 中你的解答只需要修改 schedule.go。如果你在调试过程中修改了其他文件，请在提交前恢复到初试状态并测试。  
  使用 go test -run TestParallel 来测试你的解答。会执行两个测试，TestParallelBasic 和 TestParallelCheck，后者会验证你的调度程序是否让任务并行执行。   

     ***任务：当我们在我们机器上运行你程序的时候，如果通过了 TestParallelBasic 和 TestParallelCheck，你就能得到本部分的满分***

    > 提示：    
    * [RPC package](https://golang.org/pkg/net/rpc/) 记录了 Go RPC 包。  
    * schedule() 应该并行地给 worker 发送 RPC，这样 worker 就可以并发执行任务。看看 [Go 并发](https://golang.org/doc/effective_go.html#concurrency)，你会发现 go 的声明对此很有用。   
    * schedule() 必须等 worker 完成当前任务之后才能给他分配另一个任务。你可能发现 GO 的 channel 很有用。     
    * 你可能会发现[sync.WaitGroup](https://golang.org/pkg/sync/#WaitGroup)很有用。   
    * 跟踪 bug 的最简单的方法就是增加标准输出(也许调用 common.go 中的 debug()方法)，go test -run TestParallel > out 可以将输出收集到一个文件中，然后想想输出是否符合你关于代码如何运行的预期。最后一步最重要。   
    * 用测试运行 GO 的 [Race Detector](https://golang.org/doc/articles/race_detector.html)：go test -race -run TestParallel > out 来检查你的代码是否存在竞态条件。
    
    ***请注意：我们给出的代码在将 workers 作为单个 UNIX 进程中的线程运行，并且可以在单个机器上利用多个内核。为了在通过网络进行通信的多台机器上运行 workers，需要进行一些修改。RPC必须使用 TCP 而不是 UNIX 域套接字；需要有一种方法来启动所有机器上的 worker 进程；并且所有机器都必须通过某种网络文件系统共享存储。***

### Part IV： 处理 worker 故障
* 在这个部分你要让 master 处理 worker 的错误。MapReduce 让这个相对容易，因为 worker 没有持久化状态。如果一个 worker 在处理 master 的 RPC 请求时失败了，master 的 call() 会因为超时而返回 false。在这种情况下，master 应该将分配给这个失败了的 worker 的任务给另一个 worker。    
RPC 错误并不一定意味着 worker 没有执行任务，worker 可能已经执行了任务但是回复丢失了，或者 worker 还在执行只不过 master 的 PRC 超时了。因此，两个 worker 收到同一个任务，计算并生成输出是可能发生的。需要对 map 和 reduce 执行两次才能对给定的输入生成相同的输出( map 和 reduce 功能是功能性的)，因此如果后续处理有时会读取一个输出而有时读取另一个，则不会出现不一致。另外，MapReduce 框架保证了 map 和 reduce 方法的输出是原子的，输出要么没有，要么就是 map 或者 reduce 操作的完整输出(本实验代码实际上并没有实现这一点，而是只在任务结束时使 worker 失败，因此没有任务的并发执行)。    
***注意：你不需要去处理 master 的故障。使 master 容错更复杂因为它是有状态的，必须恢复状态才能在 master 故障后恢复运行。后续的很多实验都致力于这一挑战。***  

  你的实现必须通过 test_test.go 中另外两个测试。第一个测试案例测试单个 worker 的故障处理，第二个测试多个 worker 的失败处理。这些测试用例定期启动新的 worker, master 可以使用它们去推进进度，但是这些 worker 会在执行一些任务后失败。用下面命令执行测试案例：
  > $ go test -run Failure  

  ***任务：当我们在我们机器上运行你们程序的时候，如果能通过上面命令的关于 worker 故障的测试，那你可以得到本部分的满分***  

  在 Part IV 中你的解答只需要修改 schedule.go。如果你在调试过程中修改了其他文件，请在提交前恢复到初试状态并测试。
   
### Part V： 反向索引生成(可选，不计入成绩)
* ***挑战***
 在本可选不计入成绩的练习中，你要构建 Map 和 Reduce 方法来实现反向索引生成。  
 反向索引在计算机科学中广泛应用，在文档搜索中特别有用。广义来说，反向索引是从有关基础数据的事实到该数据的原始位置的映射。举例来说，在搜索的上下文中，它可能是从关键字到包含这些单词的文档的映射。 
 我们已经建立了跟你建立的 wc.go 类似的第二个库 main/ii.go。你应该修改 main/ii.go 文件中的 mapF 和 reduceF 来让它们一起生成反向索引。运行 ii.go 应该生成一系列的元素，一个行，就像下面这样的格式：
  > $ go run ii.go master sequential pg-*.txt   
    $ head -n5 mrtmp.iiseq    
    A: 8 pg-being_ernest.txt,pg-dorian_gray.txt,pg-frankenstein.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-metamorphosis.txt,pg-sherlock_holmes.txt,pg-tom_sawyer.txt    
    ABOUT: 1 pg-tom_sawyer.txt    
    ACT: 1 pg-being_ernest.txt    
    ACTRESS: 1 pg-dorian_gray.txt   
    ACTUAL: 8 pg-being_ernest.txt,pg-dorian_gray.txt,pg-frankenstein.txt,pg-grimm.txt,    pg-huckleberry_finn.txt,pg-metamorphosis.txt,pg-sherlock_holmes.txt,pg-tom_sawyer.txt
    
  如果上面的清单还不够清晰，格式如下：
  > word: #documents documents,sorted,and,separated,by,commas   

  你可以运行 bash ./test-ii.sh 看看你的方法是否正确：
  > $ LC_ALL=C sort -k1,1 mrtmp.iiseq | sort -snk2,2 | grep -v '16' | tail -10    
      www: 8 pg-being_ernest.txt,pg-dorian_gray.txt,pg-frankenstein.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-metamorphosis.txt,pg-sherlock_holmes.txt,pg-tom_sawyer.txt    
      year: 8 pg-being_ernest.txt,pg-dorian_gray.txt,pg-frankenstein.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-metamorphosis.txt,pg-sherlock_holmes.txt,pg-tom_sawyer.txt   
      years: 8 pg-being_ernest.txt,pg-dorian_gray.txt,pg-frankenstein.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-metamorphosis.txt,pg-sherlock_holmes.txt,pg-tom_sawyer.txt    
      yesterday: 8 pg-being_ernest.txt,pg-dorian_gray.txt,pg-frankenstein.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-metamorphosis.txt,pg-sherlock_holmes.txt,pg-tom_sawyer.txt    
      yet: 8 pg-being_ernest.txt,pg-dorian_gray.txt,pg-frankenstein.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-metamorphosis.txt,pg-sherlock_holmes.txt,pg-tom_sawyer.txt    
      you: 8 pg-being_ernest.txt,pg-dorian_gray.txt,pg-frankenstein.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-metamorphosis.txt,pg-sherlock_holmes.txt,pg-tom_sawyer.txt    
      young: 8 pg-being_ernest.txt,pg-dorian_gray.txt,pg-frankenstein.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-metamorphosis.txt,pg-sherlock_holmes.txt,pg-tom_sawyer.txt    
      your: 8 pg-being_ernest.txt,pg-dorian_gray.txt,pg-frankenstein.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-metamorphosis.txt,pg-sherlock_holmes.txt,pg-tom_sawyer.txt   
      yourself: 8 pg-being_ernest.txt,pg-dorian_gray.txt,pg-frankenstein.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-metamorphosis.txt,pg-sherlock_holmes.txt,pg-tom_sawyer.txt   
      zip: 8 pg-being_ernest.txt,pg-dorian_gray.txt,pg-frankenstein.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-metamorphosis.txt,pg-sherlock_holmes.txt,pg-tom_sawyer.txt  

### 运行所有测试
* 你可以通过运行 src/main/test-mr.sh 来运行所有测试。如果你的解法正确，输出应该类似如下：
  > $ bash ./test-mr.sh   
  ==> Part I    
  ok  	mapreduce	2.053s    
  ==> Part II   
  Passed test   
  ==> Part III    
  ok  	mapreduce	1.851s    
  ==> Part IV   
  ok  	mapreduce	10.650s   
  ==> Part V (inverted index)   
  Passed test   
### 提交代码
 * ***重要：在提交前，请最后运行一次完整测试***
   > $ bash ./test-mr.sh

    通过课程提交网站提交您的代码，地址是：  
    https://6824.scripts.mit.edu/2018/handin.py/

    你可以使用MIT证书或通过电子邮件请求API密钥首次登录。登录后会显示你的API密钥（XXX），可用于从控制台上传 lab1，如下所示：
    > $ cd "$GOPATH"    
      $ echo XXX > api.key    
      $ make lab1   
    
    ***重要：检查提交网站，确保已经提交了本实验***  
    **提示：你可能提交多次代码，我们会使用最后一次提交的时间戳来计算是否按时提交**