package client

import (
	"fmt"

	model "../proto"
	"../util"
	"github.com/golang/protobuf/proto"
)

type Requestor struct {
	mashaler *util.Mashaler
	crh      *ClientRequestHandler
}

func NewRequestor(options util.Options) (*Requestor, error) {
	mashaler, err := util.NewMashaler()
	if err != nil {
		return nil, err
	}

	crh, err := NewClientRequestHandler(options)
	if err != nil {
		return nil, err
	}

	return &Requestor{
		mashaler: mashaler,
		crh:      crh,
	}, nil
}

func (e *Requestor) Invoke(req proto.Message, res proto.Message) error {
	data, err := util.SelfDescribingMessage(req)
	if err != nil {
		return err
	}

	err = e.crh.Send(data)
	if err != nil {
		return err
	}

	response, err := e.crh.Receive()
	if err != nil {
		return err
	}

	selfDescribingMessage := &model.SelfDescribingMessage{}
	if err := e.mashaler.Unmarshal(response, selfDescribingMessage); err != nil {
		return err
	}

	if selfDescribingMessage.Error != nil {
		return fmt.Errorf("%d: %s", selfDescribingMessage.Error.Code, selfDescribingMessage.Error.Message)
	}

	if err := e.mashaler.Unmarshal(selfDescribingMessage.MessageData, res); err != nil {
		return err
	}

	return nil
}
