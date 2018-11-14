package bonjour

import (
	"fmt"
	"net"
	"sync"
	"time"
	"unsafe"

	model "../proto"
	"github.com/golang/protobuf/proto"
)

var (
	callbacks           = make(map[unsafe.Pointer]*struct{})
	callbacksMutex      = &sync.RWMutex{}
	remoteServices      = make(map[string]map[*Service]time.Time)
	remoteServicesMutex = &sync.RWMutex{}
)

func ripperLoop() (ret error) {
	defer func() {
		if r := recover(); r != nil {
			ret = fmt.Errorf("%v", r)
		}
	}()

	for {
		var once sync.Once
		remoteServicesMutex.Lock()
		defer once.Do(remoteServicesMutex.Unlock)
		for k, services := range remoteServices {
			for service, timestamp := range services {
				if time.Now().After(timestamp.Add(timeout)) {
					delete(services, service)
				}
			}
			if len(services) == 0 {
				delete(remoteServices, k)
			}
		}
		once.Do(remoteServicesMutex.Unlock)

		time.Sleep(timeout)
	}
}

func listenerLoop() (ret error) {
	defer func() {
		if r := recover(); r != nil {
			ret = fmt.Errorf("%v", r)
		}
	}()

	var connections []*net.UDPConn
	for _, addr := range addresses {
		conn, err := net.ListenMulticastUDP(addr.Network(), nil, addr)
		if err != nil {
			return err
		}
		defer conn.Close()

		if err := conn.SetReadBuffer(datagramSize); err != nil {
			return err
		}

		connections = append(connections, conn)
	}

	ch := make(chan func() ([]byte, *net.UDPAddr, error))
	for _, conn := range connections {
		go func(conn *net.UDPConn) {
			for {
				buffer := make([]byte, datagramSize)
				n, addr, err := conn.ReadFromUDP(buffer)
				ch <- func() ([]byte, *net.UDPAddr, error) {
					return buffer[:n], addr, err
				}
			}
		}(conn)
	}

	for {
		for fn := range ch {
			buffer, addr, err := fn()
			if err != nil {
				return err
			}

			announcement := &model.Announcement{}
			if err := proto.Unmarshal(buffer, announcement); err != nil {
				fmt.Println(err)
				continue
			}

			remoteServicesMutex.RLock()
			_, ok := remoteServices[announcement.Type]
			remoteServicesMutex.RUnlock()
			if !ok {
				remoteServicesMutex.Lock()
				remoteServices[announcement.Type] = make(map[*Service]time.Time)
				remoteServicesMutex.Unlock()
			}

			service := &Service{
				Type: announcement.Type,
				Host: uint64(announcement.Host),
				Port: uint16(announcement.Port),
			}
			remoteServicesMutex.Lock()
			remoteServices[announcement.Type][service] = time.Now()
			remoteServicesMutex.Unlock()

			var once sync.Once
			callbacksMutex.RLock()
			defer once.Do(callbacksMutex.RUnlock)
			for callback := range callbacks {
				(*(*func(net.Addr, *model.Announcement))(callback))(net.Addr(addr), announcement)
			}
			once.Do(callbacksMutex.RUnlock)
		}
	}
}

func RegisterCallback(fn func(net.Addr, *model.Announcement)) {
	callbacksMutex.Lock()
	defer callbacksMutex.Unlock()

	callbacks[unsafe.Pointer(&fn)] = nil
}

func UnregisterCallback(fn func(net.Addr, *model.Announcement)) {
	callbacksMutex.Lock()
	defer callbacksMutex.Unlock()

	delete(callbacks, unsafe.Pointer(&fn))
}
