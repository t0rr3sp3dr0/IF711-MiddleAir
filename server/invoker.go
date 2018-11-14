package server

import (
	"io"
	"log"

	"../bonjour"
	model "../proto"
	"../util"
)

type Invoker struct {
	mashaler *util.Mashaler
	srh      *ServerRequestHandler
}

func NewInvoker(options util.Options) (*Invoker, error) {
	mashaler, err := util.NewMashaler()
	if err != nil {
		return nil, err
	}

	srh, err := NewServerRequestHandler(options)
	if err != nil {
		return nil, err
	}

	return &Invoker{
		mashaler: mashaler,
		srh:      srh,
	}, nil
}

func (e *Invoker) Loop(sp ServerProxy) error {
	for _, uuid := range sp.Registry() {
		s := &bonjour.Service{
			UUID: uuid,
			Provider: bonjour.Provider{
				Port: e.srh.options.Port,
			},
		}
		bonjour.RegisterService(s)
		defer bonjour.UnregisterService(s)
	}

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

		innerMessage, handler, err := sp.Demux(message.TypeName)
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
