6.824 2017 Lecture 8: Zookeeper Case Study

* 读论文： "ZooKeeper: wait-free coordination for internet-scale systems"
* 我们为什么读这篇论文？
    * 广泛应用的复制状态机服务
        * 受到 Chubby 的启发(Google 的全局锁服务)
        * 最开始在雅虎，现在外面公司也用(Mesos, HBase等)
    * 开源
        * 作为 Apache 项目(http://zookeeper.apache.org/)
    * 构建复制服务的案例学习，提供一个 Paxos/ZAB/Raft 库
        * 实验3中也出现了类似的问题
    * API 支持广泛的用例
        * 需要容错 "master" 的应用程序不需要自己轮训获取
        * Zookeeper 很通用，他们应该可以使用Zookeeper
    * 高性能
        * 不像实验3中的复制 key/value 服务
* 动机：数据中心集群中的许多应用程序需要协调
    * 例子：GFS
        * master 拥有每个块的块服务器列表 
        * master 决定哪个块服务器是主块服务器
        * 等等
    * 其他的例子：YMB, Crawler 等
        * YMB 需要 master 来切分主题
        * Crawler 需要 master 控制页面抓取(例如有点像 mapreduce 中的 master)
    * 应用程序也需要互相发现
        * MapReduce 需要知道 GFS master 的 IP:端口
        * 负载均衡需要知道 web 服务器在哪
    * 协调服务通常用于这些目的
* 动机：性能 -- 实验3
    * 由 Raft 主导
    * 考虑3节点的 Raft
    * 在回复客户端之前，Raft 执行
        * 领导者持久化日志条目
        * 同时，领导者发送消息给追随者
            * 每个追随者持久化日志条目
            * 每个追随者应答
        * 每个回合2次磁盘写入
            * 如果磁盘：2*10msec = 50 msg/sec
            * 如果 SSD：2*2msec+1msec = 200 msg/sec
        *  Zookeeper 执行 21,000 msg/sec
            * 异步调用
            * 允许流水线
* 替代计划：为每个应用开发容错 master
    * 通过 DNS 发布位置
    * 如果写入 master 不复杂是 OK 的
    * 但是，master 经常需要
        * 容错
            * 每个应用程序都说明如何使用 Raft？
        * 高性能
            * 每个应用程序都说明如何让读操作更快？
    * DNS 传播很慢
        * 故障转移需要很长时间
    * 一些应用适用于单点故障
        * 例如 GFS 和 MapReduce
        * 不太理想
* Zookeeper：一个通用的协调服务
    * 设计挑战
        * 提供什么 API？
        * 怎么令 master 容错
        * 如何获得高性能
    * 互动挑战
        * 高性能可能影响 API
        * 例如异步接口允许流水线