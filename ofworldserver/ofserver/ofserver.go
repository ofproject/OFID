package main

import (
	"flag"
	mq "github.com/nw/ofworldserver"
	"log"
	"net"
	"net/http"
	"os"
	//"runtime/pprof"
	_ "net/http/pprof"
	"os/signal"
	"syscall"
)

var (
	cpuprofile = ""
)

func init() {
	flag.StringVar(&cpuprofile, "cpu", "./cpu.prof", "usage")
}

func main() {
	flag.Parse()
	/*if cpuprofile == "" {
		return
	}

	f, err := os.Create(cpuprofile)
	if err != nil {
		return
	}

	pprof.StartCPUProfile(f)*/

	mq.CreateLogger("service.log")
	mq.NewConsistent()
	mq.Logger.Warning("Server start")

	//l, err := net.Listen("tcp", ":64934")
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":64934")
	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		mq.Logger.Fatal(err)
	}

	server := mq.NewService(*l)
	s := mq.NewHttpService()
	go func() {
		s.Start()
	}()
	server.Dump = false
	server.Start()

	//<-server.Done
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	//signal start
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, _ = <-sigs
		//pprof.StopCPUProfile()
		//f.Close()
		server.Close()
	}()

	//signal edn

	server.Wait()
	//server.Close()
	s.Close()
	mq.Logger.Info("Server stop")
}
