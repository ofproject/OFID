package ofworldserver

import (
	"fmt"
	"math/rand"
	"net"
	"sync/atomic"
	"time"
)

type stats struct {
	recv       int64
	sent       int64
	clients    int64
	clientsMax int64
	lastmsgs   int64
}

func (s *stats) messageRecv()      { atomic.AddInt64(&s.recv, 1) }
func (s *stats) messageSend()      { atomic.AddInt64(&s.sent, 1) }
func (s *stats) clientConnect()    { atomic.AddInt64(&s.clients, 1) }
func (s *stats) clientDisconnect() { atomic.AddInt64(&s.clients, -1) }

func (s *stats) PrintStatus(service *Service) {

	for {
		select {
		case <-service.Done:
			return
		case <-time.After(60000 * time.Second):
			recvStr := fmt.Sprintf("recv count: %d", s.recv)
			sendStr := fmt.Sprintf("send count: %d", s.sent)
			clientStr := fmt.Sprintf("Client Max: %d, Client count: %d", s.clientsMax, s.clients)
			msgStr := fmt.Sprintf("LastMessage: %d", s.lastmsgs)
			Logger.Noticef("STATUS:[%s][%s][%s][%s]\n", recvStr, sendStr, clientStr, msgStr)
		}
	}
}

//test

func (s *stats) PrintStatus2() {
	recvStr := fmt.Sprintf("recv count: %d", s.recv)
	sendStr := fmt.Sprintf("send count: %d", s.sent)
	clientStr := fmt.Sprintf("Client Max: %d, Client count: %d", s.clientsMax, s.clients)
	msgStr := fmt.Sprintf("LastMessage: %d", s.lastmsgs)
	Logger.Noticef("STATUS:[%s][%s][%s][%s]\n", recvStr, sendStr, clientStr, msgStr)
}

var (
	globID uint32 = 1
)

func nextID() uint32 {
	if globID == 0xFFFFFFFF {
		globID = 1
	}

	id := globID
	globID++
	return id
}

type Service struct {
	//listener      net.Listener
	listener      net.TCPListener
	stats         *stats
	Done          chan struct{}
	statsInterval time.Duration
	Dump          bool
	subs          *rand.Rand
}

func NewService(l net.TCPListener) *Service {
	svr := &Service{
		listener:      l,
		stats:         &stats{},
		Done:          make(chan struct{}),
		statsInterval: time.Second * 10,
	}

	go func() {
		for {
			select {
			case <-svr.Done:
				return
			case <-time.After(svr.statsInterval):
			}

			//time.Sleep(svr.statsInterval)

		}
	}()

	go func() {
		svr.stats.PrintStatus(svr)
	}()

	return svr
}

func (s *Service) GetHandler(conn net.Conn) *Node {
	return &Node{
		s:         s,
		conn:      conn,
		Done:      make(chan struct{}),
		writeChan: make(chan sendPackage, sendingQueueLength),
	}
}

func (s *Service) Wait() {
	<-s.Done
}

func (s *Service) Close() {
	s.listener.Close()
	//close(s.Done)
}

func (s *Service) Start() {
	go func() {
		Logger.Notice("Start Listener ", s.listener.Addr().String())
		for {
			conn, err := s.listener.AcceptTCP()
			if err != nil {
				Logger.Error("Accept:", err)
				break
			}

			Logger.Debug("Client coming in")
			conn.SetNoDelay(true)
			node := s.GetHandler(conn)
			//Logger.Debug("clinet ip2: ", node.conn.RemoteAddr().String())
			//s.stats.PrintStatus2()
			//s.stats.clientConnect()
			node.start()
		}

		close(s.Done)
	}()
}
