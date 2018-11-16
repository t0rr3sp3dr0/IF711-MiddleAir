package server

import (
	"fmt"
	"net"
	"sync"

	"../crypto"
	model "../proto"
	"../util"
	"github.com/golang/protobuf/proto"
)

var (
	listeners                = make(map[uint16]net.Listener)
	listenersMutex           = &sync.RWMutex{}
	listenersWaitGroups      = make(map[uint16]*sync.WaitGroup)
	listenersWaitGroupsMutex = &sync.RWMutex{}
)

type ServerRequestHandler struct {
	options  util.Options
	listener net.Listener
	netConn  crypto.SecureConn
}

func NewServerRequestHandler(options util.Options) (*ServerRequestHandler, error) {
	e := &ServerRequestHandler{
		options: options,
	}

	listenersMutex.Lock()
	listener, ok := listeners[e.options.Port]
	if !ok {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", e.options.Port))
		if err != nil {
			return nil, err
		}
		listener = ln
		listeners[e.options.Port] = listener

		waitGroup := &sync.WaitGroup{}
		waitGroup.Add(1)
		defer waitGroup.Done()
		listenersWaitGroupsMutex.Lock()
		listenersWaitGroups[e.options.Port] = waitGroup
		listenersWaitGroupsMutex.Unlock()
		go func() {
			waitGroup.Wait()

			listenersMutex.Lock()
			delete(listeners, e.options.Port)
			listenersMutex.Unlock()

			listenersWaitGroupsMutex.Lock()
			delete(listenersWaitGroups, e.options.Port)
			listenersWaitGroupsMutex.Unlock()

			if err := e.listener.Close(); err != nil {
				panic(err)
			}
		}()
	}
	listenersMutex.Unlock()
	e.listener = listener

	listenersWaitGroupsMutex.RLock()
	listenersWaitGroups[e.options.Port].Add(1)
	listenersWaitGroupsMutex.RUnlock()

	return e, nil
}

func (e *ServerRequestHandler) Accept() error {
	if e.netConn.Conn != nil {
		return fmt.Errorf("Already Accepted")
	}

	conn, err := e.listener.Accept()
	if err != nil {
		return err
	}
	secureConn, err := crypto.NewSecureConn(conn)
	if err != nil {
		return err
	}
	e.netConn = *secureConn

	return nil
}

func (e *ServerRequestHandler) Close() error {
	if e.netConn.Conn == nil {
		return fmt.Errorf("Not Accepted")
	}

	switch e.options.Protocol {
	case "tcp":
		if err := e.netConn.Close(); err != nil {
			return err
		}
		listenersWaitGroupsMutex.RLock()
		listenersWaitGroups[e.options.Port].Done()
		listenersWaitGroupsMutex.RUnlock()
		return nil

	default:
		return fmt.Errorf("Unknown Protocol")
	}
}

func (e *ServerRequestHandler) Receive() ([]byte, error) {
	if e.netConn.Conn == nil {
		return nil, fmt.Errorf("Not Accepted")
	}

	switch e.options.Protocol {
	case "tcp":
		return e.netConn.ReadData()

	default:
		return nil, fmt.Errorf("Unknown Protocol")
	}
}

func (e *ServerRequestHandler) Send(message []byte) error {
	if e.netConn.Conn == nil {
		return fmt.Errorf("Not Accepted")
	}

	switch e.options.Protocol {
	case "tcp":
		_, err := e.netConn.WriteData(message)
		return err

	default:
		return fmt.Errorf("Unknown Protocol")
	}
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
