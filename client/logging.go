package client

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
	logger       = log.New(os.Stderr, "[client] ", log.LstdFlags)
	loggingLevel = LogDisabled
)

func SetLoggingLevel(ll LoggingLevel) {
	loggingLevel = ll
}
