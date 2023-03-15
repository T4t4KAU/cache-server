package proto

import (
	"bufio"
	"errors"
	"net"
	"strings"
	"sync"
)

var (
	// 未找到对应命令处理器
	errCommandHandlerNotFound = errors.New("failed to find a handler of command")
)

type Server struct {
	listener net.Listener
	handlers map[byte]func(args [][]byte) (body []byte, err error) // 处理函数
}

// 创建新服务器
func NewServer() *Server {
	return &Server{
		handlers: map[byte]func(args [][]byte) (body []byte, err error){},
	}
}

// 注册命令处理器
func (s *Server) RegisterHandler(command byte, handler func(args [][]byte) (body []byte, err error)) {
	s.handlers[command] = handler
}

// 监听并处理连接
func (s *Server) ListenAndServe(network string, address string) (err error) {
	s.listener, err = net.Listen(network, address)
	if err != nil {
		return err
	}

	// 使用WaitGroup记录连接数 等待所有连接处理完毕
	wg := &sync.WaitGroup{}
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			}
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.handleConn(conn)
		}()
	}

	// 等待所有连接处理完毕
	wg.Wait()
	return nil
}

// 处理连接
func (s *Server) handleConn(conn net.Conn) {
	reader := bufio.NewReader(conn)
	defer conn.Close()
	for {
		// 读取并解析请求
		command, args, err := readRequestFrom(reader)
		if err != nil {
			if err == errProtocolVersionMismatch {
				continue
			}
			return
		}

		// 处理请求
		reply, body, err := s.handleRequest(command, args)
		if err != nil {
			writeErrorResponseTo(conn, err.Error())
			continue
		}

		// 发送处理结果
		_, err = writeResponseTo(conn, reply, body)
		if err != nil {
			continue
		}
	}
}

// 处理请求
func (s *Server) handleRequest(command byte, args [][]byte) (reply byte, body []byte, err error) {
	handle, ok := s.handlers[command] // 获取对应处理函数
	if !ok {
		return ErrorReply, nil, errCommandHandlerNotFound
	}

	// 将处理结果返回
	body, err = handle(args)
	if err != nil {
		return ErrorReply, body, err
	}
	return SuccessReply, body, err
}

// 关闭服务端的方法
func (s *Server) Close() error {
	if s.listener == nil {
		return nil
	}
	return s.listener.Close()
}
