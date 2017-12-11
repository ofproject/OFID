package ofworldnodes

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"
	//"time"
)

var (
	id uint64 = 2
)

func getID() string {
	sid := strconv.FormatUint(id, 10)
	id++
	return sid
}

func BenchmarkClient(B *testing.B) {
	logName := fmt.Sprintf("%s.log", getID())
	CreateLogger(logName)
	for i := 0; i < B.N; i++ {
		clientId := "Clientid=" + getID()
		//logName := fmt.Sprintf("%s.log", getID())

		Logger.Info("Client start")

		conn, err := net.Dial("tcp", "127.0.0.1:64934")
		if err != nil {
			Logger.Error("connect server failed: ", err)
			os.Exit(1)
		}

		client := NewClient(conn, clientId)
		if client != nil {
			client.Dump = false
			//client.Connect("niwei", "123456")

			client.Disconnect()
			/*select {
			case <-time.After(1 * time.Millisecond):
				client.Disconnect()
			}*/

			client.Wait()
		}
	}
}
