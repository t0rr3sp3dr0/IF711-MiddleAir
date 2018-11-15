package server

import (
	"io"
	"log"

	"../bonjour"
	model "../proto"
	"../util"
)

type Invoker struct {
	sp       ServerProxy
	services []*bonjour.Service
	mashaler *util.Mashaler
	srh      *ServerRequestHandler
}

func NewInvoker(sp ServerProxy, options util.Options) (*Invoker, error) {
	mashaler, err := util.NewMashaler()
	if err != nil {
		return nil, err
	}

	srh, err := NewServerRequestHandler(options)
	if err != nil {
		return nil, err
	}

	registry := sp.Registry()
	services := make([]*bonjour.Service, 0, len(registry))
	for _, uuid := range registry {
		s := &bonjour.Service{
			UUID: uuid,
			Provider: bonjour.Provider{
				Port: options.Port,
			},
		}
		bonjour.RegisterService(s)
		services = append(services, s)
	}

	return &Invoker{
		sp:       sp,
		services: services,
		mashaler: mashaler,
		srh:      srh,
	}, nil
}

func (e *Invoker) Accept() error {
	return e.srh.Accept()
}

func (e *Invoker) Loop() error {
	defer func() {
		for _, service := range e.services {
			bonjour.UnregisterService(service)
		}
		if err := e.srh.Close(); err != nil {
			log.Println(err)
		}

		e.sp = nil
		e.services = nil
		e.mashaler = nil
		e.srh = nil
	}()

	for {
		bytes, err := e.srh.Receive()
		if err != nil {
			if err == io.EOF {
				break
			}
			e.srh.handleBadRequest(err)
			continue
		}

		message := &model.SelfDescribingMessage{}
		if err := e.mashaler.Unmarshal(bytes, message); err != nil {
			e.srh.handleBadRequest(err)
			continue
		}

		innerMessage, handler, err := e.sp.Demux(message.TypeName)
		if err != nil {
			e.srh.handleBadRequest(err)
			continue
		}

		if err := e.mashaler.Unmarshal(message.MessageData, innerMessage); err != nil {
			e.srh.handleBadRequest(err)
			continue
		}
		log.Println(message, "\n", innerMessage)

		response, err := handler(innerMessage)
		if err != nil {
			e.srh.handleInternalServerError(err)
			continue
		}

		var res []byte
		switch response.(type) {
		case *model.ErrorResponse:
			data, err := e.mashaler.Marshal(response)
			if err != nil {
				e.srh.handleInternalServerError(err)
			}
			res = data

		default:
			data, err := util.SelfDescribingMessage(response)
			if err != nil {
				e.srh.handleInternalServerError(err)
				continue
			}
			res = data
		}

		if err := e.srh.Send(res); err != nil {
			return err
		}
	}
	return nil
}
