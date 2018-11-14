package server

import (
	"github.com/golang/protobuf/proto"
)

type ServerProxy interface {
	Registry() []string
	Demux(string) (proto.Message, func(proto.Message) (proto.Message, error), error)
}
