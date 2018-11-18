package main

import (
	"github.com/t0rr3sp3dr0/middleair/bonjour"
)

func main() {
	bonjour.SetLoggingLevel(bonjour.LogEveryone)
	<-make(chan struct{})
}
