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

func (e *ClientProxy) Invoke(req proto.Message, res proto.Message) error {
	return e.requestor.Invoke(req, res)
}

func GetServiceInvoker(uuid string) (func(req proto.Message, res proto.Message) error, error) {
	providers := bonjour.GetProvidersForService(uuid)
	if len(providers) == 0 {
		return nil, ErrNotFound
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

		return ErrServiceUnavailable
	}, nil
}

func GetServiceInvokerWithAllTags(uuid string, tags []string) (func(req proto.Message, res proto.Message) error, error) {
	providers := bonjour.GetProvidersForService(uuid)
	if len(providers) == 0 {
		return nil, ErrNotFound
	}

	return func(req proto.Message, res proto.Message) error {
		for _, provider := range providers {
			// TODO: continue if provider has not all tags

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

		return ErrServiceUnavailable
	}, nil
}

func GetServiceInvokerWithAnyTags(uuid string, tags []string) (func(req proto.Message, res proto.Message) error, error) {
	providers := bonjour.GetProvidersForService(uuid)
	if len(providers) == 0 {
		return nil, ErrNotFound
	}

	return func(req proto.Message, res proto.Message) error {
		for _, provider := range providers {
			// TODO: continue if provider has not any tags

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

		return ErrServiceUnavailable
	}, nil
}

func GetServiceBroadcaster(uuid string) (func(req proto.Message, res proto.Message) error, error) {
	providers := bonjour.GetProvidersForService(uuid)
	if len(providers) == 0 {
		return nil, ErrNotFound
	}

	return func(req proto.Message, res proto.Message) error {
		b := false
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
			b = true
		}

		if !b {
			return ErrServiceUnavailable
		}
		return nil
	}, nil
}

func GetServiceBroadcasterWithAllTags(uuid string, tags []string) (func(req proto.Message, res proto.Message) error, error) {
	providers := bonjour.GetProvidersForService(uuid)
	if len(providers) == 0 {
		return nil, ErrNotFound
	}

	return func(req proto.Message, res proto.Message) error {
		b := false
		for _, provider := range providers {
			// TODO: continue if provider has not all tags

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
			b = true
		}

		if !b {
			return ErrServiceUnavailable
		}
		return nil
	}, nil
}

func GetServiceBroadcasterWithAnyTags(uuid string, tags []string) (func(req proto.Message, res proto.Message) error, error) {
	providers := bonjour.GetProvidersForService(uuid)
	if len(providers) == 0 {
		return nil, ErrNotFound
	}

	return func(req proto.Message, res proto.Message) error {
		b := false
		for _, provider := range providers {
			// TODO: continue if provider has not any tags

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
			b = true
		}

		if !b {
			return ErrServiceUnavailable
		}
		return nil
	}, nil
}
