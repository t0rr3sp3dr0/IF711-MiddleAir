package client

import (
	"fmt"
	"log"

	"../bonjour"
	"../util"
	"github.com/golang/protobuf/proto"
)

var (
	proxies = make(map[bonjour.Provider]*ClientProxy)
)

type ClientProxy struct {
	requestor *Requestor
}

func NewClientProxy(options util.Options) (*ClientProxy, error) {
	requestor, err := NewRequestor(options)
	if err != nil {
		return nil, err
	}

	return &ClientProxy{
		requestor: requestor,
	}, nil
}

func (e *ClientProxy) Invoke(req proto.Message, res proto.Message) error {
	return e.requestor.Invoke(req, res)
}

func GetServiceInvoker(uuid string) (func(req proto.Message, res proto.Message) error, error) {
	providers := bonjour.GetProvidersForService(uuid)
	if len(providers) == 0 {
		return nil, fmt.Errorf("404 - Not Found")
	}

	return func(req proto.Message, res proto.Message) error {
		for _, provider := range providers {
			proxy, ok := proxies[provider]
			if !ok {
				clientProxy, err := NewClientProxy(util.Options{
					Host:     provider.Host,
					Port:     provider.Port,
					Protocol: "tcp",
				})
				if err != nil {
					log.Println(err)
					continue
				}

				proxies[provider] = clientProxy
				proxy = clientProxy
			}

			if err := proxy.Invoke(req, res); err != nil {
				log.Println(err)
				continue
			}
			return nil
		}

		return fmt.Errorf("503 - Service Unavailable")
	}, nil
}
