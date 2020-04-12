package kcache

import (
	"github.com/thanhpk/randstr"
)

func refreshUserInfo(item *Item) interface{} {
	// generate fake user info
	return randstr.Hex(16)
}
