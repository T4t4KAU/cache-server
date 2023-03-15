package caches

import (
	"errors"
	"sync"
)

// 数据块 将锁和数据放置内部
type segment struct {
	Data    map[string]*value // 存储数据块数据
	Status  *Status           // 记录该数据块状态
	options *Options          // 选项设置
	mutex   *sync.RWMutex     // 用于保证该数据块并发安全
}

// 返回一个使用options初始化过的segment实例
func newSegment(options *Options) *segment {
	return &segment{
		Data:    make(map[string]*value, options.MapSizeOfSegment),
		Status:  NewStatus(),
		options: options,
		mutex:   &sync.RWMutex{},
	}
}

// 返回指定key数据
func (seg *segment) get(key string) ([]byte, bool) {
	seg.mutex.RLock()
	defer seg.mutex.RUnlock()
	value, ok := seg.Data[key]
	if !ok {
		return nil, false
	}
	if !value.alive() {
		seg.mutex.RUnlock()
		seg.delete(key)
		seg.mutex.RLock()
		return nil, false
	}
	return value.visit(), true
}

// 将一个数据添加进segment
func (seg *segment) set(key string, value []byte, ttl int64) error {
	seg.mutex.Lock()
	defer seg.mutex.Unlock()
	if oldValue, ok := seg.Data[key]; ok {
		seg.Status.subEntry(key, oldValue.Data)
	}
	if !seg.checkEntrySize(key, value) {
		if oldValue, ok := seg.Data[key]; ok {
			seg.Status.addEntry(key, oldValue.Data)
		}
		return errors.New("the entry size will exceed if you set this entry")
	}
	seg.Status.addEntry(key, value)
	seg.Data[key] = newValue(value, ttl)
	return nil
}

// 从segment中删除指定key
func (seg *segment) delete(key string) {
	seg.mutex.Lock()
	defer seg.mutex.Unlock()
	if oldValue, ok := seg.Data[key]; ok {
		seg.Status.subEntry(key, oldValue.Data)
		delete(seg.Data, key)
	}
}

// 返回该segment状态
func (seg *segment) status() Status {
	seg.mutex.RLock()
	defer seg.mutex.RUnlock()
	return *seg.Status
}

// 判断segment数据容量是否已经到了设定的上限
func (seg *segment) checkEntrySize(newKey string, newValue []byte) bool {
	return seg.Status.entrySize()+int64(len(newKey))+int64(len(newValue)) <=
		int64((seg.options.MaxEntrySize*1024*1024)/seg.options.SegmentSize)
}

// 清理segment中过期数据
func (seg *segment) gc() {
	seg.mutex.Lock()
	defer seg.mutex.Unlock()
	count := 0

	for key, value := range seg.Data {
		if !value.alive() {
			seg.Status.subEntry(key, value.Data)
			delete(seg.Data, key)
			count++
			if count >= seg.options.MaxGcCount {
				break
			}
		}
	}
}
