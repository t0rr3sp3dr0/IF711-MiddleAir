package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"reflect"
	"runtime"

	proto "github.com/golang/protobuf/proto"
	"github.com/t0rr3sp3dr0/middleair/server"
)

type Server struct{}

func (e *Server) Registry() []*server.Service {
	services := []*server.Service{
		&server.Service{
			Interface: reflect.TypeOf((*RemoteShellRequest)(nil)),
			Handle:    e.remoteShell,
		},
		&server.Service{
			Interface: reflect.TypeOf((*TextToSpeechRequest)(nil)),
			Handle:    e.textToSpeech,
		},
	}

	return services
}

func (e *Server) Tags() (tags [12]string) {
	return tags
}

func (e *Server) remoteShell(message proto.Message) (proto.Message, error) {
	request := message.(*RemoteShellRequest)
	response := &RemoteShellResponse{}

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	cmd := exec.Command(request.Name, request.Args...)
	cmd.Stdin = bytes.NewBuffer(request.Stdin)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		response.ExitCode = ^0
	}

	response.Stdout = stdout.Bytes()
	response.Stderr = stderr.Bytes()

	return response, nil
}

func (e *Server) textToSpeech(message proto.Message) (proto.Message, error) {
	request := message.(*TextToSpeechRequest)

	var name string
	switch runtime.GOOS {
	case "darwin":
		name = "say"

	case "linux":
		name = "espeak"

	default:
		return nil, fmt.Errorf("501 - Not Implemented")
	}

	cmd := exec.Command(name, request.Message)
	cmd.Stdin = nil
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return &TextToSpeechResponse{}, nil
}
