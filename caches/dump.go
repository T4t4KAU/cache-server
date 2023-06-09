package caches

import (
	"encoding/gob"
	"os"
	"sync"
	"time"
)

// 持久化结构体
type dump struct {
	SegmentSize int
	Segments    []*segment
	Options     *Options
}

// 返回空持久化实例
func newEmptyDump() *dump {
	return &dump{}
}

// 返回一个从缓存实例初始化过来的持久化实例
func newDump(c *Cache) *dump {
	return &dump{
		SegmentSize: c.segmentSize,
		Segments:    c.segments,
		Options:     c.options,
	}
}

// 时间戳
func nowSuffix() string {
	return "." + time.Now().Format("20060102150405")
}

// 将dump实例持久化文件中
func (d *dump) to(dumpFile string) error {
	newDumpFile := dumpFile + nowSuffix()
	file, err := os.OpenFile(newDumpFile,
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	err = gob.NewEncoder(file).Encode(d)
	if err != nil {
		file.Close()
		os.Remove(newDumpFile)
		return err
	}
	os.Remove(dumpFile)
	file.Close()
	return os.Rename(newDumpFile, dumpFile)
}

// 从dump文件中恢复cache结构对象
func (d *dump) from(dumpFile string) (*Cache, error) {
	file, err := os.Open(dumpFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if err = gob.NewDecoder(file).Decode(d); err != nil {
		return nil, err
	}
	for _, segment := range d.Segments {
		segment.options = d.Options
		segment.mutex = &sync.RWMutex{}
	}
	return &Cache{
		segmentSize: d.SegmentSize,
		segments:    d.Segments,
		options:     d.Options,
		dumping:     0,
	}, nil
}
