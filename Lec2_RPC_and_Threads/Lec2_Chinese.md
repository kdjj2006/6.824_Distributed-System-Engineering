6.824 2018 第2课：基础：RPC 和线程

### 最常见的问题：为什么用GO语言？
* 6.824 已经使用 C++ 很多年了   
  C++ 工作的很好    
  但是学生在指针和 alloc/free 内存管理的 bug 跟踪上要花不少时间     
  而且没有让人满意的 C++ RPC 包     
* GO 对于我们来说比 C++ 好  
  对并发的良好支持(goroutines, channels, &c)    
  对 RPC 的良好支持     
  内存回收(解决问题后无用)  
  类型安全  
  线程和 GC 很吸引人
* 我们喜欢用 GO 编程，相当简单和传统    
  在本教程后，使用 https://golang.org/doc/effective_go.html     
  Russ Cox将于3月8日举行客座讲座

### 线程

   * 线程是一种有用的结构化工具   
    GO 将他们称为协程，其他的称它们为线程    
    它们可能很棘手

   * 为什么需要线程     
   它们表达并发，所以很自然出现在分布性系统中   
   I/O 并发：当等待另一个服务器回复的时候，执行下一个请求   
   多核：线程可以在多核上并行执行

   * 线程 = “执行的线程”    
   线程允许一个程序一次执行(逻辑上)很多事情     
   线程共享内存     
   每个线程包含一些每个线程状态的信息：程序计数器，寄存器，栈
   
   * 一个程序内有多少线程   
        * 有时候由结构驱动，例如每个客户端一个线程，后台任务一个线程   
        * 有时候由对多核并行性的期望驱动   
        每个核心一个活动线程    
        GO 运行时自动在活动的核心上调度可执行的协程    
        * 有时候由 I/O 并发的期望驱动   
         延迟和容量决定数量     
         一直增长直到吞吐无法增加   
        * GO 线程很轻量级     
        几百几千很常见，但几百万可能不行    
        新建线程比一个方法调用要昂贵    
* 线程的挑战
    * 共享数据  
        一个线程读取另一个线程正在修改的数据？  
        例如，两个线程执行 count = count +1, 这就是一个“竞态”，而且往往是一个bug    
        -> 使用 Mutexes (或者其他同步)  
        -> 或者避免共享数据 
    * 线程的协作    
        如何等待所有 Map 线程结束？     
        -> 使用 GO channel(通道) 或者 WaitGroup
    * 并发的粒度    
        粗粒度 -> 简单，但是并发/并行量小   
        细粒度 -> 更多并发，更多的竞态和死锁
### 爬虫
* 什么是爬虫    
目标是抓取所有网页，例如提供给索引器        
网页组成一张图  
每个网页有多个链接  
图有闭环
* 爬虫的挑战
    * 安排 I/O 并发     
    同时抓取很多 URL    
    增加每秒抓取的 URL 数量     
    由于网络延迟远远超过网络容量的限制      
    * 每个 URL 仅抓取一次   
    防止浪费网络带宽    
    对远程服务器有好处  
    => 需要记住哪些 URL 已经访问过了
    * 知道什么时候完成
* 爬虫的解决方案[ crawler.go 链接 schedule 页面 ]
    * 串行化爬虫    
