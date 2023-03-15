package client

import "encoding/json"

type Status struct {
	Count     int `json:"count"` //
	KeySize   int `json:"leySize"`
	ValueSize int `json:"valueSize"`
}

type request struct {
	command    byte
	args       [][]byte
	resultChan chan *Response
}

type Response struct {
	Body []byte
	Err  error
}

func (r *Response) ToStatus() (*Status, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	status := &Status{}
	return status, json.Unmarshal(r.Body, status)
}
