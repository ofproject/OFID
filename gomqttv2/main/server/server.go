package main

import (
	mq "github.com/nw/gomqttv2"
	"net"
	//"time"
	//"regexp"
)

/*
	1. 解析 .conf
		定义一个结构，记录.conf里的信息
		1. server or client，
		2. 是否开启Dump
		3. 设置log文件目录
		4. 如果是server 需要监听的端口
		5. 如果是client 需要连接的服务器地址
		6. redis 的连接地址
	2. flag

*/

func main() {

	/*var addr string = "0x009c48936bef28dc2a9d38a3cb87d869a56869bab3cc"
	ok, err := regexp.MatchString("^0x[0-9a-f]*$", addr)
	if err != nil {
		return
	}

	fmt.Println("ok : ", ok)

	c := make(chan int)
	fmt.Println(time.Now().String())

	go func() {
		time.Sleep(3 * time.Second)
		c <- 0
	}()

	select {
	case _ = <-c:
		fmt.Println("data")
	case <-time.After(5 * time.Second):
		fmt.Println("timed out")
	}

	fmt.Println(time.Now().String())*/

	mq.CreateLogger("service.log")
	//mq.LogName = "service.log"
	mq.RedisAddr = "127.0.0.1:6379"
	mq.NewDataBase()
	mq.Logger.Warning("Server start")
	mq.InitManagers()

	l, err := net.Listen("tcp", ":64934")
	if err != nil {
		mq.Logger.Fatal(err)
	}

	server := mq.NewService(l)
	server.Dump = false
	server.Start()

	//<-server.Done
	server.Wait()
	server.Close()
	mq.Logger.Info("Server stop")
}