fetched map 用来避免重复，退出闭环    
是简单的 map，通过引用传递给递归调用    
但是：一次只抓取一个页面
    * 并发互斥型爬虫
        * 为每个页面抓取新建线程    
        很多并发抓取，更高的抓取率
        * 线程共享 fetched map
        * 为什么使用 Mutex(== lock)
            * 没有锁    
            两个网页含有同一个 URL 链接     
            两个线程同时抓取这2个页面   
            T1 检查 fetched[url], T2 检查 fetched[url]   
            两个线程都发现 url 没有被抓取     
            两个线程都抓取，这就错了
            * 同时读和写(或者写和写)是“竞态”    
            经常预示着一个 bug  
            bug 可能只有线程交互中才会不幸发生
            * 如果我提出 Lock()/Unlock() 的调用会怎样？     
            go run crawler.go   
            go run -race crawler.go 
            * 锁操作导致 check 和 update 操作原子化
        * 它是如何决定已经完成的    
        sync.WaitGroup  
        隐式等待子调用完成递归提取
    * 并发通道型爬虫
        * GO 通道
            * 一个通道是一个对象，可以有很多    
                ch := make(chan int)
            * 一个通道允许一个线程发送一个对象给另外一个线程    
                ch <- x     
                发送方一直等到某个 goroutine(协程) 收到消息
            *  y := <- ch
                 for y := range ch  
               接收方等到某个 goroutine(协程) 发送消息
            * 你可以使用一个通道来通信和同步    
            多个线程可以在一个通道上收发    
            记住：发送方阻塞直到接收方收到  
            可能在发送时持有锁很危险    
        * 并发通道 master()
            * master() 新建一个 worker 协程来抓取每个页面
            * worker() 在通道中发送 URL 地址    
                多个 worker 在一个通道中发送
            * master() 从通道读取 URL 地址
        * 无须将 fetched map 加锁，因为它没有共享
        * 有没有某些共享数据？
            * 通道
            * 在通道中发送的切片和字符串
            * master() 发送给 worker() 的参数
    * 什么时候使用共享，锁，与通道
        * 大多数问题可以用其中的一种方式解决
        * 最有意义的事情取决于程序员的想法
            * 状态 -- 共享和锁
            * 通信 -- 通道
            * 事件等待 -- 通道
        * 使用 GO 的竞态探测器
            * https://golang.org/doc/articles/race_detector.html
            * go test -race

### Remote Procedure Call (RPC)
* 分布式系统机制的一个关键块；所有的实验都使用 RPC  
    目标：易编程的 client/server 通信

* RPC 消息图    
    Client         &nbsp; &nbsp;  &nbsp; &nbsp;   &nbsp; &nbsp;  &nbsp; &nbsp;   Server       
    request--->     
      &nbsp; &nbsp;  &nbsp; &nbsp;   &nbsp; &nbsp;  &nbsp; &nbsp;    <---response 
* RPC 尝试模仿本地 fn 调用      
    > Client:       
      &nbsp; &nbsp;  z = fn(x, y)    
     Server:    
      &nbsp; &nbsp;  fn(x, y) {  
      &nbsp; &nbsp;&nbsp; &nbsp;   compute   
      &nbsp; &nbsp;&nbsp; &nbsp;   return z  
      &nbsp; &nbsp;  }   

    实际中没这简单
* 软件结构  
     client app  &nbsp; &nbsp;&nbsp;&nbsp; &nbsp;&nbsp;       handlers    
   &nbsp; stubs   &nbsp; &nbsp;&nbsp;&nbsp; &nbsp;&nbsp;&nbsp; &nbsp;&nbsp;        dispatcher  
   &nbsp;RPC lib  &nbsp; &nbsp;&nbsp;&nbsp; &nbsp;&nbsp;&nbsp; &nbsp;&nbsp;         RPC lib    
   &nbsp;  net  &nbsp; &nbsp;&nbsp;------------&nbsp; &nbsp;net

* GO 例子：kv.go 链接至 schedule 页面   
   一个 key/value 存储的玩具服务器 -- Put(key,value), Get(key)->value       
   使用 GO 的 RPC 库    
   * 通用：   
       你需要为每个 RPC 类型定义参数和返回类型
   * 客户端：   
        connect() 的 Dial() 建立一个和服务器的 TCP 连接     
        Call () 让 RPC 库执行调用   
       &nbsp;&nbsp;&nbsp; &nbsp;你定义了服务器的功能名字，参数，放置回复的地方  
       &nbsp;&nbsp;&nbsp;&nbsp; 库组织参数，发送请求，等等，解码回复    
       &nbsp;&nbsp;&nbsp;&nbsp; Call() 的返回值代表它是否得到了回复     
       &nbsp;&nbsp;&nbsp;&nbsp; 通常你也需要 reply.err 来表明服务层的失败   
   *  服务器：    
        GO 需要你定义一个含有方法的对象作为 RPC 的 处理器   
        然后用 RPC 库注册这个对象   
        接收 TCP 链接，把它们送给 RCP 库    
        RCP 库  
         &nbsp; &nbsp;&nbsp;   读取每个请求    
        &nbsp; &nbsp;&nbsp;    为请求新建一个协程  
        &nbsp; &nbsp;&nbsp;    解码请求    
        &nbsp; &nbsp;&nbsp;    调用命名方法(调度)  
        &nbsp; &nbsp;&nbsp;    组织    
            &nbsp; &nbsp;&nbsp;&nbsp;将应答写入 TCP 连接  
        服务器的 Get() 和 Put() 处理器      
          &nbsp; &nbsp;&nbsp;  必须加锁，因为 RPC 库为每个请求建立协程     
         &nbsp; &nbsp;&nbsp;   读取参数，修改应答
