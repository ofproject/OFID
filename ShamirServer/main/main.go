package main

import (
	"net"
	"ShamirServer/server"
	"fmt"
	"ShamirServer/httpapi"
	"os"
	"github.com/op/go-logging"
)

func main() {

	startHTTP(":8080", nil)

}
func startHTTP(endpoint string, cors []string) error {

	// Short circuit if the HTTP endpoint isn't being exposed
	if endpoint == "" {
		return nil
	}
	// Register all the APIs exposed by the services
	handler := rpc.NewServer()

    logger :=createLogger()

	for _, object := range apis(logger) {
		handler.RegisterName(object.name, object.object)
	}
	// All APIs registered, start the HTTP listener
	var (
		listener net.Listener
		err      error
	)
	if listener, err = net.Listen("tcp", endpoint); err != nil {
		logger.Error("failed to listen port: ",err)
		return err
	}
	go rpc.NewHTTPServer(cors, handler).Serve(listener)

	logger.Notice(fmt.Sprintf("HTTP endpoint opened: http://%s", endpoint))

	test := make(chan struct{})
	for {
		<-test
	}

	return nil

}


func createLogger() *logging.Logger{

	log:=logging.MustGetLogger("logger")

	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} > %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)

	logFile, err := os.OpenFile("/usr/local/logger.txt", os.O_WRONLY,0777)
	if err != nil{
		fmt.Println(err)
	}

	backend1 := logging.NewLogBackend(logFile, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)

	backend1Formatter:=logging.NewBackendFormatter(backend1,format)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	logging.SetBackend(backend1Formatter, backend2Formatter)

	return log
}


type API struct {

	name   string
	object interface{}

}


func apis(logger *logging.Logger) []API {

	return []API{
		{
			name:   "shamir",
			object: httpapi.NewShamirAPI(logger),
		},
	}

}

// All listeners booted successfully
