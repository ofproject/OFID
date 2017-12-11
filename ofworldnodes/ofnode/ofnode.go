package main

import (
	"flag"
	"fmt"
	mq "github.com/nw/ofworldnodes"
	"net"
	"os"
	"time"
	//"strconv"
)

var (
	clientid   string
	serverip   string
	serverport string
	uname      string
	upass      string = "no"
)

func init() {
	flag.StringVar(&clientid, "id", "", "setting client id")
	flag.StringVar(&serverip, "ip", "127.0.0.1", "setting server ip")
	flag.StringVar(&serverport, "port", "64934", "setting server port")
	flag.StringVar(&uname, "name", "", "setting client user name")
}

func main() {
	flag.Parse()

	if len(clientid) == 0 || len(uname) == 0 {
		fmt.Println("clientid or name is empty")
		return
	}

	//clientId := "ABCDEFGHIGKLMN"
	//clientIdto := "ABC1"
	logName := fmt.Sprintf("node-%s.log", clientid)
	mq.CreateLogger(logName)
	mq.Logger.Info("Client start")

	serverAddr := fmt.Sprintf("%s:%s", serverip, serverport)

	conn, err := net.Dial("tcp", serverAddr)
	//conn, err := net.Dial("tcp", "127.0.0.1:64934")
	if err != nil {
		mq.Logger.Error("connect server failed: ", err)
		os.Exit(1)
	}

	rpc := new(mq.Rpcapi)

	client := mq.NewClient(conn, clientid)

	if client != nil {
		go rpc.StartForTCPCodecJson(":64935", client)
		time.Sleep(2)
		client.Dump = false
		client.Connect(uname, upass)

		client.Wait()
		rpc.Close()
	}
}
