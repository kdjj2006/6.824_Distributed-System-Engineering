6.824 2018 第1课：课程介绍      
6.824 分布式系统工程  
#### 什么是分布式系统
* 多台共同协作的计算机
* 为大型网址存储，MapReduce ，P2P 共享
* 大量关键基础设施是分布式的
#### 为什么需要分布式
* 为了组织单个物理实体
* 为了通过隔离实现安全性
* 为了通过复制实现容错
* 为了通过并行CPUs/内存/磁盘/网络来实现吞吐量扩容
#### 但是
* 复杂性：多个并发部分
* 必须应对处理部分失败的问题
* 实现性能潜力问题很棘手
#### 为什么要学习这门课？
* 有趣 — 题目难，解决方法强大
* 真实系统使用 — 由大型网站的兴起驱动
* 热门的研究领域 — 有大量未解决并伴随着不断进步的领域
* 实践 - 你需要在实验中构建多个系统
### 课程相关内容
* #### [课程结构](http://pdos.csail.mit.edu/6.824)
* #### 课程工作人员
  Malte Schwarzkopf, lecturer   
  Robert Morris, lecturer   
  Deepti Raghavan, TA   
  Edward Park, TA   
  Erik Nguyen, TA       
  Anish Athalye, TA 
* #### 课程组成
    * 课程
    * 阅读
    * 两次考试
    * 实验 
    * 最终项目（可选）
    * 助教办公时间
    * 发公告和实验帮助的piazza
* 课程
    * 重大思想，论文讨论，实验
* 阅读
    * 论文研究，一些经典的，一些新的
    * 论文阐述了关键思想和重要细节
    * 许多课程集中在这些论文上
    * 上课前请阅读论文
    * 每篇论文都有一个简短的问题需要你回答
    * 你需要给我们发送一个你关于论文的疑问
    * 在课程前一天的午夜前提交问题和答案
* 考试
    * 期中考试为课堂考试
    * 期末考试在最后一周
* 实验目标
    * 对重要技术的深入理解
    * 积累分布式编程经验
    * 第一个实验从周五开始计时一周
    * 以后一周一次
* 实验
    * 实验一：MapReduce
    * 实验二：使用Raft实现复制的容错
    * 实验三：容错 key/value 存储
    * 实现四：分片 key/value 存储
* 可选的最终项目在最后，可以分成2到3人一组
    * 最终项目可代替实验4
    * 你自己想一个项目，并和我们一起搞清楚它
    * 代码，简短评论，在最后一天简短演示
* 实验成绩取决于通过了多少测试用例
    * 我们提供用例，所以你可以知道你自己做的好不好
    * 担心：如果它经常运行通过，偶尔失败，我们运行的时候可能会失败
* 调试实验会非常耗时
    * 尽早开始
    * 参与助教办公时间
    * 在 Piazza 提问

### 主要主题    
* 这是一个关于基础设施的课程，供应用程序使用  
关于隐藏应用程序分发的抽象  
* 主要抽象：
   * 存储
   * 通讯
   * 计算
一再出现的一些主题
* 主题：实现    
  * RPC，线程，并发控制
* 主题：性能
    * 目标：可扩展的吞吐量
      Nx 个服务器 -> Nx 吞吐量通过并行CPU，磁盘，网络   
      因此处理更高负载只需要购买更多机器。
    * 随着 N 不断变大，扩展越来越难
      * 负载均衡
      * 不可并行化的代码：初始化，交互
      * 来自恭喜资源的瓶颈，例如网络
    * 请注意，有些性能问题不能简单的通过扩展解决
      * 例如，减少但个用户的请求响应时间，可能需要程序员的努力而不是堆积更多机器
* 主题：容错
    * 数千台服务器，复杂的网络往往导致常有一些罢工
    * 我们希望从应用程序中隐藏这些故障
    * 我们经常希望：
        * 可靠性 -- 应用运行而忽略故障
        * 持久性 -- 应用可以在故障修复后恢复运行
    * 重要思想：复制服务器
        * 如果一台服务器奔溃了，客户端可以使用其他服务器
* 主题：一致性
    * 通用的基础架构需要明确定义的行为
        * 例如Get(k)取的值要是最近的Put(k,v)的值
    * 实现良好好的行为很难
        * 复制的服务器很难保证相同
        * 客户端可能在多步更新的中途奔溃
        * 服务器在不合时宜的时候奔溃了，例如执行完成，但是还没有返回结果
        * 网络可能让可用的服务器看着像死了一样，“闹裂”风险
    * 一致性和性能是不能同时兼顾的
        * 一致性需要通信，例如，取得最近的Put()
        * “强一致性”通常会导致系统变慢
        * 在应用中，高性能通常导致弱一致性
    * 人们已经在这方面有了许多设计方法

