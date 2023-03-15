package client

import (
	"cache-server/proto"
	"encoding/binary"
)

const (
	getCommand    = byte(1)
	setCommand    = byte(2)
	deleteCommand = byte(3)
	statusCommand = byte(4)
)

type AsyncClient struct {
	client      *proto.Client
	requestChan chan *request
}

func NewAsyncClient(address string) (*AsyncClient, error) {
	client, err := proto.NewClient("tcp", address)
	if err != nil {
		return nil, err
	}
	c := &AsyncClient{
		client:      client,
		requestChan: make(chan *request, 163840),
	}
	c.handleRequests()
	return c, nil
}

func (c *AsyncClient) handleRequests() {
	go func() {
		for req := range c.requestChan {
			body, err := c.client.Do(req.command, req.args)
			req.resultChan <- &Response{
				Body: body,
				Err:  err,
			}
		}
	}()
}

func (c *AsyncClient) do(command byte, args [][]byte) <-chan *Response {
	resultChan := make(chan *Response, 1)
	c.requestChan <- &request{
		command:    command,
		args:       args,
		resultChan: resultChan,
	}
	return resultChan
}

func (c *AsyncClient) Get(key string) <-chan *Response {
	return c.do(getCommand, [][]byte{[]byte(key)})
}

func (c *AsyncClient) Set(key string, value []byte, ttl int64) <-chan *Response {
	t := make([]byte, 8)
	binary.BigEndian.PutUint64(t, uint64(ttl))
	return c.do(setCommand, [][]byte{
		t, []byte(key), value,
	})
}

func (c *AsyncClient) Delete(key string) <-chan *Response {
	return c.do(deleteCommand, [][]byte{[]byte(key)})
}

func (c *AsyncClient) Status() <-chan *Response {
	return c.do(statusCommand, nil)
}

func (c *AsyncClient) Close() error {
	close(c.requestChan)
	return c.client.Close()
}
