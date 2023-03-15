package caches

import (
	"cache-server/utils"
	"sync/atomic"
	"time"
)

const (
	NeverDie = 0
)

type value struct {
	Data    []byte // 数据
	TTL     int64  // 存活时限
	Created int64  // 数据创建时间
}

// 返回一个封装好的数据
func newValue(data []byte, TTL int64) *value {
	return &value{
		Data:    utils.Copy(data),
		TTL:     TTL,
		Created: time.Now().Unix(),
	}
}

// 返回该数据是否存活
func (v *value) alive() bool {
	return v.TTL == NeverDie || time.Now().Unix()-v.Created < v.TTL
}

// 返回该数据实际存储数据
func (v *value) visit() []byte {
	// 更新访问时间
	atomic.SwapInt64(&v.Created, time.Now().Unix())
	return v.Data
}
