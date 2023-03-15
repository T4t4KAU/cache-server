package caches

// 缓存信息
type Status struct {
	Count     int   `json:"count"`     // 记录缓存数据个数
	KeySize   int64 `json:"keySize"`   // 记录key占用空间大小
	ValueSize int64 `json:"valueSize"` // 记录value占用空间大小
}

// 返回一个缓存信息对象指针
func NewStatus() *Status {
	return &Status{
		Count:     0,
		KeySize:   0,
		ValueSize: 0,
	}
}

// 储存key-value
func (s *Status) addEntry(key string, value []byte) {
	s.Count++
	s.KeySize += int64(len(key))
	s.ValueSize += int64(len(value))
}

// 删除key-value
func (s *Status) subEntry(key string, value []byte) {
	s.Count--
	s.KeySize -= int64(len(key))
	s.ValueSize -= int64(len(value))
}

// 返回key-value占用总和
func (s *Status) entrySize() int64 {
	return s.KeySize + s.ValueSize
}
