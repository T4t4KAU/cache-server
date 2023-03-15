package servers

import (
	"cache-server/caches"
	"cache-server/proto"
	"encoding/binary"
	"encoding/json"
)

// TCP客户端
type TCPClient struct {
	client *proto.Client
}

// 创建TCP客户端
func NewTCPClient(address string) (*TCPClient, error) {
	client, err := proto.NewClient("tcp", address)
	if err != nil {
		return nil, err
	}
	return &TCPClient{
		client: client,
	}, nil
}

// 从缓存中获取指定key-value
func (c *TCPClient) Get(key string) ([]byte, error) {
	return c.client.Do(getCommand, [][]byte{[]byte(key)})
}

// 添加key-value到缓存中
func (c *TCPClient) Set(key string, value []byte, ttl int64) error {
	b := make([]byte, 8)
	// 使用大端形式储存数字
	binary.BigEndian.PutUint64(b, uint64(ttl))
	_, err := c.client.Do(setCommand, [][]byte{
		b, []byte(key), value,
	})
	return err
}

// 删除指定key-value
func (c *TCPClient) Delete(key string) error {
	_, err := c.client.Do(deleteCommand, [][]byte{[]byte(key)})
	return err
}

// 返回缓存状态
func (c *TCPClient) Status() (*caches.Status, error) {
	body, err := c.client.Do(statusCommand, nil)
	if err != nil {
		return nil, err
	}
	status := caches.NewStatus()
	err = json.Unmarshal(body, status)
	return status, err
}

// 关闭客户端
func (c *TCPClient) Close() error {
	return c.client.Close()
}
