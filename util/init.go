package util

import (
	"errors"
	"reflect"

	model "../proto"
	"github.com/golang/protobuf/proto"
)

var (
	ErrUnknown            = errors.New("000 - Unknown")
	ErrUnauthorized       = errors.New("401 - Unauthorized")
	ErrForbidden          = errors.New("403 - Forbidden")
	ErrNotFound           = errors.New("404 - Not Found")
	ErrExpectationFailed  = errors.New("417 - Expectation Failed")
	ErrServiceUnavailable = errors.New("503 - Service Unavailable")
)

type Options struct {
	Host        string
	Port        uint16
	Protocol    string
	Credentials []byte
}

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
