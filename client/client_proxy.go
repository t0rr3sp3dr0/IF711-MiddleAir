package client

import (
	"errors"
	"log"

	"../bonjour"
	"../util"
	"github.com/golang/protobuf/proto"
)

var (
	proxies               = make(map[bonjour.Provider]*ClientProxy)
	ErrNotFound           = errors.New("404 - Not Found")
	ErrExpectationFailed  = errors.New("417 - Expectation Failed")
	ErrServiceUnavailable = errors.New("503 - Service Unavailable")
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

func (e *ClientProxy) Close() error {
	return e.requestor.Close()
}

func (e *ClientProxy) Invoke(req proto.Message, res proto.Message) error {
	return e.requestor.Invoke(req, res)
}

type Options struct {
	Tags        []string
	StrictMatch bool
	Broadcast   bool
	Persistent  bool
}

type InvokeFn func(req proto.Message, res proto.Message) error

func GetServiceInvokeFn(uuid string, options *Options) (InvokeFn, error) {
	if options == nil {
		options = &Options{}
	}
	if options.Tags == nil {
		options.Tags = []string{}
	}

	providers := bonjour.GetProvidersForService(uuid)
	if len(providers) == 0 {
		return nil, ErrNotFound
	}

	return func(req proto.Message, res proto.Message) error {
		b := false
		for _, provider := range providers {
			// TODO: tag matching

			proxy, ok := func(provider *bonjour.Provider) (*ClientProxy, bool) {
				if !options.Persistent {
					return nil, false
				}

				proxy, ok := proxies[*provider]
				return proxy, ok
			}(&provider)
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

				if options.Persistent {
					proxies[provider] = clientProxy
				}

				proxy = clientProxy
			}

			if err := proxy.Invoke(req, res); err != nil {
				log.Println(err)
				continue
			}
			if !options.Persistent {
				proxy.requestor.crh.Close()
			}

			if !options.Broadcast {
				return nil
			}
			b = true
		}

		if !b {
			return ErrServiceUnavailable
		}
		return nil
	}, nil
}

func ClosePersistentConns() (errs []error) {
	for provider, proxy := range proxies {
		if err := proxy.Close(); err != nil {
			errs = append(errs, err)
			continue
		}

		delete(proxies, provider)
	}

	return errs
}
