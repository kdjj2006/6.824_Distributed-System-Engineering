2018 Lecture 7: Raft (3) -- Snapshots, Linearizability, Duplicate Detection

* 本次课程
    * Raft 快照
    * 线性化
    * 重复 RPC
    * 更快的 get
### Raft 日志压缩和快照(实验 3B)
* 注：本部分与 Lec 6 的"日志压缩和快照"几乎一模一样，故不翻译，可以参考 Lec 6 的“日志压缩和快照”
### 线性化
* 我们需要为实验3定义“正确”
    * 客户应该如何期待 Put 和 Get 的行为？
    * 通常称为一致性协议
    * 帮助我们推断如何正确处理复杂情况
        * 例如并发，副本，故障，RPC重复，领导者变更，优化
    * 我们会在6.824看到很多一致性的定义
        * 例如，Spinnaker 的时间线一致性
* “线性化”是最常见和直观的规范化单个服务器的行为的定义
* 线性化定义：
    * 执行历史是线性化的，如果：
        * 可以找到所有操作的总顺序
        * 与实时匹配(对于非重叠的操作)，并且
        * 其中每个读取都按顺序查看其前面的写入值
* 历史是客户端操作的记录，每个都有参数，返回值，开始时间，完成时间
* 例子1:
    * > -Wx1-| &nbsp;&nbsp;|-Wx2-|    
       &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;   |---Rx2---|     
     &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; |-Rx1-|   
    * "Wx1"代表“将值1写入记录x”
    * "Rx1"代码“对记录x的一次读取得出值1”
    * 顺序：Wx1 Rx1 Wx2 Rx2
        * 顺序服从值约束(W -> R)
        * 顺序服从实时约束(Wx1 -> Wx2)
        * 所以历史是线性的
* 例子2：
    * > |-Wx1-|&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; |-Wx2-|     
         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;   |--Rx2--|   
         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;     |-Rx1-|
    * Wx2 然后 Rx2 (值)，Rx2 然后 Rx1 (时间)，Rx1 然后 Wx2(值)。但是这是个环 -- 因此它不能转换为线性顺序，所以不是线性的
* 例子3：
    * > |--Wx0--| &nbsp;&nbsp; |--Wx1--|    
         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;   |--Wx2--|   
        &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|-Rx2-| &nbsp;&nbsp;|-Rx1-|
    * 顺序：Wx0 Wx2 Rx2 Wx1 Rx1
    * 所以这个是线性的
    * 注意服务可以选择并发写入的顺序
        * 例如，Raft 将并发操作放在日志里
* 例子4：
    * > |--Wx0--| &nbsp;&nbsp;&nbsp;&nbsp; |--Wx1--|    
        &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;    |--Wx2--|   
         C1: &nbsp;&nbsp;&nbsp;&nbsp; &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;  |-Rx2-| &nbsp;&nbsp;&nbsp;&nbsp;|-Rx1-|     
         C2: &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;   |-Rx1-| &nbsp;&nbsp;&nbsp;&nbsp;|-Rx2-| 
    * 我们必须能够将所有操作整合到一个单独的顺序中
        * 可能是：Wx2 C1:Rx2 Wx1 C1:Rx1 C2:Rx1
            * 但是 C2:Rx2 该放在哪里呢？
                * 必须在 C1:Rx1 之后马上到来
                * 但是然后它应该有读值为1
            * 没有什么顺序可行
                * C1 的读需要 Wx2 在 Wx1之前
                * C2 的读需要 Wx1 在 Wx2之前
                * 这是一个环，所有无序
            * 不是线性化的
    * 所以：所有客户端必须以相同的顺序查看并发写入
* 例子5：
    * 忽略最近的写入是不可线性化的
    * > Wx1  &nbsp;&nbsp;&nbsp;&nbsp;   &nbsp;&nbsp;&nbsp;&nbsp; Rx1        
       &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; Wx2
    * 这排除了脑裂，也忘记了已提交的写入
* 你可能会发现下面的页面很有用：    
https://www.anishathalye.com/2017/06/04/testing-distributed-systems-for-linearizability/

### 重复 RPC 检测(实验 3)
* 如果 Put 或者 Get RPC 超时了，客户端应该怎么做？
    * 也就是 Call() 返回 false
    * 如果服务器死了，或者请求删除了：重发
    * 如果服务器执行了，但是请求丢失了：重发危险
* 问题
    * 这两个案例对客户端看起来一样(没有返回)
    * 如果已经执行了，客户端仍然需要结果
* 思想：重复 RPC 检测
    * 我们让 k/v 服务检测重复的客户端请求
    * 客户端为每个请求选择一个 ID,在 RPC 中发送
        * 在重发中相同的 RPC 有相同的 ID
    * k/v 服务维护以 ID 为所有的表
    * 为每个 RPC 生成一个条目
        * 在执行后记录值
    * 如果有相同 ID 的第二个 RPC 到达，就是重复的
        * 从表中的值生成回复
* 设计难题
    * 我们何时(如果有的话)可以删除表的条目？
    * 如果新领导者接管了，它如何获得这个重复的表？
    * 如果服务器崩溃了，它如何回复自己的表？
