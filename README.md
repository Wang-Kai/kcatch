## 问题

有一个面向应用客户端的 API 网关，10 万级 QPS，千万 DAU，为了支持基于用户灰度能力，需要实现一个缓存模块，它会基于用户 ID 缓存用户的灰度配置信息（小于 1k），要求：

1. 灰度配置信息存在有效期（比如10分钟），数据失效后应该重新从灰度配置服务中读取最新灰度信息；
2. 为了节约内存，应该淘汰一定时间内不活跃的缓存数据。

## 思路

1. 使用 [rsync.Map](https://golang.org/pkg/sync/#Map) 来做数据的存储
2. 使用 [container/list](https://golang.org/pkg/container/list/) 来实现 LRU 算法
3. 要具备 GC 能力，回收 cold data



开启一个 GC goroutine，定时（活跃周期的一半）删除不活跃的用户数据，保证内存中没有不活跃的缓存数据

实现 `lazy update` 方案，每次从缓存中 `Get` 先判断返回值是否已过期，如果过期就从灰度配置服务中读取最新灰度信息，并更新到缓存中


## 实现

### 创建了 16 个 hashmap，数据随机落到每个 hashmap 中

1. 减少 hashmap 冲突带来的查询复杂度
2. 读写互斥场景下，避免 ”一写全不可读“ 的情况，即使一个 hashmap 在写入，但其他 15 个 hashmap 均可并发读

![](https://steve-1254173768.cos.ap-shanghai.myqcloud.com/random-hashmap.png)

### `lazy update` 解决有效期问题，定时 GC 删除不活跃用户



# kcache

kcache 实现了一个用户灰度信息的缓存模块，具备如下能力：

- 缓存数据超时淘汰
- 实现 LRU cache invalidation algorithms，仅保留 hot data 在内存中

优点：

1. 最大化的节约内存，提升查询速度
2. 应对高并发

### 场景条件

1. 10 万级 QPS
2. 千万 DAU

### 处理策略

1. 因为日活用户基数很高，所以 LRU 实现中，做缓存提升需达到一定指标
2. 由于用户基数大，单一加锁 hashmap 会成为瓶颈，可以把缓存数据放在多个 hashmap 中
3. 高并发下，缓存数据要尽可能多的处于可读状态，所以要减少写的频次，当数据超过上限，可以一次删一批不活跃的用户。


### 具体实现

1. 用户缓存数据分多个 sync.Map 来存储
2. 使用一个双向链表来实现 LRU，提升活跃用户信息，删除不活跃用户信息
3. 


要求：

1. 处理失效的数据，不断更新缓存模块数据 （lazy update）
2. 淘汰不活跃数据 (regular delete)
