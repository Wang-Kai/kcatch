package kcache

import (
	"container/list"
	"errors"
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

type bucket struct {
	sync.Map
}

type Cache struct {
	ActivePeriod    time.Duration // 用户停止活跃后，灰度信息可存活周期
	AvailablePeriod time.Duration // 灰度信息可用周期
	getsPerPromote  int16         // 缓存数据被提升需要被 get 的次数
	buckets         []*bucket     // 存储缓存数据的 hashmap
	bucketMask      uint32        // 用来随机分配数据存放的掩码
	list            *list.List    // 用来实现 LRU 的双向链表

	deleteChan  chan *Item
	promoteChan chan *Item
}

func New(activePeriod, availablePeriod time.Duration) (*Cache, error) {
	if availablePeriod > activePeriod {
		return nil, errors.New("The active period should be longer than available period")
	}

	c := &Cache{
		list:            list.New(),
		ActivePeriod:    activePeriod,
		AvailablePeriod: availablePeriod,
		bucketMask:      15,
		getsPerPromote:  10,
		buckets:         make([]*bucket, 16),
	}

	for i := 0; i < 16; i++ {
		c.buckets[i] = new(bucket)
	}

	c.start()

	return c, nil
}

func (c *Cache) refreshTTL(item *Item) {
	item.shouldDeleteAt = time.Now().Add(c.ActivePeriod)
}

func (c *Cache) start() {
	c.deleteChan = make(chan *Item, 1024)
	c.promoteChan = make(chan *Item, 1024)

	go c.worker()
	go c.gc()
}

// gc will delete inactive user info
// 淘汰一定时间内不活跃的缓存数据
func (c *Cache) gc() {
	gcPeriod := int64(c.ActivePeriod.Nanoseconds()) / 2
	for t := range time.Tick(time.Duration(gcPeriod)) {
		fmt.Printf("=========> %v trigger GC \n", t.UTC())
		element := c.list.Back()
		for element != nil {
			fmt.Printf("=========> %v GC Start \n", t.UTC())
			preElement := element.Prev()
			item, _ := element.Value.(*Item)

			if item.InActive() {
				c.bucket(item.key).Delete(item.key)
				c.list.Remove(element)
				element = preElement

				fmt.Printf("=========> Delete %s \n", item.key)
			} else {

				fmt.Printf("=========> %v GC end \n", t.UTC())
				break
			}
		}
	}
}

// worker will
func (c *Cache) worker() {
	for {
		select {
		case dItem := <-c.deleteChan:
			if dItem.element == nil {
				dItem.promotions = -2
			} else {
				// delete itemm from linked list
				fmt.Printf("=========> %+v\n", dItem)
				c.list.Remove(dItem.element)
			}
		case pItem := <-c.promoteChan:
			if pItem.promotions == -2 {
				// 已经被删除，do nothing
				return
			}

			// not a new item
			if pItem.element != nil {
				if pItem.shouldPromote(c.getsPerPromote) {
					c.list.MoveToFront(pItem.element)
				}
				return
			}
			// new item
			fmt.Printf("=========> Insert a new item \n")
			pItem.element = c.list.PushFront(pItem)
		}
	}
}

func (c *Cache) Get(key string) *Item {
	bucket := c.bucket(key)
	val, ok := bucket.Load(key)
	if !ok {
		return nil
	}
	item, _ := val.(*Item)

	if item.Unavailable() {
		// 数据失效后应该重新从灰度配置服务中读取最新灰度信息
		newVal := refreshUserInfo(item)
		item = c.Set(key, newVal)
	} else {
		// 更新数据的活跃期
		c.refreshTTL(item)
	}

	c.promoteChan <- item
	return item
}

func (c *Cache) bucket(key string) *bucket {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.buckets[h.Sum32()&c.bucketMask]
}

// set save user info into cache
func (c *Cache) Set(key string, value interface{}) *Item {
	now := time.Now()

	item := &Item{
		value:          value,
		key:            key,
		shouldDeleteAt: now.Add(c.ActivePeriod),
		shouldUpdateAt: now.Add(c.AvailablePeriod),
	}

	bucket := c.bucket(key)

	oldValue, ok := bucket.Load(key)
	if ok {
		// delete oldValue from linked list
		oldItem, _ := oldValue.(*Item)
		c.deleteChan <- oldItem
	}

	c.bucket(key).Store(key, item)

	// promote new item on link
	c.promoteChan <- item

	return item
}