* 保持重复表尽量小的思想
    * 一个客户端一条表条目，而不是一个 RPC 一条记录
    * 每个客户端一次只有一个 RPC 未完成
    * 每个客户端按顺序编号 RPC
    * 当服务器收到客户端的 RPC
        * 它可以忘记客户端编号更小的条目
        * 因为这意味着客户端不会重新发送旧的 RPC
* 一些细节
    * 每个客户端需要一个唯一的客户端 ID -- 可以是一个64位的随机数
    * 客户端在每个 RPC 中发送客户端 ID 和 序列号(seq)
        * 如果重发则发送重复的序列号(seq)
    * 在 k/v 服务中的重复的表用客户端 ID 做索引
        * 只包括序列号(seq)和已执行的值
    * RPC 处理器首先检查表，只有序列号(seq)不在表中才 Start()
    * 每个日志条目必须包括客户端 ID，序列号(seq)
    * 当在 applyCh 出现操作的时候
        * 修改客户端的表条目中的序列号(seq)和值
        * 唤醒在等待的 RPC 处理器(如果有的话)
* 如果在原来的请求执行前又来了一个重复的请求会怎样？
    * 可以调用 Start() (再次调用)
    * 它可能会在日志中出现两次(相同的客户端 ID，相同的序列号(seq))
    * 当命令出现在 applyCh中，如果表中已有记录则不会执行
* 新的领导者如何获得这个重复表？
    * 任何副本应该在执行的时候更新它们的重复表
    * 所以当变成领导者的时候信息已经存在了
* 当服务器崩溃了如何恢复它的表？
    * 如果没有快照，日志重放会填充表记录
    * 如果有快照，快照必须有表的一个拷贝
* 但是
    * k/v 服务器从重复表中返回旧的值
    * 如果表中的回复值是旧的怎么办？
    * 这样可以吗？
* 例子
    * > C1      &nbsp; &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;    C2     
      > --             --      
      >  put(x,10)  &nbsp;&nbsp;&nbsp;         
      >    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;     first send of get(x), 10 > reply dropped       
      >    put(x,20)      
      >      &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;   re-sends get(x), gets 10  from table, not 20
* 线性化怎么表示
    * >C1: |-Wx10-|  &nbsp;&nbsp;&nbsp;&nbsp; &nbsp;&nbsp;&nbsp;&nbsp;         |-Wx20-|   
        C2:   &nbsp;&nbsp;&nbsp;&nbsp; &nbsp;&nbsp;&nbsp;&nbsp; &nbsp;&nbsp;&nbsp;&nbsp;        |-Rx10-------------|
    * 顺序：Wx10 Rx10 Wx20
    * 所以返回记住的(旧的)值10是正确的

### 只读操作 (第8章节结尾)
* 问题：Raft 领导者要不要在回复之前提交只读的操作到日志中？例如 Get(key)？也就是说，领导者可不可以使用 key/value 表的当前内容立刻放回 Get（）？
* 回答：
    * 不用，不是图2中的方案或实验中的方案
    * 假设 S1 认为它是领导者，并收到了 Get(K)。它可能最近错过了选举，但没有意识到，由于网络数据包丢失。
    * 新的领导者 S2，可能已经执行了这个 key 的 Put()，所以在 S1 的 key/value 表中的值是陈旧的
    * 提供陈旧数据不是线性化，是脑裂
* 所以
    * 图2需要将 Get() 提交到日志中
    * 如果领导者能够提交 Get()，那么(在日志中的这个点)它仍是领导者。在上面例子中的 S1，不知不觉中失去领导力，它无法获得大多数积极的需要提交 Get() 的 AppendEntries 回复，所以它不会回复给客户端
* 但是：许多应用是读多的。提交 Get() 花费时间。有没有防止提交只读操作的方法？这是实际系统中的一个重要考虑因素
* 思想：租约
    * 像下面这么修改 Raft 协议
    * 定义租约周期，例如5秒
    * 每次领导者得到 AppendEntries 大多数之后
        * 它有权在没有向日志提交只读请求的情况下响应租约期间的只读请求，也就是没有发送 AppendEntries
    * 一个新的领导者不能执行 Put() 直到上一个租期时间已经过期
    * 所以追随者会跟踪他们上次回复 AppendEntries 的时间，并通知新的领导者(在 RequestVote 回复中)
    * 结果：更快的只读操作，还是线性的
* 注意：在实验中，你要提交 Get() 到日志中，而不是实现租约
* Spinnaker 可以选择来让读更快，牺牲线性化
    * Spinnaker 的“时间线读取”不需要反映最近的写入
        * 允许返回一个旧(虽然已提交)的值
    * Spinnaker 利用这种自由来加速读取
        * 任何副本可以回复一个读操作，允许读取负载并行化
    * 但是时间线读取不是线性化的
        * 副本可能还没得到最近的写入
        * 副本可能和领导者在不同分区
* 在实践中，人们经常（但并不总是）愿意使用过时数据以换取更高的性能