* 一些细节  
    * 绑定：客户端如何知道和谁通信  
            在 GO 的 PRC 里，服务器的名字/端口是 Dial 的参数    
            大型系统有一些命名或者配置服务器
    * 编组：将数据格式化为数据包    
            GO 的 PRC 库可以传送字符串，数组，对象，map, &c     
            GO 通过拷贝传送指针(服务器不能直接使用客户端指针)   
            不能传送通道或者 function
* RPC 的问题：如何处理故障？    
        例如，丢包，网络故障，服务器缓慢，服务器崩溃
* 客户端 RCP 库的失败是什么样的     
     *  客户端不能从服务器收到回复  
      * 客户端不知道服务器是否收到请求  
         * 可能没有收到请求 
         * 可能执行了，在回复前忽然奔溃了  
         * 可能执行了，但是网络在回复返回前损坏了  
* 最简单的故障处理方案：最大的努力  
    *  Call() 等回复等一会     
    * 如果没有回复，再发送一次请求    
    * 重复数次    
    * 放弃，返回错误
* 一个特别的失败案例：  
        客户端执行  
       &nbsp; &nbsp;&nbsp;   Put("k", 10);  
       &nbsp; &nbsp;&nbsp;   Put("k", 20);  
        都成功了    
        如果执行 Get("k") 得到什么值？
* 好一点的操作：最多一次    
    * 思想：服务器检测到重复请求，返回以前的返回信息而不是重新执行  
    * 问题？如何检测重复请求    
      客户端在每个请求中包含唯一的ID(XID),重新发送时使用同样的 XID  
    * 服务器：  
      >if seen[xid]:    
       &nbsp; &nbsp;&nbsp; r = old[xid]  
        else    
      &nbsp; &nbsp;&nbsp;  r = handler() 
      old[xid] = r  
       &nbsp; &nbsp;&nbsp; seen[xid] = true
* 一些“最多一次”的复杂性    
    * 在实验3里会遇到   
    * 如何保证 XID 唯一性   
        * 大的随机数？     
        * 将唯一的客户ID(IP 地址)和序列号结合起来？    
    * 服务器必须丢弃旧 RPC 的信息
        * 何时丢失是安全的？
        * 思想：
            * 每个客户端有一个唯一ID(可能是一个很大的随机数)
            * 每个客户端 RPC 序列号
            * 客户端每个 PRC 包含“看见所有 <= X 的回复”
            * 更像 TCP 序列号和 ack
        * 或者一次只允许客户端使用一个未完成的RPC   
          第 seq+1 个请求到达允许服务器丢弃所有 <= seq 的请求
    * 如何在原请求还在执行的时候处理重复请求？
        * 服务器还不知道返回值
        * 思想：每个执行的 RPC 设置"pending" 标记；等待或者忽略
* 如果“最多一次”的服务器崩溃了或者重启会怎么样？
    * 如果“最多一次”的重复信息放在内存中，服务器会再重启后忘记这些信息，并接受重复的请求
    * 可能应该将这些重复的信息写入磁盘
    * 可能复制服务器也应该复制重复信息
* GO RPC 是“最多一次”的简单形式
    * 打开 TCP 连接
    * 将请求写入 TCP 连接
    * GO RPC 永远不会重发请求，所以服务器不会收到重复请求
    * GO RPC 代码在如果没有收到回复时会返回一个错误
        * 可能在超时后(TCP)
        * 可能服务器没有收到请求
        * 可能服务器处理了请求，但是服务器/网络在返回前失败了
* “仅仅一次”怎么样？
    * 无限次重试以及重复检测和容错服务
    * 实验 3