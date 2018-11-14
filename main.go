package main

import (
	_ "./bonjour"
)

func main() {
	<-make(chan struct{})
}
