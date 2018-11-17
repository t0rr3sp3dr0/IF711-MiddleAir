package bonjour

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"time"

	model "../proto"
	"../util"
	"github.com/golang/protobuf/proto"
)

var (
	registeredServices      = make(map[*Service]*struct{})
	registeredServicesMutex = &sync.RWMutex{}
)

func broadcasterLoop() (ret error) {
	defer func() {
		if r := recover(); r != nil {
			ret = fmt.Errorf("%v", r)
		}
	}()

	var connections []*net.UDPConn
	for _, addr := range addresses {
		conn, err := net.DialUDP(addr.Network(), nil, addr)
		if err != nil {
			return err
		}
		defer conn.Close()

		connections = append(connections, conn)
	}

	for {
		var once sync.Once
		registeredServicesMutex.RLock()
		defer once.Do(registeredServicesMutex.RUnlock)
		for service := range registeredServices {
			announcement := &model.ServiceAnnouncement{
				Uuid: service.UUID,
				Port: int32(service.Provider.Port),
				Tags: append(service.Tags[:], service.Metadata.OS, service.Metadata.Arch, service.Metadata.Host, service.Metadata.Lang),
			}

			message, err := proto.Marshal(announcement)
			if err != nil {
				return err
			}

			for _, conn := range connections {
				if _, err := conn.Write(message); err != nil {
					return err
				}
			}
		}
		once.Do(registeredServicesMutex.RUnlock)

		time.Sleep(500 * time.Millisecond)
	}
}

func RegisterService(service *Service) error {
	if len(service.UUID) > 256 {
		return util.ErrPayloadTooLarge
	}
	if len(service.Provider.Host) > 256 {
		return util.ErrPayloadTooLarge
	}
	for _, tag := range service.Tags {
		if len(tag) > 256 {
			return util.ErrPayloadTooLarge
		}
	}

	service.Metadata.OS = runtime.GOOS
	service.Metadata.Arch = runtime.GOARCH
	if hostname, err := os.Hostname(); err == nil {
		service.Metadata.Host = hostname
		if len(service.Metadata.Host) > 256 {
			service.Metadata.Host = service.Metadata.Host[:256]
		}
	}
	if language, ok := os.LookupEnv("LANG"); ok {
		service.Metadata.Lang = language
		if len(service.Metadata.Lang) > 256 {
			service.Metadata.Lang = service.Metadata.Lang[:256]
		}
	}

	registeredServicesMutex.Lock()
	defer registeredServicesMutex.Unlock()

	registeredServices[service] = nil
	return nil
}

func UnregisterService(service *Service) {
	registeredServicesMutex.Lock()
	defer registeredServicesMutex.Unlock()

	delete(registeredServices, service)
}
