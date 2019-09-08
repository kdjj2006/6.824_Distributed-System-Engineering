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

