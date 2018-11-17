package server

import (
	"reflect"

	"github.com/golang/protobuf/proto"
)

type HandleFn func(proto.Message) (proto.Message, error)

type Service struct {
	Interface reflect.Type
	Handle    HandleFn
}

type ServerProxy interface {
	Registry() []*Service
	Tags() [12]string
}
