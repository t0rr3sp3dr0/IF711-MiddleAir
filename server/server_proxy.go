package server

import (
	"reflect"

	"github.com/golang/protobuf/proto"
)

type HandleFn func(proto.Message) (proto.Message, error)

type Service struct {
	UUID   string
	Tags   []string
	Handle HandleFn
	InType reflect.Type
}

type ServerProxy interface {
	Registry() []*Service
}
