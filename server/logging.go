package server

import (
	"log"
	"os"
)

type LoggingLevel int

const (
	LogDisabled LoggingLevel = 00
	LogEnabled  LoggingLevel = ^0
)

var (
	logger       = log.New(os.Stderr, "[server] ", log.LstdFlags)
	loggingLevel = LogDisabled
)

func SetLoggingLevel(ll LoggingLevel) {
	loggingLevel = ll
}
