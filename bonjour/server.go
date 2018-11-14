package bonjour

import (
	"fmt"
	"net"
	"sync"
	"time"

	model "../proto"
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

func RegisterService(service *Service) {
	registeredServicesMutex.Lock()
	defer registeredServicesMutex.Unlock()

	registeredServices[service] = nil
}

func UnregisterService(service *Service) {
	registeredServicesMutex.Lock()
	defer registeredServicesMutex.Unlock()

	delete(registeredServices, service)
}
