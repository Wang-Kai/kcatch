### 问题

有一个面向应用客户端的 API 网关，10 万级 QPS，千万 DAU，为了支持基于用户灰度能力，需要实现一个缓存模块，它会基于用户 ID 缓存用户的灰度配置信息（小于 1k），要求：

1. 灰度配置信息存在有效期（比如10分钟），数据失效后应该重新从灰度配置服务中读取最新灰度信息；
2. 为了节约内存，应该淘汰一定时间内不活跃的缓存数据。

### 思路

1. 使用 [rsync.Map](https://golang.org/pkg/sync/#Map) 来做数据的存储
2. 使用 [container/list](https://golang.org/pkg/container/list/) 来实现 [LRU](https://en.wikipedia.org/wiki/Cache_replacement_policies#Least_Recently_Used) 算法
3. 为每个 value 绑定有效期 & 活跃期时间属性
4. 具备 GC 能力，回收 cold data


### 实现

#### 1. 创建了 16 个 hashmap，数据随机落到每个 hashmap 中

1. 因为数据量很大，分散在多个 hashmap 中可以减少 hashmap ”冲突“ 带来的查询复杂度
2. 读写互斥场景下，避免 ”一写全不可读“ 的情况，即使一个 hashmap 在写入，但其他 15 个 hashmap 均可并发读

#### 2. `lazy update` 解决有效期问题，定时 GC 删除不活跃用户

1. 在调用方调用 `Get` 方法的时候检查数据是否过期，如果过期则从灰度配置服务中读取最新灰度信息，并异步更新到缓存中
2. 默认以 ”用户活跃期“ 的一半为周期，定时执行 GC 操作，从双向链表的 tail 起，开始向前执行 ”不活跃删除“，直到遇到第一个活跃的缓存数据
3. 为应对高并发场景改进 LRU 算法，设置 `getsPerPromote` 参数，当请求总数到 10 时才对该用户信息做 ”提升“，减轻高并发场景下 `worker goroutine` 的压力