package client

import (
	"fmt"
	"net"

	"../util"
)

type ClientRequestHandler struct {
	options util.Options
	netConn util.WrapperConn
}

func NewClientRequestHandler(options util.Options) (*ClientRequestHandler, error) {
	switch options.Protocol {
	case "udp":
	case "tcp":
	default:
		return nil, fmt.Errorf("Unknown Protocol")
	}

	e := &ClientRequestHandler{
		options: options,
	}
	return e, nil
}

func (e *ClientRequestHandler) Close() error {
	return e.netConn.Close()
}

func (e *ClientRequestHandler) Send(message []byte) error {
	return e.send(message)
}

func (e *ClientRequestHandler) send(bytes []byte) error {
	if e.netConn.Conn == nil {
		conn, err := net.Dial(e.options.Protocol, fmt.Sprintf("%s:%d", e.options.Host, e.options.Port))
		if err != nil {
			return err
		}
		e.netConn = util.WrapperConn{conn}
	}

	_, err := e.netConn.WriteData(bytes)
	return err
}

func (e *ClientRequestHandler) Receive() ([]byte, error) {
	return e.receive()
}

func (e *ClientRequestHandler) receive() ([]byte, error) {
	return e.netConn.ReadData()
}
