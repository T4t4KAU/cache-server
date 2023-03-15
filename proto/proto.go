package proto

import (
	"errors"
)

const (
	ProtocolVersion        = byte(1)
	headerLengthInProtocol = 6
	argsLengthInProtocol   = 4
	argLengthInProtocol    = 4
	bodyLengthInProtocol   = 4
)

var (
	errProtocolVersionMismatch = errors.New("protocol version between client and server doesn't match")
)
