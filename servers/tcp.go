package servers

import (
	"cache-server/caches"
	"cache-server/proto"
	"encoding/binary"
	"encoding/json"
	"errors"
)

const (
	getCommand    = byte(1)
	setCommand    = byte(2)
	deleteCommand = byte(3)
	statusCommand = byte(4)
)

var (
	errCommandNeedsMoreArguments = errors.New("command needs more arguments")
	errNotFound                  = errors.New("not found")
)

type TCPServer struct {
	cache  *caches.Cache // 内部用于存储数据的缓存组件
	server *proto.Server //  内部真正用于服务的服务器
}

// 返回TCP服务器
func NewTCPServer(cache *caches.Cache) *TCPServer {
	return &TCPServer{
		cache:  cache,
		server: proto.NewServer(),
	}
}

// 运行TCP服务器
func (s *TCPServer) Run(address string) error {
	// 注册处理函数
	s.server.RegisterHandler(getCommand, s.getHandler)
	s.server.RegisterHandler(setCommand, s.setHandler)
	s.server.RegisterHandler(deleteCommand, s.deleteHandler)
	s.server.RegisterHandler(statusCommand, s.statusHandler)
	return s.server.ListenAndServe("tcp", address)
}

// 关闭服务器
func (s *TCPServer) Close() error {
	return s.server.Close()
}

// 处理get指令
func (s *TCPServer) getHandler(args [][]byte) (body []byte, err error) {
	if len(args) < 1 {
		return nil, errCommandNeedsMoreArguments
	}

	// 调用缓存Get方法 如果不存在则返回NotFound错误
	value, ok := s.cache.Get(string(args[0]))
	if !ok {
		return value, errNotFound
	}
	return value, nil
}

// 处理set指令
func (s *TCPServer) setHandler(args [][]byte) (body []byte, err error) {
	if len(args) < 3 {
		return nil, errCommandNeedsMoreArguments
	}

	// 读取ttl 使用大端方式读取 客户端同样使用大端方式存储
	ttl := int64(binary.BigEndian.Uint64(args[0]))
	err = s.cache.SetWithTTL(string(args[1]), args[2], ttl)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// 处理delete指令
func (s *TCPServer) deleteHandler(args [][]byte) (body []byte, err error) {
	if len(args) < 1 {
		return nil, errCommandNeedsMoreArguments
	}
	err = s.cache.Delete(string(args[0]))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// 处理status指令
func (s *TCPServer) statusHandler(args [][]byte) (body []byte, err error) {
	return json.Marshal(s.cache.Status())
}

func NewServer(serverType string, cache *caches.Cache) Server {
	if serverType == "tcp" {
		return NewTCPServer(cache)
	}
	return NewHTTPServer(cache)
}
