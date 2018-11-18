package bonjour

import (
	"fmt"
	"net"
	"strings"
	"time"

	model "github.com/t0rr3sp3dr0/middleair/proto"
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
			logger.Println(ripperLoop())
		}
	}()
	go func() {
		for {
			logger.Println(listenerLoop())
		}
	}()
	go func() {
		for {
			logger.Println(broadcasterLoop())
		}
	}()

	s := &Service{
		UUID: "HELO",
	}
	c := func(addr net.Addr, announcement *model.ServiceAnnouncement) {
		host := strings.Split(addr.String(), ":")[0]
		b := host == s.Provider.Host
		if (b && loggingLevel&LogLocalhost != LogDisabled) || (!b && loggingLevel&LogOthers != LogDisabled) {
			logger.Println(addr, announcement)
		}

		if announcement.Uuid == s.UUID {
			s.Provider.Host = host
			UnregisterService(s)
		}
	}
	RegisterCallback(c)
	RegisterService(s)
}
