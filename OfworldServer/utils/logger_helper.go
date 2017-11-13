package utils

import (
	"os"
	"fmt"
	"github.com/op/go-logging"
)

func CreateLogger() *logging.Logger {

	log := logging.MustGetLogger("logger")

	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} > %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)

	logFile, err := os.OpenFile("/usr/local/logger.txt", os.O_WRONLY, 0777)
	if err != nil {
		fmt.Println(err)
	}

	backend1 := logging.NewLogBackend(logFile, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)

	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	logging.SetBackend(backend1Formatter, backend2Formatter)

	return log
}
