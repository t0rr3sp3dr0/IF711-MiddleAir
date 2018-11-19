package main

import (
	"reflect"

	proto "github.com/golang/protobuf/proto"
	"github.com/t0rr3sp3dr0/middleair/server"
)

type BenchmarkServer struct{}

func (e *BenchmarkServer) Registry() []*server.Service {
	services := []*server.Service{
		&server.Service{
			Interface: reflect.TypeOf((*Request)(nil)),
			Handle: func(message proto.Message) (proto.Message, error) {
				return &Response{}, nil
			},
		},
	}

	return services
}

func (e *BenchmarkServer) Tags() (tags [12]string) {
	return tags
}
