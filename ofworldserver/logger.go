package ofworldserver

import (
	"github.com/nw/ofworldserver/go-logging"
	"log"
	"os"
	"path/filepath"
)

var (
	Logger  *logging.Logger = nil
	logPath string          = ""
)

func init() {
	//var curdir string = ""
	curdir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("get log save path failed", err)
		os.Exit(1)
	}

	curdir += "/log"
	if _, err = os.Stat(curdir); os.IsNotExist(err) {
		err = os.Mkdir(curdir, 0777)
		if err != nil {
			log.Fatal("make log dir failed ", err)
			os.Exit(1)
		}
	}

	logPath = curdir + "/"
}

func CreateLogger(fileName string) {
	logger := logging.MustGetLogger("logger")
	//logging.SetLevel(logging.WARNING, "logger")
	//
	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfile} %{longfunc} >>> %{level:.4s} %{id:04d} %{message}%{color:reset}`,
	)

	logPath += fileName
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		log.Fatal("open log file error", err)
	}

	backend1 := logging.NewLogBackend(logFile, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)

	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	logging.SetBackend(backend1Formatter, backend2Formatter)
	//logging.SetLevel(logging.INFO, "logger")

	Logger = logger

}
