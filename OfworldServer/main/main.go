package main

import (
	"net"
	"OfworldServer/server"
	"fmt"
	"OfworldServer/httpapi"
	"github.com/op/go-logging"

	"OfworldServer/utils"
)

func main() {
	fmt.Println("server start")

	startHTTP(":8080")
}
func startHTTP(endpoint string) error {
	// Short circuit if the HTTP endpoint isn't being exposed
	if endpoint == "" {
		return nil
	}
	// Register all the APIs exposed by the services
	handler := rpc.NewServer()

	logger :=utils.CreateLogger()

	for _, object := range apis(logger) {
		handler.RegisterName(object.name, object.object)
	}
	// All APIs registered, start the HTTP listener
	var (
		listener net.Listener
		err      error
	)
	if listener, err = net.Listen("tcp", endpoint); err != nil {
		fmt.Println(err)
		return err
	}

	go rpc.NewHTTPServer(nil, handler).Serve(listener)

	logger.Info(fmt.Sprintf("HTTP endpoint opened: http://%s", endpoint))

	test := make(chan struct{})
	for {
		<-test
	}
	return nil

}

type API struct {
	name   string
	object interface{}
}

func apis(logger *logging.Logger) []API {

	return []API{
		{
			name:   "ofworld",
			object: httpapi.NewOfWorldAPI(logger),
		},
	}

}

