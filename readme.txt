SmartBuffer

基本需求：
    在多IO操作时缓解内存分配压力
    （多网络连接时）

特性：
    根据 user 的使用场景类别自动选择大小合适的 buffer

这些对象被使用：
    pool
    pools
    block
    buffer
    user
    level
    watcher

关系描述：
    pools , 按照 level 存放 pool 的数组。
    user 使用者，从 pool 申请 buffer 时，具有一个特定的使用场景，因此很自然地抽离出『user』概念，为了给每个场景分别进行统计并依此优化
                接通多个 pool，并且只能向下调用 pool.get() 和 pool.grow(), 对 block 无联系
                具有 maxLevel 属性，约束了 buffer 的最大大小
                同时对 buffer 的使用情况进行统计
                因为 buffer 的 levelup(扩容) 依赖 user 的统计和 maxlevel
                所以 levelup() 主体虽然是 buffer ，但是是 user 的方法
                具有 resize() 方法，根据统计信息调整自身属性来做到优化
    pool，一个 pool 持有多个 block
                每个 block 是一个小池
                get() 会遍历 blocks，如果 block 有空闲资源就取出
                每个 block 具有各自的锁，在 pool.get() 时， pool 自身不需要加锁，可以避免锁的争抢
                grow() 会增加 block
                release() 会释放 block
                没有 put() 方法来放回资源，这是因为 put 的操作并不由 pool 或 user 或 block 进行，而是 buffer
    block，对应一个 buffer 链表，存储链表头
    buffer, 记录了调用方 user 和资源提供方 block 的指针
                具有 GoHome() 方法，相当于 put(), 用来回收资源
                因为 buffer 记录了 block， gohome 可以直接找到回归处
                同时 gohome 时，会调用 user 的 stat_hook() 方法来给 user 更新统计信息
    watcher, 定时任务，给 user 进行 resize， 对 pool 进行 release
                按周期进行调整检查
                过程：
                    检查一个周期内，每个池的使用情况，如果使用率小，进行回收
                    根据 user 的统计情况，调整 user 的默认申请大小和buffer升级路线
                    需要注意的是，改变默认申请大小后，流量会冲击到另一个池，原来的池会因此空闲
                    因此还需要对新旧两个池分别扩容、回收



测试： go test . -v -bench .
    提供了六个基准测试
        goos: linux
        goarch: amd64

        BenchmarkLoops-8                20000000     114 ns/op 测试单次  get put 和 readall 的速度
        BenchmarkParallel-8             10000000     227 ns/op 使用 test 提供的并发测试 get put 和 readall 的速度
        BenchmarkLoopsResize-8          10000000     228 ns/op 测试单次  get put 和 readall 的速度
                                                               这次需要读写的大小是随机的，可能会发生扩容
        BenchmarkParallelResize-8        3000000     475 ns/op 使用 test 提供的并发测试 get put 和 readall 的速度
                                                               这次需要读写的大小是随机的，可能会发生扩容
        BenchmarkParallel2-8             3000000     470 ns/op 使用非 test 提供的并发测试（手动创建了多个 goroutine 来测试）
        BenchmarkParallelResize2-8       3000000     592 ns/op 使用非 test 提供的并发测试（手动创建了多个 goroutine 来测试）
                                                               这次需要读写的大小是随机的，可能会发生扩容


        note： 目前还没有添加验证 watcher 效果的测试