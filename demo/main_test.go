package main

import (
	"log"
	"testing"
	"time"

	"github.com/t0rr3sp3dr0/middleair/client"
	"github.com/t0rr3sp3dr0/middleair/server"
	"github.com/t0rr3sp3dr0/middleair/util"
)

func init() {
	go func() {
		opt := util.Options{
			Port:     1337,
			Protocol: "tcp",
		}

		for {
			invoker, err := server.NewInvoker(&BenchmarkServer{}, opt)
			if err != nil {
				log.Println(err)
				continue
			}

			if err := invoker.Accept(nil); err != nil {
				log.Println(err)
				continue
			}

			go func() {
				if err := invoker.Loop(); err != nil {
					panic(err)
				}
			}()
		}
	}()

	time.Sleep(2 * time.Second)
}

func BenchmarkAmnesia(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Invoke(&Request{}, &Response{}, nil)
	}
}

func BenchmarkPersistent(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Invoke(&Request{}, &Response{}, &client.Options{
			Persistent: true,
		})
	}
}
