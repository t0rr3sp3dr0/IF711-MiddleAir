package client

import (
	"fmt"
	"net"

	"../crypto"
	"../util"
)

type ClientRequestHandler struct {
	options util.Options
	netConn crypto.SecureConn
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
		secureConn, err := crypto.NewSecureConn(conn)
		if err != nil {
			return err
		}

		if _, err := secureConn.WriteData(e.options.Credentials); err != nil {
			defer secureConn.Close()
			return err
		}

		data, err := secureConn.ReadData()
		if err != nil {
			defer secureConn.Close()
			return err
		}

		if data[0] != 200 {
			defer secureConn.Close()
			switch data[0] {
			case 401 % 256:
				return util.ErrUnauthorized

			case 403 % 256:
				return util.ErrForbidden

			default:
				return util.ErrUnknown
			}
		}

		e.netConn = *secureConn
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
