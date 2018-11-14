package bonjour

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	model "../proto"
)

const (
	ipv4Host     = "224.0.0.57"
	ipv4Port     = 13374
	ipv6Host     = "ff01::39"
	ipv6Port     = 13376
	datagramSize = 8192
	timeout      = 2 * time.Second
)

var (
	addresses = []*net.UDPAddr{
		// IPv4
		func() *net.UDPAddr {
			addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", ipv4Host, ipv4Port))
			if err != nil {
				panic(err)
			}
			return addr
		}(),

		// IPv6
		func() *net.UDPAddr {
			addr, err := net.ResolveUDPAddr("udp6", fmt.Sprintf("[%s]:%d", ipv6Host, ipv6Port))
			if err != nil {
				panic(err)
			}
			return addr
		}(),
	}[:1]
)

func init() {
	go func() {
		for {
			log.Println(ripperLoop())
		}
	}()
	go func() {
		for {
			log.Println(listenerLoop())
		}
	}()
	go func() {
		for {
			log.Println(broadcasterLoop())
		}
	}()

	s := &Service{
		Type: func() string {
			hostname, _ := os.Hostname()
			return fmt.Sprintf("helloWorld@%s", hostname)
		}(),
		Host: 0,
		Port: 0,
	}
	RegisterCallback(func(addr net.Addr, announcement *model.Announcement) {
		log.Println(addr, announcement)

		if announcement.Type == s.Type {
			UnregisterService(s)
		}
	})
	RegisterService(s)
}
