package proto

import (
	"bufio"
	"errors"
	"io"
	"net"
)

type Client struct {
	conn   net.Conn // 服务器连接
	reader io.Reader
}

func NewClient(network string, address string) (*Client, error) {
	// 和服务端建立连接
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

// 执行命令
func (c *Client) Do(command byte, args [][]byte) (body []byte, err error) {
	// 封装请求后发送给服务端
	_, err = writeRequestTo(c.conn, command, args)
	if err != nil {
		return nil, err
	}
	// 读取服务端返回的响应
	reply, body, err := readResponseFrom(c.reader)
	if err != nil {
		return body, err
	}

	// 错误响应码
	if reply == ErrorReply {
		return body, errors.New(string(body))
	}
	return body, nil
}

// 关闭客户端
func (c *Client) Close() error {
	return c.conn.Close()
}
