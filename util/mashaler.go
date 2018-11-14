package util

import (
	"github.com/golang/protobuf/proto"
)

type Mashaler struct {
}

func NewMashaler() (*Mashaler, error) {
	return &Mashaler{}, nil
}

func (e *Mashaler) Marshal(message proto.Message) ([]byte, error) {
	return proto.Marshal(message)
}

func (e *Mashaler) Unmarshal(buf []byte, message proto.Message) error {
	return proto.Unmarshal(buf, message)
}