### 案例学习：MapReduce
* 让我们讨论下 MapReduce (MR) 作为一个案例学习，MR 是 6.824 学习的很好的案例，也是Lab 1 的焦点
* MR 总览   
  * 背景：
    * 对数 TB 的数据集进行数小时的计算，例如分析爬虫网页的图形结构
    * 仅适用于数千台计算器
    * 经常不是由分布式系统专家开发
    * 分布式会非常痛苦，例如故障处理
  * 总体目标：  
    * 非专业编程人员可以轻易区分完成    
    * 在很多服务器上以合理的效率处理数据
  * 编程人员定义 Map 和 Reduce 方法
    * 顺序性的代码，通常很容易
  * MR 在数千台机器上用巨大的输入来执行方法并且隐藏了分布式的细节
* MapReduce 抽象视图
    * 输入被分成了 M 个文件
    * 【图：map 生成 K-V 键值对的行，reduce 消费每一列】  
     Input1 -> Map -> a,1 &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;b,1&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; c,1   
     Input2 -> Map ->    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; b,1  
     Input3 -> Map -> a,1     &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;c,1  
               &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; &nbsp;&nbsp;  |  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; |       
                   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;   |&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;   -> Reduce -> c,2    
                   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; |  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; -----> Reduce -> b,2    
                    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;---------> Reduce -> a,2    
    * MR 为每个输入文件调用 Map()，产生 k2,v2 的数据集
        * 中间数据
        * 每个 Map() 调用成为一个任务
    * MR 为给定的 k2 收集所有的中间数据 v2,并将它们传给一个 Recude 调用
    * 最终输出是一个来自 Reduce() 的 <k2,v3> 键值对的数据集，存在 R 个输出文件中
* 示例：单词统计
    * 输入是有数千个文本文件
    * 
    >  Map(k, v)    
    &nbsp;&nbsp; split v into words  
    &nbsp;&nbsp; for each word w    
    &nbsp;&nbsp;&nbsp;&nbsp;  emit(w, "1")  
Reduce(k, v)  
   &nbsp;&nbsp;&nbsp;&nbsp; emit(len(v))
* MapReduce 很好的扩展：
    * N 个机器可以获得 Nx 的吞吐量  
      假设 M 和 R >= N (例如很多输入文件和 map 输出文件)    
      Map() 可以并行执行，因为它们互相没有交互  
      Reduce() 也是一样     
      唯一的交互就是在 map() 和 reduce() 之间的转换
    * 可以通过买更多机器来获得更多的吞吐量  
      而不是每个应用程序的专用高效并行化处理    
      计算机比程序便宜
* 什么可能会限制性能？  
  我们关心的原因是这是可优化的事情  
  CPU？内存？磁盘？网络？   
  在 2004 年作者受到了“网络横截面带宽”的限制
    * 注意在 Map -> Reduce 转换过程中，所有数据通过网络
    * 论文中的根交换机：100-200GB/s
    * 1800 台机器，所以 55M/秒/台机器
    * 慢，远小于磁盘(约50-100M/s在当时)或者 RAM 的速度  
 所以他们关心的是在网络间最小的数据移动(数据中心的网络今天快多了)
* 更多的细节(论文中的图1) 
    * master：给 workers 分配任务；记住中间输出的位置
    * M 个 Map 任务，R 个 Reduce 任务
    * 输入文件存在 GFS，每个输入文件有3份拷贝
    * 所有机器都运行 GFS 和 MR worker
    * 输入任务远大于 worker 数量
    * 旧任务完成后分发新任务
    * Map worker 在本地磁盘将中间 key 哈希成 R 个分区
      * 问题：有什么好的数据结构来实现这个？
    * 所有 Map 任务执行完成后才会开始调用 Reduce
    * master 告诉 Reducers 从 Map worker 抓取中间数据
    * Reduce workers 将最终输出写入 GFS (每个 Reduce 一个文件)

* reduce 是如何详细设计来降低慢网络影响的
    * Map 的输入是从 GFS 存于本地的副本读取，而不是通过网络
    * 中间文件只通过网络传输一次
        * Map worker 将文件写入本地磁盘，而不是GFS
    * 中间文件被切分成了很多份包含很多 Key 的文件
    * 问题: 为什么不将 mapper 生成的文件以流的形式传给 reducer(通过TCP)
