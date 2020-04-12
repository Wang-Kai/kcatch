package kcache

import (
	"container/list"
	"time"
)

type Item struct {
	value          interface{}
	key            string
	promotions     int16
	shouldDeleteAt time.Time
	shouldUpdateAt time.Time

	element *list.Element
}

func (i *Item) Value() interface{} {
	return i.value
}

func (i *Item) Unavailable() bool {
	return i.shouldUpdateAt.Before(time.Now())
}

func (i *Item) InActive() bool {
	return i.shouldDeleteAt.Before(time.Now())
}

func (i *Item) shouldPromote(getsPerPromote int16) bool {
	i.promotions += 1
	return i.promotions > getsPerPromote
}
