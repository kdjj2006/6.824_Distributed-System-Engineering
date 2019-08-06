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