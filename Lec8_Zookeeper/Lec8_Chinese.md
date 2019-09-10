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
* Zookeeper API 概览
    * 复制状态机
        * 多个服务器实现这个服务
        * 操作以全局顺序执行
            * 一些特例，如果一致性不重要
        * 复制的对象是：znodes
            * znodes 的层级
                * 以 pathnames 命名
            * znodes 包含应用的元数据
                * 配置信息
                    * 参与应用程序的机器
                    * 哪个机器是主机
                * 时间戳
                * 版本号
            * znodes 类型
                * 正常类型
                * 短暂类型
                * 顺序的：名字 + 序列号
                    * 如果 n 是新的 znode 而 p 是父 znode，那么序列 n 的值永远不会小于任何 p 下面其他名称中的创建的顺序的 znode 的值
    * session
        * 客户端登录到 zookeeper
        * session 允许客户端可以故障转移到另一台 Zookeeper 服务
            * 客户端知道上一次完成的操作(zxid)的任期和索引
            * 在每个请求中发送
                * 服务只有在客户已经看到的情况下执行操作
        * session 可以超时
            * 客户端必须连续刷新 session
                * 发送一个心跳给服务器(像一个租期)
            * Zookeeper 如果没有从客户端收到消息则认为客户端“死了”
            * 客户端可能一直在做这个(例如网络分区)
                * 但是不能在那个 session 中执行其他 Zookeeper 操作
        * 在Raft + 实验 3 KV 存储中没有类似的东西
* znodes 上的操作
    * create(path, data, flags)
    *  delete(path, version)
        * 如果 znode.version = version, 然后删除
    * exists(path, watch)
    * getData(path, watch)
    * setData(path, data, version)
        * 如果 znode.version = version, 然后更新
    * getChildren(path, watch)
    * sync()
        * 上面的操作是异步的
        * 每个客户端的所有操作是先进先出排序的
        * 同步直到以前所有的操作都被广播了
* 检查：我们可不可以用实验 3 的服务来做到这个？
    * 有缺陷的方案：GFS master 在启动的时候执行 Put("gfs-master", my-ip:port)
        * 其他应用和 GFS 节点执行 Get("gfs-master")
    * 问题：如果两个 master 候选人的 Put() 竞争怎么办？
        * 后来的 Put() 胜出
        * 每个假定的 master 需要读取密钥以查看它是否真的是 master
            * 什么时候我们保证没有延迟 Put() 让我们感到震惊？
            * 其他所有客户端肯定已经看到了我们的 Put() -- 很难保证
    * 问题：当 master 发生故障，谁决定去删除/更新 KV 存储条目
        * 需要某种超时
        * 所以 master 必须存储(my-ip:port, timestamp)的元组
            * 并且一直调用 Put() 来更新时间戳
            * 其他人轮询该条目以查看时间戳是否停止更新了
    * 过多的轮询和不明的竞态行为 -- 复杂
    *  ZooKeeper API 有更好的方法：观察，session，原子的 znode 创建
        * 只有一个创建可以成功 -- 没有 Put() 竞争
        * session 让超时简单 -- 不需要存储和刷新特定的时间戳
        * 观察是懒通知 -- 避免提交过多的轮询读
* 顺序保证
    * 所有的写操作都是全局排序的
        * 如果 ZooKeeper 执行了一个写操作，后续的其他客户端的写可以看到这个操作
            * 例如两个客户端创建了一个 znode，ZooKeeper 按全局顺序执行它们
    * 每个客户端的所有操作都是先进先出的排序
    * 启发：
        * 一个读操作读取来自同一客户端的早期写入的结果
        * 一个读操作读取了一些写入的前缀，可能不包括最近的写入
            * 读可能返回陈旧的数据
        * 如果读取观察到一些写入前缀，则稍后的读取也会观察到该前缀
