package kcache

import (
	"log"
	"testing"
	"time"

	"github.com/thanhpk/randstr"
)

//
func TestSetCache(t *testing.T) {
	c, err := New(time.Second*20, time.Second*2)
	if err != nil {
		t.Fatal(err)
	}

	c.Set("user:1", "steve")
	c.Set("user:1", "haiting")
	c.Set("user:1", "kiko")

	val := c.Get("user:1")

	t.Logf("==> %v", val.Value())
}

func TestGetCache(t *testing.T) {
	c, err := New(time.Second*20, time.Second*2)
	if err != nil {
		t.Fatal(err)
	}

	c.Set("user:1", "steve")

	time.Sleep(time.Second * 1)

	val := c.Get("user:1").Value()
	t.Logf("After 1s, Value ==> %+v\n", val)

	time.Sleep(time.Second * 3)
	val = c.Get("user:1").Value()
	t.Logf("After 4s, Value ==> %+v\n", val)

	time.Sleep(time.Second * 1)
	val = c.Get("user:1").Value()
	t.Logf("After 5s, Value ==> %+v\n", val)

}

func TestGC(t *testing.T) {
	c, err := New(time.Second*4, time.Second*3)
	if err != nil {
		log.Fatal(err)
	}

	var tryTimes = 5

	for tryTimes > 0 {
		for i := 0; i < 9; i++ {
			c.Set(randstr.Hex(16), randstr.Hex(16))
		}
		time.Sleep(time.Second * 2)

		tryTimes--
	}

	time.Sleep(time.Second * 100)
}
