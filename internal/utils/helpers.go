package utils

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	Debug = false
)

func init() {
	if os.Getenv("DEBUG") == "1" {
		Debug = true
	}
}

func WaitForSigInt() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	<-c
}

func LeftPad(value string, length int, padChar string) string {
	if len(value) > length {
		return value
	}

	output := ""

	for i := 0; i < length-len(value); i++ {
		output += padChar
	}

	return output + value
}

func DebugLog(a, b, message string) {
	if !Debug {
		return
	}

	log.Printf(
		"%v\t%v\t%v",
		LeftPad(a, 14, "_"),
		LeftPad(b, 15, "_"),
		message,
	)
}
