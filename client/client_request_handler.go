package client

import (
	"fmt"
	"net"

	"github.com/t0rr3sp3dr0/middleair/crypto"
	"github.com/t0rr3sp3dr0/middleair/util"
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
		return nil, util.ErrMethodNotAllowed
	}

	conn, err := net.Dial(options.Protocol, fmt.Sprintf("%s:%d", options.Host, options.Port))
	if err != nil {
		return nil, err
	}
	secureConn, err := crypto.NewSecureConn(conn)
	if err != nil {
		return nil, err
	}

	if _, err := secureConn.WriteData(options.Credentials); err != nil {
		defer secureConn.Close()
		return nil, err
	}

	data, err := secureConn.ReadData()
	if err != nil {
		defer secureConn.Close()
		return nil, err
	}

	if data[0] != 200 {
		defer secureConn.Close()
		switch data[0] {
		case 401 % 256:
			return nil, util.ErrUnauthorized

		case 403 % 256:
			return nil, util.ErrForbidden

		default:
			return nil, util.ErrUnknown
		}
	}

	e := &ClientRequestHandler{
		options: options,
		netConn: *secureConn,
	}

	return e, nil
}

func (e *ClientRequestHandler) Close() error {
	return e.netConn.Close()
}

func (e *ClientRequestHandler) Send(message []byte) error {
	_, err := e.netConn.WriteData(message)
	return err
}

func (e *ClientRequestHandler) Receive() ([]byte, error) {
	return e.netConn.ReadData()
}