* 如何获得好的负载均衡
    * 扩展至关重要 -- N-1个 servers 等待1个服务器完成任务很不好     
      但是很多任务可能比其他任务需要的时间更多
    * 解决方法：task 数量比 worker 多很多   
      master 给已经完成任务的 worker 分配新任务    
      所有没有一个太大以至于占据多数时间（期望）    
      所有快的 server 比慢的做更多的工作，差不多同时结束
* 关于容错  
  例如，一个服务器在执行一个 MR 任务的期间崩溃了会怎么样？
    * 故障隐藏是易用性编程的重要部分    
    问题：为什么不从一开始就重启整个任务
    * MR 只会重新执行失败的 Map() 和 Reduce()
        * MR 需要它们成为纯函数
            * 它们不在调用间保持状态
            * 它们除了 MR 的输入输出之外不会读取或写入其他文件
            * 任务之间没有隐式通信
        * 所以重新执行会产生相同的输出
    * 与其他并行编程方式相比，MR 对纯函数的要求是主要的限制，但这是 MR 易用性的关键
* worker 崩溃恢复的细节
    * Map worker 奔溃
        * master 发现 worker 无法 ping 通
        * 奔溃的 worker 的临时 Map 输出丢失，但是每个 Reduce 任务都可能需要这个输出
        * master 重新执行 worker，通过其他 GFS 的副本的输入
        * 其他 Reduce worker 可能已经读取了这个奔溃的 Map worker 的数据
        * master 不需要重新执行 Map 如果 Recude 已经抓取了所有的临时数据，虽然一个 Reduce 的奔溃会导致重新执行失败的 Map
    * Reduce worker 奔溃
        * 已经完成的任务没关系 - 已经使用副本，存储在 GFS 中
        * master 重新分配未完成的任务给其他 worker
    * Reduce 在输出的过程中奔溃了
        * GFS 已经原子重命名防止输出可见，直到完成，所以 master 在其他地方重新执行 Reduce 任务是安全的
* 其他失败/问题
    * 如果 master 将相同的 Map 任务给了两个 worker 会怎么样?        
    可能 master 错误地认为一个worker 一个不可用了  
    master 只会将其中的一个通知给 Reduce worker
    * 如果 master 将两个相同的 Map 任务分给了两个 worker 会怎么样？     
    它们都会尝试向 GFS 写入相同的输出   
    原子的 GFS 重命名防止混淆，一个完整的文件可见
    * 如果一个单独的 worker 很慢怎么办 - 一个拖后腿的？     
    可能是奇怪的硬件问题导致的      
    master 会启动最后一些任务的第二克隆任务
    * 如果一个 worker 由于损坏读入硬件或者软件写了不正确的输出怎么办？  
    太糟糕了。MR 假设 CPU 和软件会故障停止
    * 如果 master 奔溃了怎么办？    
    从 check-point 恢复，或者放弃当前任务
* 那些应用不适合使用 MapReduce
    * 不是所有文件都符合 map/shuffle/reduce 格式
    * 数据量小，因为开销大。例如不是网站后端
    * 大数据的小更新，例如对大索引加入一些文档
    * 不可预测的读(Map 和 Redude 都无法选择成输入)
    * 多次合并，例如 page-rank(可以使用多次 MR 但效率不高)
    * 更灵活的系统允许这些，但是需要更复杂的模型
* 现实中的网络公司应该如何使用 MapReduce    
    "CatBook",一个运营猫的社交网络的新公司，需要：
    1) 建立一个搜索索引，索引人们可以发现其他人的猫
    2) 分析不同猫品种的人气，来决定广告价值
    3) 检测狗，并删除它们的资料     
    可以使用 MapReduce 来完成以上所有的目标

    
    * 每天晚上对所有的资料允许大量批量任务
    1) 构建反向索引     
        map(profile text) -> (word, cat_id)     
        reduce(word, list(cat_id) -> list(word, list(cat_id))
    2) 统计资料访问量       
        map(web logs) -> (cat_id, "1")  
        reduce(cat_id, list("1")) -> list(cat_id, count)
    3) 过滤资料     
    map(profile image) -> img analysis -> (cat_id, "dog!")  
    reduce(cat_id, list("dog!")) -> list(cat_id)
### 总结
* MapReduce 一手引领了大集群计算的流行
    * 虽然不是最有效最灵活
    * 扩展良好
    * 易于编程 -- 故障和数据移动被隐藏了

    这些是实践中的很好的平衡    
    在后续的课程中会学习更高级的新技术      
    享受实现的乐趣  