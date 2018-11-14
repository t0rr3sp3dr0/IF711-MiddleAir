package util

import (
	"reflect"

	model "../proto"
	"github.com/golang/protobuf/proto"
)

func SelfDescribingMessage(message proto.Message) ([]byte, error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}

	e := &model.SelfDescribingMessage{
		TypeName:    reflect.TypeOf(message).String(),
		MessageData: data,
	}

	bytes, err := proto.Marshal(e)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
