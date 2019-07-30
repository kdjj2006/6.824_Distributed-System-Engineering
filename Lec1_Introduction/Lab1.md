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
  
  ***任务： 当我们在我们的机器上运行你的代码，如果通过顺序操作的测试(运行上述命令)，你将会得到本部分的所有分数。***  

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
  回顾 MapReduce 论文的第二章，你的 mapF() 和 reduceF() 要和原论文章节 2.1 有点区别。你的 mapF() 要传递文件名，文件内容，应该将内容分割为单词，并返回一个 mapreduce.KeyValue 的 GO 切片。


 TODO:后面还未翻译
   
    
