package bonjour

import (
	"log"
	"os"
)

type LoggingLevel int

const (
	LogDisabled  LoggingLevel = 0x00
	LogLocalhost LoggingLevel = 0x01
	LogOthers    LoggingLevel = 0x10
	LogEveryone  LoggingLevel = ^0x0
)

var (
	logger       = log.New(os.Stderr, "[bonjour] ", log.LstdFlags)
	loggingLevel = LogDisabled
)

func SetLoggingLevel(ll LoggingLevel) {
	loggingLevel = ll
}
