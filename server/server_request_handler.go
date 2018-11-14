package server

import (
	"fmt"
	"net"

	model "../proto"
	"../util"
	"github.com/golang/protobuf/proto"
)

type ServerRequestHandler struct {
	options  util.Options
	listener net.Listener
	netConn  util.WrapperConn
	pktConn  util.WrapperPacketConn
	address  net.Addr
}

func NewServerRequestHandler(options util.Options) (*ServerRequestHandler, error) {
	e := &ServerRequestHandler{
		options: options,
	}
	return e, nil
}

func (e *ServerRequestHandler) Close() error {
	switch e.options.Protocol {
	case "udp":
		return e.pktConn.Close()

	case "tcp":
		if err := e.netConn.Close(); err != nil {
			return err
		}
		return e.listener.Close()

	default:
		return fmt.Errorf("Unknown Protocol")
	}
}

func (e *ServerRequestHandler) Receive() ([]byte, error) {
	switch e.options.Protocol {
	case "udp":
		return e.udpReceive()

	case "tcp":
		return e.tcpReceive()

	default:
		return nil, fmt.Errorf("Unknown Protocol")
	}
}

func (e *ServerRequestHandler) udpReceive() ([]byte, error) {
	if e.pktConn.PacketConn == nil {
		ln, err := net.ListenPacket("udp", fmt.Sprintf(":%d", e.options.Port))
		if err != nil {
			return nil, err
		}
		e.pktConn = util.WrapperPacketConn{ln}
	}

	bytes, addr, err := e.pktConn.ReadData()
	e.address = addr

	return bytes, err
}

func (e *ServerRequestHandler) tcpReceive() ([]byte, error) {
	if e.netConn.Conn == nil {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", e.options.Port))
		if err != nil {
			return nil, err
		}
		e.listener = ln

		conn, err := e.listener.Accept()
		if err != nil {
			return nil, err
		}
		e.netConn = util.WrapperConn{conn}
	}

	return e.netConn.ReadData()
}

func (e *ServerRequestHandler) Send(message []byte) error {
	switch e.options.Protocol {
	case "udp":
		return e.udpSend(message)

	case "tcp":
		return e.tcpSend(message)

	default:
		return fmt.Errorf("Unknown Protocol")
	}
}

func (e *ServerRequestHandler) udpSend(message []byte) error {
	_, err := e.pktConn.WriteData(e.address, message)
	return err
}

func (e *ServerRequestHandler) tcpSend(message []byte) error {
	_, err := e.netConn.WriteData(message)
	return err
}

func (e *ServerRequestHandler) handleBadRequest(err error) {
	er := &model.ErrorResponse{
		Error: &model.Error{
			Code:    400,
			Message: err.Error(),
		},
	}

	data, err := proto.Marshal(er)
	if err != nil {
		panic(err)
	}

	if err := e.Send(data); err != nil {
		panic(err)
	}
}

func (e *ServerRequestHandler) handleInternalServerError(err error) {
	er := &model.ErrorResponse{
		Error: &model.Error{
			Code:    500,
			Message: err.Error(),
		},
	}

	data, err := proto.Marshal(er)
	if err != nil {
		panic(err)
	}

	if err := e.Send(data); err != nil {
		panic(err)
	}
}
