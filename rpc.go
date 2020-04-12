package kcache

import (
	"github.com/thanhpk/randstr"
)

func refreshUserInfo(item *Item) interface{} {
	return randstr.Hex(16)
}
