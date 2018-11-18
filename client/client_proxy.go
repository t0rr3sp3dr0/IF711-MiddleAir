package client

import (
	"log"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/t0rr3sp3dr0/middleair/bonjour"
	"github.com/t0rr3sp3dr0/middleair/util"
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
	Credentials []byte
}

func Invoke(req proto.Message, res proto.Message, options *Options) error {
	if options == nil {
		options = &Options{}
	}
	if options.Tags == nil {
		options.Tags = []string{}
	}
	if options.Credentials == nil {
		options.Credentials = []byte{}
	}

	instances := bonjour.InstancesOfService(reflect.TypeOf(req).String())
	if len(instances) == 0 {
		return util.ErrNotFound
	}

	b := false
	for _, instance := range instances {
		if len(options.Tags) > 0 {
			matches := 0
		loop:
			for _, localTag := range options.Tags {
				for _, remoteTag := range append([]string{
					instance.Metadata.OS,
					instance.Metadata.Arch,
					instance.Metadata.Host,
					instance.Metadata.Lang,
				}, instance.Tags[:]...) {
					if remoteTag == localTag {
						matches++
						if !options.StrictMatch {
							break loop
						}
						continue loop
					}
				}
			}
			if matches == 0 || (options.StrictMatch && matches < len(options.Tags)) {
				continue
			}
		}

		proxy, ok := func(provider *bonjour.Provider) (*ClientProxy, bool) {
			if !options.Persistent {
				return nil, false
			}

			proxy, ok := proxies[*provider]
			return proxy, ok
		}(&instance.Provider)
		if !ok {
			clientProxy, err := NewClientProxy(util.Options{
				Host:        instance.Provider.Host,
				Port:        instance.Provider.Port,
				Protocol:    "tcp",
				Credentials: options.Credentials,
			})
			if err != nil {
				log.Println(err)
				continue
			}

			if options.Persistent {
				proxies[instance.Provider] = clientProxy
			}

			proxy = clientProxy
		}
		if !options.Persistent {
			defer proxy.Close()
		}

		if err := proxy.Invoke(req, res); err != nil {
			log.Println(err)
			continue
		}

		if !options.Broadcast {
			return nil
		}
		b = true
	}

	if !b {
		return util.ErrServiceUnavailable
	}
	return nil
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
