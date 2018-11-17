package main

import (
	"./bonjour"
)

func main() {
	bonjour.SetLoggingLevel(bonjour.LogEveryone)
	<-make(chan struct{})
}
