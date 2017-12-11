package main

import (
	"bytes"
	"fmt"
	mq "github.com/nw/gomqttv2"
	"net"
	"os"
)

func main() {
	clientId := "ABCDEFGHIGKLMN"
	clientIdto := "ABC1"
	logName := fmt.Sprintf("client-%s.log", clientId)
	mq.CreateLogger(logName)
	//mq.LogName = "service.log"
	mq.Logger.Info("Client start")

	conn, err := net.Dial("tcp", "127.0.0.1:64934")
	if err != nil {
		mq.Logger.Error("connect server failed: ", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	buf.WriteString("kfsdajklfjaslkjflka;sd")

	client := mq.NewClient(conn, clientIdto)
	if client != nil {
		client.Dump = false
		client.Connect("niwei", "123456")
		//client.Publish(clientIdto, buf.Bytes())
		go client.RecvPuback()

		//buf.WriteString("sfdasfasdfasfasdfas")

		/*var sub []string
		sub = append(sub, "fdsafasdfasd")
		sub = append(sub, "fdasfasfdsafdsaf")
		client.Subscribe(sub)*/

		//client.Disconnect()

		client.Wait()
	}
}
